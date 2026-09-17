package particles

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

type Header struct {
	Magic               uint32
	MinLifetime         float32
	MaxLifetime         float32
	UnkFloat            float32
	UnkInt              uint32
	VariableCount       uint32
	ParticleSystemCount uint32
	UnkCount            uint32
	_                   [48]uint8
}

type Variable struct {
	Name  stingray.ThinHash
	Value mgl32.Vec3
}

type Graph struct {
	XValues [10]util.FloatJSON `json:"x"`
	YValues [10]util.FloatJSON `json:"y"`
}

type GradientGraph struct {
	XValues     [10]util.FloatJSON `json:"x"`
	ColorValues [10]util.Vec3JSON  `json:"color"`
}

type EmitterType uint32

const (
	EmitterType_Unk0 EmitterType = iota
	EmitterType_Unk1
	EmitterType_Unk2
	EmitterType_Acceleration
	EmitterType_Unk4
	EmitterType_Unk5
	EmitterType_Unk6
	EmitterType_Unk7
	EmitterType_Unk8
	EmitterType_Unk9
	EmitterType_Unk10
	EmitterType_Rate
	EmitterType_Burst
	EmitterType_Unk13
	EmitterType_Unk14
	EmitterType_Unk15
	EmitterType_Unk16
	EmitterType_Unk17
	EmitterType_Unk18
	EmitterType_Unk19
	EmitterType_Unk20
	EmitterType_Unk21
	EmitterType_Unk22
	EmitterType_Unk23
	EmitterType_Unk24
	EmitterType_Unk25
	EmitterType_Unk26
	EmitterType_Unk27
	EmitterType_Unk28
	EmitterType_Unk29
	EmitterType_Unk30
	EmitterType_Unk31
	EmitterType_Unk32
	EmitterType_Unk33
	EmitterType_Unk34
	EmitterType_Unk35
	EmitterType_Unk36
	EmitterType_Unk37
)

func (c EmitterType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EmitterType

type EmitterHeader struct {
	Type EmitterType `json:"type"`
}

func (e *EmitterHeader) GetType() EmitterType {
	return e.Type
}

func (e *EmitterHeader) CanSimplify() bool {
	return e.Type == EmitterType_Unk33
}

func (e *EmitterHeader) Simplify(_ func(stingray.Hash) string, _ func(stingray.ThinHash) string) Emitter {
	return nil
}

// This might be something a bit more interesting like a root node?
// idk it "works" but it doesn't seem right
type Unk0Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [5]int32 `json:"unk_ints"`
}

type Unk1Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [3]int32 `json:"unk_ints"`
}

type Unk2Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [4]int32 `json:"unk_ints"`
	Graphs        [2]Graph `json:"graphs"`
	UnkInts2      [3]int32 `json:"unk_ints2"`
}

type AccelerationEmitter struct {
	EmitterHeader `json:"header"`
	UnkInt        int32         `json:"unk_int"`
	Acceleration  util.Vec3JSON `json:"acceleration"`
	UnkInts       [2]int32      `json:"unk_ints"`
}

type Unk4Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [2]int32       `json:"unk_ints"`
	UnkFloat      util.FloatJSON `json:"unk_float"`
}

type Unk5Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [4]int32 `json:"unk_ints"`
	Graphs        [2]Graph `json:"graphs"`
}

type Unk8Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [5]int32       `json:"unk_ints"`
	UnkFloat      util.FloatJSON `json:"unk_float"`
	UnkInts2      [4]int32       `json:"unk_ints2"`
}

type Unk13Emitter struct {
	EmitterHeader `json:"header"`
	UnkFloat      util.FloatJSON `json:"unk_float"`
	UnkInts       [2]int32       `json:"unk_ints"`
}

type Unk17Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [4]int32 `json:"unk_ints"`
}

type Unk18Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [5]int32 `json:"unk_ints"`
}

type Unk19Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [3]int32 `json:"unk_ints"`
}

type Unk20Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [2]int32 `json:"unk_ints"`
}

type Unk21Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [6]int32 `json:"unk_ints"`
	Graphs        [2]Graph `json:"graphs"`
	UnkInt        int32    `json:"unk_int"`
}

type Unk25Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [4]int32 `json:"unk_ints"`
	Graphs        [2]Graph `json:"graphs"`
	UnkInt        int32    `json:"unk_int"`
}

type Unk26Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [2]int32      `json:"unk_ints"`
	UnkVector     util.Vec3JSON `json:"unk_vector"`
}

type Unk28Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [3]int32       `json:"unk_ints"`
	UnkFloat      util.FloatJSON `json:"unk_float"`
	Graphs        [2]Graph       `json:"graphs"`
	UnkInts2      [2]int32       `json:"unk_ints2"`
}

type Unk29Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [8]int32 `json:"unk_ints"`
	Graphs        [2]Graph `json:"graphs"`
}

