package particle

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

type ControllerFormat uint32

const (
	ControllerFormat_Unk0 ControllerFormat = iota
	ControllerFormat_MinMax
	ControllerFormat_Cone
	ControllerFormat_Unk3
	ControllerFormat_Box
	ControllerFormat_Unk5
	ControllerFormat_Unk6
	ControllerFormat_Unk7
	ControllerFormat_Unk8
	ControllerFormat_Unk9
	ControllerFormat_Unk10
	ControllerFormat_Unk11
	ControllerFormat_Cylinder
	ControllerFormat_Unk13
	ControllerFormat_Unk14
	ControllerFormat_Unk15
	ControllerFormat_Unk16
	ControllerFormat_Unk17
	ControllerFormat_Unk18
	ControllerFormat_Unk19
	ControllerFormat_Unk20
	ControllerFormat_Unk21
	ControllerFormat_Unk22
)

func (c ControllerFormat) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ControllerFormat

type ControllerType uint32

const (
	ControllerType_Luminance ControllerType = 40
	ControllerType_Life      ControllerType = 36
)

func (c ControllerType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ControllerType

type ControllerHeader struct {
	Format ControllerFormat `json:"format"`
	Type   ControllerType   `json:"type"`
}

func (c *ControllerHeader) GetFormat() ControllerFormat {
	return c.Format
}

type Unk0Controller struct {
	ControllerHeader `json:"header"`
	UnkInt           int32 `json:"unk_int"`
}

type MinMaxController struct {
	ControllerHeader `json:"header"`
	Min              util.FloatJSON `json:"min"`
	Max              util.FloatJSON `json:"max"`
}

type ConeController struct {
	ControllerHeader `json:"header"`
	//UnkFloats        [2]util.FloatJSON `json:"unk_floats"`
	MinVelocity util.FloatJSON `json:"min_velocity"`
	MaxVelocity util.FloatJSON `json:"max_velocity"`
	MinTheta    util.FloatJSON `json:"min_theta"`
	MaxTheta    util.FloatJSON `json:"max_theta"`
	Axis        util.Vec3JSON  `json:"axis"`
	UnkGraph1   Graph          `json:"graph1"`
	UnkGraph2   Graph          `json:"graph2"`
	UnkInts     [4]int32       `json:"unk_ints"`
}

type Unk3Controller struct {
	ControllerHeader `json:"header"`
	UnkFloats        [2]util.FloatJSON `json:"unk_floats"`
}

type BoxController struct {
	ControllerHeader `json:"header"`
	Min              util.Vec3JSON `json:"min"`
	Max              util.Vec3JSON `json:"max"`
}

// Might be a cylinder controller?
type Unk6Controller struct {
	ControllerHeader `json:"header"`
	UnkInts          [2]uint32 `json:"unk_ints"`
}

type Unk7Controller struct {
	ControllerHeader `json:"header"`
	UnkInts          [6]uint32 `json:"unk_ints"`
}

type Unk8Controller struct {
	ControllerHeader `json:"header"`
	Min              util.Vec3JSON `json:"min"`
	Max              util.Vec3JSON `json:"max"`
}

type CylinderController struct {
	ControllerHeader `json:"header"`
	MinRadial        util.FloatJSON `json:"min_radial"`
	MaxRadial        util.FloatJSON `json:"max_radial"`
	MinZ             util.FloatJSON `json:"min_z"`
	MaxZ             util.FloatJSON `json:"max_z"`
	MinAngular       util.FloatJSON `json:"min_angular"`
	MaxAngular       util.FloatJSON `json:"max_angular"`
	UnkInts          [3]int32       `json:"unk_ints"`
}

type Unk11Controller struct {
	ControllerHeader `json:"header"`
	UnkVectors       [2]util.Vec3JSON  `json:"unk_vectors"`
	UnkInts          [3]uint32         `json:"unk_ints"`
	UnkFloats        [2]util.FloatJSON `json:"unk_floats"`
}

type Unk14Controller struct {
	ControllerHeader `json:"header"`
	UnkInt           uint32            `json:"unk_int"`
	UnkFloats        [6]util.FloatJSON `json:"unk_floats"`
	UnkGraph1        Graph             `json:"unk_graph1"`
	UnkGraph2        Graph             `json:"unk_graph2"`
	UnkInts          [4]int32          `json:"unk_ints"`
}

type Unk18Controller struct {
	ControllerHeader `json:"header"`
	UnkFloat         util.FloatJSON `json:"unk_float"`
}

type Unk22Controller struct {
	ControllerHeader `json:"header"`
	UnkFloats        [2]util.FloatJSON `json:"unk_floats"`
	UnkVector        util.Vec3JSON     `json:"unk_vector"`
}

type UnimplementedController struct {
	ControllerHeader `json:"header"`
}

func ReadController(r io.ReadSeeker, format ControllerFormat) (Controller, error) {
	switch format {
	case ControllerFormat_Unk0:
		var controller Unk0Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_MinMax:
		var controller MinMaxController
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Cone:
		var controller ConeController
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Unk3:
		var controller Unk3Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Box:
		var controller BoxController
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	// case ControllerFormat_Unk5:
	// 	var controller Unk5Controller
	// 	if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
	// 		return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
	// 	}
	// 	return &controller, nil
	case ControllerFormat_Unk6:
		var controller Unk6Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	// case ControllerFormat_Unk7:
	// 	var controller Unk7Controller
	// 	if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
	// 		return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
	// 	}
	// 	return &controller, nil
	case ControllerFormat_Unk8:
		var controller BoxController
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	// case ControllerFormat_Unk9:
	// 	var controller Unk9Controller
	// 	if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
	// 		return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
	// 	}
	// 	return &controller, nil
	// case ControllerFormat_Unk10:
	// 	var controller Unk10Controller
	// 	if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
	// 		return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
	// 	}
	// 	return &controller, nil
	case ControllerFormat_Unk11:
		var controller Unk11Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Cylinder:
		var controller CylinderController
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Unk14:
		var controller Unk14Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Unk18:
		var controller Unk18Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	case ControllerFormat_Unk22:
		var controller Unk22Controller
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("Reading controller of format %v: %v", format, err)
		}
		return &controller, nil
	default:
		var controller UnimplementedController
		if err := binary.Read(r, binary.LittleEndian, &controller); err != nil {
			return nil, fmt.Errorf("skipping unimplemented controller format %v: %v", format.String(), err)
		}
		return &controller, nil
	}
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
	EmitterType_Rate EmitterType = iota + 10
	EmitterType_Burst
)

func (c EmitterType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EmitterType

type EmitterHeader struct {
	Type EmitterType
}

func (e *EmitterHeader) GetType() EmitterType {
	return e.Type
}

// This might be something a bit more interesting like a root node?
// idk it "works" but it doesn't seem right
type ZeroEmitter struct {
	EmitterHeader
	UnkInts [5]uint32
}

type BurstRow struct {
	Time float32
	Min  uint32
	Max  uint32
}

type BurstEmitter struct {
	EmitterHeader
	Rows [10]BurstRow
}

type RateEmitter struct {
	EmitterHeader
	Min     float32
	Max     float32
	Graphs  [2]Graph
	UnkInts [5]int32
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
