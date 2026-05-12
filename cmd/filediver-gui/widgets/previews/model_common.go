package previews

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
	"image"
	"io"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

//go:embed shaders/*
var modelPreviewShaderCode embed.FS

// stingray coords to OpenGL coords
var stingrayToGLCoords = mgl32.Mat4FromRows(
	mgl32.Vec4{1, 0, 0, 0},
	mgl32.Vec4{0, 0, 1, 0},
	mgl32.Vec4{0, -1, 0, 0},
	mgl32.Vec4{0, 0, 0, 1},
)

type modelUniform struct {
	name      string
	blockIdx  int32
	gltype    int32
	location  int32
	offset    int32
	arraySize int32
	bufSlice  []byte
}

func (m modelUniform) size() int32 {
	arrayMult := int32(1)
	if m.arraySize > 0 {
		arrayMult = m.arraySize
	}
	switch m.gltype {
	case gl.INT, gl.UNSIGNED_INT, gl.FLOAT, gl.BOOL:
		return arrayMult * 4
	case gl.FLOAT_VEC2, gl.INT_VEC2, gl.UNSIGNED_INT_VEC2, gl.BOOL_VEC2, gl.DOUBLE:
		return arrayMult * 8
	case gl.FLOAT_VEC3, gl.INT_VEC3, gl.UNSIGNED_INT_VEC3, gl.BOOL_VEC3:
		return arrayMult * 12
	case gl.FLOAT_VEC4, gl.INT_VEC4, gl.UNSIGNED_INT_VEC4, gl.BOOL_VEC4, gl.DOUBLE_VEC2:
		return arrayMult * 16
	case gl.FLOAT_MAT2, gl.FLOAT_MAT3x2, gl.FLOAT_MAT4x2, gl.DOUBLE_MAT2:
		return arrayMult * 32
	case gl.FLOAT_MAT2x3, gl.FLOAT_MAT4x3, gl.FLOAT_MAT3:
		return arrayMult * 48
	case gl.FLOAT_MAT2x4, gl.FLOAT_MAT3x4, gl.FLOAT_MAT4, gl.DOUBLE_MAT2x4, gl.DOUBLE_MAT4x2:
		return arrayMult * 64
	case gl.DOUBLE_MAT3:
		return arrayMult * 72
	case gl.DOUBLE_MAT3x4, gl.DOUBLE_MAT4x3:
		return arrayMult * 96
	case gl.DOUBLE_MAT4:
		return arrayMult * 128
	}
	panic(fmt.Errorf("unhandled case: %v", m.gltype))
}

func (m modelUniform) set(val any) (count int) {
	var err error
	if m.bufSlice != nil {
		count, err = binary.Encode(m.bufSlice, binary.LittleEndian, val)
		if err != nil {
			panic(err)
		}
		return
	} else {
		switch m.gltype {
		case gl.SAMPLER_1D, gl.SAMPLER_2D, gl.SAMPLER_3D, gl.SAMPLER_CUBE, gl.SAMPLER_BUFFER,
			gl.SAMPLER_1D_SHADOW, gl.SAMPLER_2D_SHADOW,
			gl.SAMPLER_1D_ARRAY, gl.SAMPLER_2D_ARRAY,
			gl.SAMPLER_1D_ARRAY_SHADOW, gl.SAMPLER_2D_ARRAY_SHADOW,
			gl.SAMPLER_2D_MULTISAMPLE, gl.SAMPLER_2D_MULTISAMPLE_ARRAY,
			gl.SAMPLER_CUBE_SHADOW, gl.SAMPLER_2D_RECT, gl.SAMPLER_2D_RECT_SHADOW,
			gl.INT_SAMPLER_1D, gl.INT_SAMPLER_2D, gl.INT_SAMPLER_3D, gl.INT_SAMPLER_CUBE, gl.INT_SAMPLER_BUFFER,
			gl.INT_SAMPLER_1D_ARRAY, gl.INT_SAMPLER_2D_ARRAY,
			gl.INT_SAMPLER_2D_MULTISAMPLE, gl.INT_SAMPLER_2D_MULTISAMPLE_ARRAY, gl.INT_SAMPLER_2D_RECT,
			gl.UNSIGNED_INT_SAMPLER_1D, gl.UNSIGNED_INT_SAMPLER_2D, gl.UNSIGNED_INT_SAMPLER_3D, gl.UNSIGNED_INT_SAMPLER_CUBE, gl.UNSIGNED_INT_SAMPLER_BUFFER,
			gl.UNSIGNED_INT_SAMPLER_1D_ARRAY, gl.UNSIGNED_INT_SAMPLER_2D_ARRAY,
			gl.UNSIGNED_INT_SAMPLER_2D_MULTISAMPLE, gl.UNSIGNED_INT_SAMPLER_2D_MULTISAMPLE_ARRAY, gl.UNSIGNED_INT_SAMPLER_2D_RECT:
			valInt, ok := val.(int32)
			if !ok {
				panic(fmt.Errorf("sampler uniform must be set with an int32 texture index"))
			}
			gl.Uniform1i(m.location, valInt)
			return 4
		}
	}
	panic(fmt.Errorf("cannot set non-sampler uniform outside of uniform block!"))
}