type Unk30Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [3]int32       `json:"unk_ints"`
	UnkFloat      util.FloatJSON `json:"unk_float"`
}

type Unk32Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [3]int32 `json:"unk_ints"`
}

type SimpleUnk33Emitter struct {
	EmitterHeader `json:"header"`
	UnkThinHash   string   `json:"unk_thinhash"`
	UnkInts       [4]int32 `json:"unk_ints"`
}

type Unk33Emitter struct {
	EmitterHeader
	UnkThinHash stingray.ThinHash
	UnkInts     [4]int32
}

func (e *Unk33Emitter) Simplify(_ func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) Emitter {
	return &SimpleUnk33Emitter{
		EmitterHeader: e.EmitterHeader,
		UnkThinHash:   lookupThinHash(e.UnkThinHash),
		UnkInts:       e.UnkInts,
	}
}

type Unk34Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [9]int32 `json:"unk_ints"`
	Graphs        [2]Graph `json:"graphs"`
}

type Unk35Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [2]int32 `json:"unk_ints"`
}

type Unk36Emitter struct {
	EmitterHeader `json:"header"`
	UnkInts       [4]int32 `json:"unk_ints"`
	UnkStructs    [10]struct {
		Float  util.FloatJSON `json:"float"`
		Ints   [2]int32       `json:"ints"`
		Float2 util.FloatJSON `json:"float2"`
		Int    int32          `json:"int"`
		Float3 util.FloatJSON `json:"float3"`
	} `json:"structs"`
}

type Unk37Emitter struct {
	EmitterHeader `json:"header"`
	UnkInt        int32             `json:"unk_ints"`
	UnkFloats     [4]util.FloatJSON `json:"unk_floats"`
}

type BurstRow struct {
	Time util.FloatJSON `json:"time"`
	Min  uint32         `json:"min_spawned"`
	Max  uint32         `json:"max_spawned"`
}

type BurstEmitter struct {
	EmitterHeader `json:"header"`
	Rows          [10]BurstRow `json:"burst_entries"`
	UnkInt        [3]uint32    `json:"unk_ints"`
}

type RateEmitter struct {
	EmitterHeader `json:"header"`
	Min           util.FloatJSON `json:"min"`
	Max           util.FloatJSON `json:"max"`
	Graphs        [2]Graph       `json:"graphs"`
	UnkInts       [5]int32       `json:"unk_ints"`
}

type UnimplementedEmitter struct {
	EmitterHeader `json:"header"`
}

func ReadEmitter(r io.ReadSeeker) (Emitter, error) {
	var header EmitterHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("Reading emitter header: %v", err)
	}
	if _, err := r.Seek(-int64(binary.Size(header)), io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("Seeking back to emitter header: %v", err)
	}
	switch header.GetType() {
	case EmitterType_Unk0:
		var emitter Unk0Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk1:
		var emitter Unk1Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk2:
		var emitter Unk2Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Acceleration:
		var emitter AccelerationEmitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk4:
		var emitter Unk4Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk5:
		var emitter Unk5Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk8:
		var emitter Unk8Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk13:
		var emitter Unk13Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk17:
		var emitter Unk17Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk18:
		var emitter Unk18Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk19:
		var emitter Unk19Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk20:
		var emitter Unk20Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk21:
		var emitter Unk21Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk25:
		var emitter Unk25Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk26:
		var emitter Unk26Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk28:
		var emitter Unk28Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk29:
		var emitter Unk29Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk30:
		var emitter Unk30Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk32:
		var emitter Unk32Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk33:
		var emitter Unk33Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk34:
		var emitter Unk34Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk35:
		var emitter Unk35Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk36:
		var emitter Unk36Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Unk37:
		var emitter Unk37Emitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Burst:
		var emitter BurstEmitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	case EmitterType_Rate:
		var emitter RateEmitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	default:
		var emitter UnimplementedEmitter
		if err := binary.Read(r, binary.LittleEndian, &emitter); err != nil {
			return nil, fmt.Errorf("Reading emitter %v: %v", header.GetType(), err)
		}
		return &emitter, nil
	}
}

type VisualizerType uint32

const (
	VisualizerType_Billboard VisualizerType = iota
	VisualizerType_Light
	VisualizerType_Mesh
	VisualizerType_Unk3
	VisualizerType_Trail
)

