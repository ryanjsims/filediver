package particle

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/particle"
)

type FloatJSON float32

func (f FloatJSON) MarshalJSON() ([]byte, error) {
	v := float32(f)
	if math.IsInf(float64(v), 1) {
		return []byte("\"+Inf\""), nil
	}
	if math.IsInf(float64(v), -1) {
		return []byte("\"-Inf\""), nil
	}
	if math.IsNaN(float64(v)) {
		return fmt.Appendf(nil, "\"NaN (%#x)\"", math.Float32bits(v)), nil
	}
	return json.Marshal(v)
}

type SimpleHeader struct {
	Magic               uint32    `json:"magic"`
	MinLifetime         FloatJSON `json:"min_lifetime"`
	MaxLifetime         FloatJSON `json:"max_lifetime"`
	UnkFloat            FloatJSON `json:"unk_float"`
	UnkInt              uint32    `json:"unk_int"`
	VariableCount       uint32    `json:"variable_count"`
	ParticleSystemCount uint32    `json:"particle_system_count"`
	UnkCount            uint32    `json:"unk_count"`
}

type SimpleVariable struct {
	Name  string     `json:"name"`
	Value mgl32.Vec3 `json:"value"`
}

type SimpleParticleSystemHeader struct {
	SpawnLimit        uint32      `json:"spawn_limit"`
	NumComponents     uint32      `json:"num_components"`
	UnkInt1           uint32      `json:"unk_int1"`
	ComponentFlags    []uint32    `json:"component_flags"`
	UnkInt2           int32       `json:"unk_int2"`
	UnkInt3           int32       `json:"unk_int3"`
	UnkInt4           int32       `json:"unk_int4"`
	UnkInt5           int32       `json:"unk_int5"`
	UnkInt6           int32       `json:"unk_int6"`
	UnkInt7           int32       `json:"unk_int7"`
	UnkInt8           int32       `json:"unk_int8"`
	UnkInt9           int32       `json:"unk_int9"`
	Name1             string      `json:"name1"`
	Name2             string      `json:"name2"`
	Transform         mgl32.Mat4  `json:"transform"`
	UnkFloats         []FloatJSON `json:"unk_floats"`
	ControllersCount  uint32      `json:"controllers_count"`
	ControllersOffset uint32      `json:"controllers_offset"`
	EmitterCount      uint32      `json:"emitter_count"`
	EmitterOffset     uint32      `json:"emitter_offset"`
	UnkInt10          uint32      `json:"unk_int10"`
	VisualizerCount   uint32      `json:"visualizer_count"`
	VisualizerOffset  uint32      `json:"visualizer_offset"`
	TotalSize         uint32      `json:"total_size"`
	UnkInt11          uint32      `json:"unk_int11"`
}

type SimpleParticleSystem struct {
	SimpleParticleSystemHeader `json:"header"`
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
	particleData, err := particle.Load(r)
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
		floats := make([]FloatJSON, 0)
		for _, val := range system.UnkFloats {
			floats = append(floats, FloatJSON(val))
		}
		particleSystems = append(particleSystems, SimpleParticleSystem{
			SimpleParticleSystemHeader: SimpleParticleSystemHeader{
				SpawnLimit:        system.SpawnLimit,
				NumComponents:     system.NumComponents,
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
		})
	}

	particle := SimpleParticle{
		SimpleHeader: SimpleHeader{
			Magic:               particleData.Magic,
			MinLifetime:         FloatJSON(particleData.MinLifetime),
			MaxLifetime:         FloatJSON(particleData.MaxLifetime),
			UnkFloat:            FloatJSON(particleData.UnkFloat),
			UnkInt:              particleData.UnkInt,
			VariableCount:       particleData.VariableCount,
			ParticleSystemCount: particleData.ParticleSystemCount,
			UnkCount:            particleData.UnkCount,
		},
		Variables:       variables,
		ParticleSystems: particleSystems,
	}

	out, err := ctx.CreateFile(".particle.json")
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
