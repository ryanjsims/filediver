package previews

import (
	"bytes"
	"embed"
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

type modelPreviewUniforms map[string]int32

// Panicks if a name is not a uniform.
func (uniforms *modelPreviewUniforms) generate(program uint32, names ...string) {
	if *uniforms == nil {
		*uniforms = modelPreviewUniforms{}
	}
	for _, name := range names {
		cStr, free := gl.Strs(name + "\x00")
		loc := gl.GetUniformLocation(program, *cStr)
		free()

		if loc == -1 {
			panic(fmt.Sprintf("Invalid uniform name \"%v\" for program %v", name, program))
		}

		(*uniforms)[name] = loc
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
	normal mgl32.Mat3,
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

	normal = pv.Model().Inv().Transpose().Mat3()
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