type modelUniformBlock struct {
	name  string
	index int32
	size  int32
	ubo   uint32
	data  []byte

	uniforms []modelUniform
	indices  map[string]uint32
}

func (block *modelUniformBlock) update() {
	gl.BindBuffer(gl.UNIFORM_BUFFER, block.ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, int(block.size), gl.Ptr(block.data), gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
}

func (block *modelUniformBlock) bind() {
	gl.BindBufferBase(gl.UNIFORM_BUFFER, uint32(block.index), block.ubo)
}

func (block *modelUniformBlock) hash() stingray.Hash {
	toHash := block.name
	for _, uniform := range block.uniforms {
		toHash += uniform.name
	}
	return stingray.Sum(toHash)
}

type modelPreviewUniformBlocks struct {
	blocks   []modelUniformBlock
	indices  map[stingray.Hash]uint32
	programs map[uint32]map[string]uint32
}

func (uniforms *modelPreviewUniformBlocks) generate(program uint32) {
	if uniforms.blocks == nil {
		uniforms.blocks = make([]modelUniformBlock, 0)
		uniforms.indices = make(map[stingray.Hash]uint32)
		uniforms.programs = make(map[uint32]map[string]uint32)
	}
	var numBlocks int32 = 0
	gl.GetProgramInterfaceiv(program, gl.UNIFORM_BLOCK, gl.ACTIVE_RESOURCES, &numBlocks)

	var blockQuery []uint32 = []uint32{gl.NUM_ACTIVE_VARIABLES, gl.NAME_LENGTH, gl.BUFFER_DATA_SIZE, gl.BUFFER_BINDING}
	var activeUniformQuery uint32 = gl.ACTIVE_VARIABLES
	var uniformQuery []uint32 = []uint32{gl.NAME_LENGTH, gl.TYPE, gl.LOCATION, gl.OFFSET, gl.ARRAY_SIZE}

	programBlockIndices, ok := uniforms.programs[program]
	if !ok {
		programBlockIndices = make(map[string]uint32)
	}
	for blockIdx := range numBlocks {
		blockProperties := make([]int32, len(blockQuery))
		gl.GetProgramResourceiv(program, gl.UNIFORM_BLOCK, uint32(blockIdx), int32(len(blockQuery)), &blockQuery[0], int32(len(blockProperties)), nil, &blockProperties[0])

		numActiveUniforms := blockProperties[0]
		if numActiveUniforms == 0 {
			continue
		}

		blockNameLen := blockProperties[1]
		blockNameData := make([]byte, blockNameLen)

		gl.GetProgramResourceName(program, gl.UNIFORM_BLOCK, uint32(blockIdx), int32(len(blockNameData)), nil, &blockNameData[0])
		blockName := string(blockNameData[:len(blockNameData)-1])

		block := modelUniformBlock{
			name:     blockName,
			index:    blockIdx,
			size:     blockProperties[2],
			ubo:      0xffffffff,
			uniforms: make([]modelUniform, 0),
			indices:  make(map[string]uint32),
		}

		blockUniforms := make([]int32, numActiveUniforms)
		gl.GetProgramResourceiv(program, gl.UNIFORM_BLOCK, uint32(blockIdx), 1, &activeUniformQuery, numActiveUniforms, nil, &blockUniforms[0])

		for _, uniformIdx := range blockUniforms {
			uniformProperties := make([]int32, len(uniformQuery))
			gl.GetProgramResourceiv(
				program, gl.UNIFORM, uint32(uniformIdx),
				int32(len(uniformQuery)), &uniformQuery[0],
				int32(len(uniformProperties)), nil, &uniformProperties[0],
			)
			if uniformProperties[0] <= 0 {
				continue
			}
			nameData := make([]byte, uniformProperties[0])
			gl.GetProgramResourceName(program, gl.UNIFORM, uint32(uniformIdx), int32(len(nameData)), nil, &nameData[0])

			name := string(nameData[:len(nameData)-1])
			uniform := modelUniform{
				name:      name,
				blockIdx:  blockIdx,
				gltype:    uniformProperties[1],
				location:  uniformProperties[2],
				offset:    uniformProperties[3],
				arraySize: uniformProperties[4],
			}
			block.indices[name] = uint32(len(block.uniforms))
			block.uniforms = append(block.uniforms, uniform)
		}

		hash := block.hash()
		if _, ok = uniforms.indices[hash]; !ok {
			gl.GenBuffers(1, &block.ubo)
			block.data = make([]byte, block.size)
			for idx := range block.uniforms {
				offset := block.uniforms[idx].offset
				size := block.uniforms[idx].size()
				block.uniforms[idx].bufSlice = block.data[offset : offset+size]
			}
			gl.BindBuffer(gl.UNIFORM_BUFFER, block.ubo)
			gl.BufferData(gl.UNIFORM_BUFFER, int(block.size), gl.Ptr(nil), gl.DYNAMIC_DRAW)
			gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
			uniforms.indices[hash] = uint32(len(uniforms.blocks))
			uniforms.blocks = append(uniforms.blocks, block)
		}
		programBlockIndices[block.name] = uniforms.indices[hash]
		gl.UniformBlockBinding(program, uint32(blockIdx), uniforms.indices[hash])
		uniforms.programs[program] = programBlockIndices
	}
}

func (uniforms *modelPreviewUniformBlocks) delete() {
	for _, block := range uniforms.blocks {
		gl.DeleteBuffers(1, &block.ubo)
		block.data = nil
	}
}

type modelPreviewUniforms map[string]modelUniform

func (uniforms *modelPreviewUniforms) generate(program uint32) {
	if *uniforms == nil {
		*uniforms = make(modelPreviewUniforms)
	}
	var numUniforms int32 = 0
	gl.GetProgramInterfaceiv(program, gl.UNIFORM, gl.ACTIVE_RESOURCES, &numUniforms)
	var properties []uint32 = []uint32{gl.BLOCK_INDEX, gl.TYPE, gl.NAME_LENGTH, gl.LOCATION}

	for uniform := range numUniforms {
		values := make([]int32, len(properties))
		gl.GetProgramResourceiv(program, gl.UNIFORM, uint32(uniform), 4, &properties[0], 4, nil, &values[0])

		// skip block uniforms
		if values[0] != -1 {
			continue
		}

		nameData := make([]byte, values[2])
		gl.GetProgramResourceName(program, gl.UNIFORM, uint32(uniform), int32(len(nameData)), nil, &nameData[0])

		name := string(nameData[:len(nameData)-1])
		(*uniforms)[name] = modelUniform{
			blockIdx: -1,
			gltype:   values[1],
			location: values[3],
		}
	}
}

type modelPreviewStateInterface interface {
	AnimTime() float32
	AnimOrigViewDistance() float32
	ViewDistance() float32
	SetViewDistance(float32)
	MaxViewDistance() float32
	AnimOrigViewRotation() mgl32.Vec2
	ViewRotation() mgl32.Vec2
	SetViewRotation(mgl32.Vec2)
	Model() mgl32.Mat4
	VFOV() float32
	AutoZoomEnabled() bool
	SetDoAutoZoomNextFrame(bool)
}

func computeMVP(pv modelPreviewStateInterface, aspectRatio float32, animate bool) (
	normal mgl32.Mat4,
	viewPosition mgl32.Vec3,
	view mgl32.Mat4,
	projection mgl32.Mat4,
) {
	var viewDistance float32
	var viewRotation mgl32.Vec2

	if animate && pv.AnimTime() >= 0 && pv.AnimTime() <= 1 {
		// Animate -> lerp original to current by animTime
		viewDistance = pv.AnimOrigViewDistance()*(1-pv.AnimTime()) + pv.ViewDistance()*pv.AnimTime()
		viewRotation = pv.AnimOrigViewRotation().Mul(1 - pv.AnimTime()).Add(pv.ViewRotation().Mul(pv.AnimTime()))
	} else {
		viewDistance = pv.ViewDistance()
		viewRotation = pv.ViewRotation()
	}

	normal = pv.Model().Inv().Transpose()
	{
		mat := mgl32.Ident3()
		mat = mat.Mul3(mgl32.Rotate3DY(viewRotation[0]))
		mat = mat.Mul3(mgl32.Rotate3DX(viewRotation[1]))
		viewPosition = mat.Mul3x1(mgl32.Vec3{0, 0, viewDistance})
	}
	view = mgl32.LookAt(
		viewPosition[0], viewPosition[1], viewPosition[2],
		0, 0, 0,
		0, 1, 0,
	)
	projection = mgl32.Perspective(
		pv.VFOV(),
		aspectRatio,
		0.001,
		32768,
	)
	return
}

func getModelPreviewProcessInputFunction(pv modelPreviewStateInterface) func() {
	return func() {
		io := imgui.CurrentIO()

		if imgui.IsItemActive() {
			md := io.MouseDelta()
			viewRotation := pv.ViewRotation().Add(mgl32.Vec2{md.X, md.Y}.Mul(-0.01))
			viewRotation[1] = mgl32.Clamp(viewRotation[1], -1.55, 1.55)
			pv.SetViewRotation(viewRotation)
		}
		if imgui.IsItemDeactivated() && pv.AutoZoomEnabled() {
			pv.SetDoAutoZoomNextFrame(true)
		}
		if imgui.IsItemHovered() {
			scroll := io.MouseWheel()
			viewDistance := pv.ViewDistance()
			viewDistance -= 0.1 * pv.ViewDistance() * scroll
			pv.SetViewDistance(viewDistance)
			if scroll != 0 {
				pv.SetDoAutoZoomNextFrame(false)
			}
		}
		viewDistance := mgl32.Clamp(
			pv.ViewDistance(),
			0.001,
			pv.MaxViewDistance(),
		)
		pv.SetViewDistance(viewDistance)
	}
}

func uploadStingrayTexture(textureID uint32, fileName stingray.Hash, getResource GetResourceFunc) error {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
	if texMain, _, err = getResource(file, stingray.DataMain); err != nil {
		return fmt.Errorf("load texture %v.texture: %w", fileName, err)
	}
	texStream, _, _ = getResource(file, stingray.DataStream)
	texGPU, _, _ = getResource(file, stingray.DataGPU)
	dataR := io.MultiReader(
		bytes.NewReader(texMain),
		bytes.NewReader(texStream),
		bytes.NewReader(texGPU),
	)
	if _, err := texture.DecodeInfo(dataR); err != nil {
		return fmt.Errorf("loading stingray DDS info: %w", err)
	}
	dds, err := dds.Decode(dataR, false)
	if err != nil {
		return fmt.Errorf("loading DDS image: %w", err)
	}
	img, ok := dds.Image.(*image.NRGBA)
	if !ok {
		return fmt.Errorf("expected texture to be of type *image.NRGBA")
	}
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return nil
}

func uploadStingrayLUT(textureID uint32, fileName stingray.Hash, getResource GetResourceFunc) error {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
	if texMain, _, err = getResource(file, stingray.DataMain); err != nil {
		return fmt.Errorf("load texture %v.texture: %w", fileName, err)
	}
	texStream, _, _ = getResource(file, stingray.DataStream)
	texGPU, _, _ = getResource(file, stingray.DataGPU)
	dataR := io.MultiReader(
		bytes.NewReader(texMain),
		bytes.NewReader(texStream),
		bytes.NewReader(texGPU),
	)
	if _, err := texture.DecodeInfo(dataR); err != nil {
		return fmt.Errorf("loading stingray DDS info: %w", err)
	}
	dds, err := dds.Decode(dataR, false)
	if err != nil {
		return fmt.Errorf("loading DDS image: %w", err)
	}
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(dds.Image.Bounds().Dx()), int32(dds.Image.Bounds().Dy()), 0, gl.RGBA, gl.FLOAT, gl.Ptr(dds.Images[0].MipMaps[0].Raw))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return nil
}

