package particles

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/xypwn/filediver/extractor"
	extr_material "github.com/xypwn/filediver/extractor/material"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/particles"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/util"
)

type SimpleHeader struct {
	Magic               uint32         `json:"magic"`
	MinLifetime         util.FloatJSON `json:"min_lifetime"`
	MaxLifetime         util.FloatJSON `json:"max_lifetime"`
	UnkFloat            util.FloatJSON `json:"unk_float"`
	UnkInt              uint32         `json:"unk_int"`
	VariableCount       uint32         `json:"variable_count"`
	ParticleSystemCount uint32         `json:"particle_system_count"`
	UnkCount            uint32         `json:"unk_count"`
}

type SimpleVariable struct {
	Name  string     `json:"name"`
	Value mgl32.Vec3 `json:"value"`
}

type SimpleParticleSystemHeader struct {
	SpawnLimit        uint32           `json:"spawn_limit"`
	NumControllers    uint32           `json:"num_controllers"`
	UnkInt1           uint32           `json:"unk_int1"`
	ComponentFlags    []uint32         `json:"component_flags"`
	UnkInt2           int32            `json:"unk_int2"`
	UnkInt3           int32            `json:"unk_int3"`
	UnkInt4           int32            `json:"unk_int4"`
	UnkInt5           int32            `json:"unk_int5"`
	UnkInt6           int32            `json:"unk_int6"`
	UnkInt7           int32            `json:"unk_int7"`
	UnkInt8           int32            `json:"unk_int8"`
	UnkInt9           int32            `json:"unk_int9"`
	Name1             string           `json:"name1"`
	Name2             string           `json:"name2"`
	Transform         mgl32.Mat4       `json:"transform"`
	UnkFloats         []util.FloatJSON `json:"unk_floats"`
	ControllersCount  uint32           `json:"controllers_count"`
	ControllersOffset uint32           `json:"controllers_offset"`
	EmitterCount      uint32           `json:"emitter_count"`
	EmitterOffset     uint32           `json:"emitter_offset"`
	UnkInt10          uint32           `json:"unk_int10"`
	VisualizerCount   uint32           `json:"visualizer_count"`
	VisualizerOffset  uint32           `json:"visualizer_offset"`
	TotalSize         uint32           `json:"total_size"`
	UnkInt11          uint32           `json:"unk_int11"`
}

type SimpleParticleSystem struct {
	SimpleParticleSystemHeader `json:"header"`
	Controllers                []particles.Controller `json:"controllers"`
	Emitters                   []particles.Emitter    `json:"emitters"`
	Visualizers                []particles.Visualizer `json:"visualizers"`
}

type SimpleParticle struct {
	SimpleHeader    `json:"header"`
	Variables       []SimpleVariable       `json:"variables"`
	ParticleSystems []SimpleParticleSystem `json:"particle_systems"`
}

func ExtractParticleJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	particleData, err := particles.Load(r)
	if err != nil {
		return err
	}

	variables := make([]SimpleVariable, 0)
	for _, variable := range particleData.Variables {
		variables = append(variables, SimpleVariable{
			Name:  ctx.LookupThinHash(variable.Name),
			Value: variable.Value,
		})
	}

	particleSystems := make([]SimpleParticleSystem, 0)
	for _, system := range particleData.ParticleSystems {
		floats := make([]util.FloatJSON, 0)
		for _, val := range system.UnkFloats {
			floats = append(floats, util.FloatJSON(val))
		}
		emitters := make([]particles.Emitter, 0)
		for _, emitter := range system.Emitters {
			if emitter.CanSimplify() {
				emitters = append(emitters, emitter.Simplify(ctx.LookupHash, ctx.LookupThinHash))
				continue
			}
			emitters = append(emitters, emitter)
		}
		visualizers := make([]particles.Visualizer, 0)
		for _, visualizer := range system.Visualizers {
			simplified, err := visualizer.Simplify(ctx.LookupHash, ctx.LookupThinHash)
			if errors.Is(err, errors.ErrUnsupported) {
				visualizers = append(visualizers, visualizer)
				continue
			}
			visualizers = append(visualizers, simplified)
		}
		particleSystems = append(particleSystems, SimpleParticleSystem{
			SimpleParticleSystemHeader: SimpleParticleSystemHeader{
				SpawnLimit:        system.SpawnLimit,
				NumControllers:    system.NumControllers,
				UnkInt1:           system.UnkInt1,
				ComponentFlags:    system.ComponentFlags[:],
				UnkInt2:           system.UnkInt2,
				UnkInt3:           system.UnkInt3,
				UnkInt4:           system.UnkInt4,
				UnkInt5:           system.UnkInt5,
				UnkInt6:           system.UnkInt6,
				UnkInt7:           system.UnkInt7,
				UnkInt8:           system.UnkInt8,
				UnkInt9:           system.UnkInt9,
				Name1:             ctx.LookupThinHash(system.Name1),
				Name2:             ctx.LookupThinHash(system.Name2),
				Transform:         system.Transform,
				UnkFloats:         floats,
				ControllersCount:  system.ControllersCount,
				ControllersOffset: system.ControllersOffset,
				EmitterCount:      system.EmitterCount,
				EmitterOffset:     system.EmitterOffset,
				UnkInt10:          system.UnkInt10,
				VisualizerCount:   system.VisualizerCount,
				VisualizerOffset:  system.VisualizerOffset,
				TotalSize:         system.TotalSize,
				UnkInt11:          system.UnkInt11,
			},
			Controllers: system.Controllers,
			Emitters:    emitters,
			Visualizers: visualizers,
		})
	}

	particle := SimpleParticle{
		SimpleHeader: SimpleHeader{
			Magic:               particleData.Magic,
			MinLifetime:         util.FloatJSON(particleData.MinLifetime),
			MaxLifetime:         util.FloatJSON(particleData.MaxLifetime),
			UnkFloat:            util.FloatJSON(particleData.UnkFloat),
			UnkInt:              particleData.UnkInt,
			VariableCount:       particleData.VariableCount,
			ParticleSystemCount: particleData.ParticleSystemCount,
			UnkCount:            particleData.UnkCount,
		},
		Variables:       variables,
		ParticleSystems: particleSystems,
	}

	out, err := ctx.CreateFile(".particles.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(particle); err != nil {
		return err
	}
	return nil
}