func (c VisualizerType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=VisualizerType

type VisualizerHeader struct {
	Type VisualizerType
}

func (v *VisualizerHeader) GetType() VisualizerType {
	return v.Type
}

type VisualizerTrailer struct {
	UnkInts    [52]uint32
	UnkCount   uint32
	UnkOffset  uint32
	UnkOffset2 uint32
	UnkCount3  uint32
	UnkOffset3 uint32
	TotalSize  uint32
}

func (v *VisualizerTrailer) Size() uint32 {
	return v.TotalSize
}

type BillboardVisualizer struct {
	VisualizerHeader
	UnkInt1  uint32
	UnkInt2  uint32
	Material stingray.Hash
	VisualizerTrailer
}

type MeshVisualizer struct {
	VisualizerHeader
	Unit     stingray.Hash
	Mesh     stingray.ThinHash
	_        [4]uint8
	Material stingray.Hash
	VisualizerTrailer
}

type TrailVisualizer struct {
	VisualizerHeader
	Material stingray.Hash
	VisualizerTrailer
}

type Controller interface {
	GetFormat() ControllerFormat
}

type Emitter interface {
	GetType() EmitterType
	CanSimplify() bool
	Simplify(func(stingray.Hash) string, func(stingray.ThinHash) string) Emitter
}

type Visualizer interface {
	GetType() VisualizerType
	Size() uint32
}

type ParticleSystemHeader struct {
	SpawnLimit        uint32
	NumControllers    uint32
	UnkInt1           uint32
	ComponentFlags    [16]uint32
	UnkInt2           int32
	UnkInt3           int32
	UnkInt4           int32
	UnkInt5           int32
	UnkInt6           int32
	UnkInt7           int32
	UnkInt8           int32
	UnkInt9           int32
	Name1             stingray.ThinHash
	Name2             stingray.ThinHash
	_                 [4]uint8
	Transform         mgl32.Mat4
	UnkFloats         [11]float32
	ControllersCount  uint32
	ControllersOffset uint32
	EmitterCount      uint32
	EmitterOffset     uint32
	UnkInt10          uint32
	VisualizerCount   uint32
	VisualizerOffset  uint32
	TotalSize         uint32
	UnkInt11          uint32
}

type ParticleSystem struct {
	ParticleSystemHeader
	Controllers []Controller
	Emitters    []Emitter
	Visualizers []Visualizer
}

type Particle struct {
	Header
	Variables       []Variable
	ParticleSystems []ParticleSystem
}

func Load(r io.ReadSeeker) (*Particle, error) {
	// base, err := r.Seek(0, io.SeekCurrent)
	// if err != nil {
	// 	return nil, err
	// }

	var header Header
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("reading header: %v", err)
	}

	variableNames := make([]stingray.ThinHash, header.VariableCount)
	if err := binary.Read(r, binary.LittleEndian, variableNames); err != nil {
		return nil, fmt.Errorf("reading variable names: %v", err)
	}

	variableValues := make([]mgl32.Vec3, header.VariableCount)
	if err := binary.Read(r, binary.LittleEndian, variableValues); err != nil {
		return nil, fmt.Errorf("reading variable values: %v", err)
	}

	variables := make([]Variable, 0)
	for i := range header.VariableCount {
		variables = append(variables, Variable{
			Name:  variableNames[i],
			Value: variableValues[i],
		})
	}

	particleSystems := make([]ParticleSystem, 0)
	for range header.ParticleSystemCount {
		systemBase, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("seeking system base: %v", err)
		}
		var systemHeader ParticleSystemHeader
		if err := binary.Read(r, binary.LittleEndian, &systemHeader); err != nil {
			return nil, fmt.Errorf("reading particle system header: %v", err)
		}

		// read particle details here...
		controllers := make([]Controller, 0)
		if _, err := r.Seek(systemBase+int64(systemHeader.ControllersOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking system controllers: %v", err)
		}
		for range systemHeader.ControllersCount {
			var format ControllerFormat
			var typ ControllerType
			if err := binary.Read(r, binary.LittleEndian, &format); err != nil {
				return nil, fmt.Errorf("reading controller format: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
				return nil, fmt.Errorf("reading controller type: %v", err)
			}

			if _, err := r.Seek(-8, io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("reseeking current controller: %v", err)
			}

			controller, err := ReadController(r, format)
			if err != nil {
				return nil, err
			}

			controllers = append(controllers, controller)
		}

		emitters := make([]Emitter, 0)
		if _, err := r.Seek(systemBase+int64(systemHeader.EmitterOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking system emitters: %v", err)
		}
		for range systemHeader.EmitterCount {
			emitter, err := ReadEmitter(r)
			if err != nil {
				return nil, err
			}

			emitters = append(emitters, emitter)
		}

		visualizers := make([]Visualizer, 0)

		if _, err := r.Seek(systemBase+int64(systemHeader.TotalSize), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking next system header: %v", err)
		}
		particleSystems = append(particleSystems, ParticleSystem{
			ParticleSystemHeader: systemHeader,
			Controllers:          controllers,
			Emitters:             emitters,
			Visualizers:          visualizers,
		})
	}

	return &Particle{
		Header:          header,
		Variables:       variables,
		ParticleSystems: particleSystems,
	}, nil
}