func setupTexture(textureID uint32) {
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func getTextureSlotPath(mat *material.Material, targetSlot stingray.ThinHash) stingray.Hash {
	for texSlot, path := range mat.Textures {
		if texSlot == targetSlot {
			return path
		}
	}
	return stingray.Hash{Value: 0x0}
}

func uploadMissingTexture(textureID uint32, color []byte) error {
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(color))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return nil
}

var aabbIndices = [12 * 3]uint32{
	1, 2, 0,
	1, 3, 2,
	0, 6, 4,
	0, 2, 6,
	4, 7, 5,
	4, 6, 7,
	5, 3, 1,
	5, 7, 3,
	2, 3, 7,
	2, 7, 6,
	0, 4, 5,
	0, 5, 1,
}

func getAABBVertices(aabb [2]mgl32.Vec3) [8]mgl32.Vec3 {
	return [8]mgl32.Vec3{
		{aabb[0][0], aabb[0][1], aabb[0][2]},
		{aabb[0][0], aabb[0][1], aabb[1][2]},
		{aabb[0][0], aabb[1][1], aabb[0][2]},
		{aabb[0][0], aabb[1][1], aabb[1][2]},
		{aabb[1][0], aabb[0][1], aabb[0][2]},
		{aabb[1][0], aabb[0][1], aabb[1][2]},
		{aabb[1][0], aabb[1][1], aabb[0][2]},
		{aabb[1][0], aabb[1][1], aabb[1][2]},
	}
}