var positions [][3]float32 = [][3]float32{
	{1, 1, 0},
	{-1, 1, 0},
	{1, -1, 0},
	{-1, -1, 0},
}

var normals [][3]float32 = [][3]float32{
	{0, 0, 1},
	{0, 0, 1},
	{0, 0, 1},
	{0, 0, 1},
}

var tangents [][4]float32 = [][4]float32{
	{1, 0, 0, 1},
	{1, 0, 0, 1},
	{1, 0, 0, 1},
	{1, 0, 0, 1},
}

var texcoords [][2]float32 = [][2]float32{
	{1, 1},
	{0, 1},
	{1, 0},
	{0, 0},
}

var indices []uint32 = []uint32{
	0, 1, 3,
	0, 3, 2,
}

// Adds a 1m^2 plane at the origin
func AddPlane(doc *gltf.Document, material *uint32) uint32 {
	planeMesh := uint32(len(doc.Meshes))
	doc.Meshes = append(doc.Meshes, &gltf.Mesh{
		Primitives: []*gltf.Primitive{{
			Indices: gltf.Index(modeler.WriteIndices(doc, indices)),
			Attributes: gltf.Attribute{
				gltf.POSITION:   modeler.WritePosition(doc, positions),
				gltf.NORMAL:     modeler.WriteNormal(doc, normals),
				gltf.TANGENT:    modeler.WriteTangent(doc, tangents),
				gltf.TEXCOORD_0: modeler.WriteTextureCoord(doc, texcoords),
			},
			Material: material,
		}},
	})

	planeNode := uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: "Billboard",
		Mesh: gltf.Index(planeMesh),
	})

	return planeNode
}

func ConvertOpts(ctx *extractor.Context, imgOpts *extr_material.ImageOptions, gltfDoc *gltf.Document) error {
	f, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	cfg := ctx.Config()
	particle, err := particles.Load(f)
	if err != nil {
		return err
	}

	doc := extractor.GetDocument(ctx, gltfDoc)

	for _, system := range particle.ParticleSystems {
		for _, visualizer := range system.Visualizers {
			componentVisualizer, ok := visualizer.(particles.ComponentsVisualizer)
			if !ok {
				ctx.Warnf("Failed to cast visualizer to component visualizer, skipping")
				continue
			}
			if visualizer.GetType() == particles.VisualizerType_Billboard {
				billboard, ok := componentVisualizer.Visualizer.(*particles.BillboardVisualizer)
				if !ok {
					ctx.Warnf("Failed to cast billboard visualizer, skipping")
					continue
				}
				materialId := stingray.NewFileID(billboard.Material, stingray.Sum("material"))
				matR, err := ctx.Open(materialId, stingray.DataMain)
				if err != nil {
					return fmt.Errorf("Failed to read billboard material: %v", err)
				}
				mat, err := material.LoadMain(matR)
				if err != nil {
					return fmt.Errorf("Failed to load billboard material: %v", err)
				}
				resPath := extractor.GetMaterialName(ctx, materialId)
				matIdx, err := extr_material.AddMaterial(
					ctx.WithFileID(materialId),
					mat,
					doc,
					imgOpts,
					system.Name1,
					"billboard "+resPath,
					nil,
				)
				plane := AddPlane(doc, gltf.Index(matIdx))
				doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, plane)
			}
		}
	}

	if gltfDoc == nil {
		err := extractor.SaveDocument(ctx, doc, "particles", cfg.Model.Format)
		if err != nil {
			return err
		}
	}
	return nil
}

func Convert(currDoc *gltf.Document) func(ctx *extractor.Context) error {
	return func(ctx *extractor.Context) error {
		opts, err := extr_material.GetImageOpts(ctx)
		if err != nil {
			return err
		}
		return ConvertOpts(ctx, opts, currDoc)
	}
}
