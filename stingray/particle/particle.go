package particle

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
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
	ControllerFormat_Unk4
	ControllerFormat_Unk5
	ControllerFormat_Unk6
	ControllerFormat_Unk7
	ControllerFormat_Box
	ControllerFormat_Unk9
	ControllerFormat_Unk10
	ControllerFormat_Unk11
	ControllerFormat_Cylinder
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
	Format ControllerFormat
	Type   ControllerType
}

func (c *ControllerHeader) GetFormat() ControllerFormat {
	return c.Format
}

type MinMaxController struct {
	ControllerHeader
	Min float32
	Max float32
}

type ConeController struct {
	ControllerHeader
	MinVelocity float32
	MaxVelocity float32
	MinTheta    float32
	MaxTheta    float32
	Axis        mgl32.Vec3
}

type Unk3Controller struct {
	ControllerHeader
	UnkInts [2]uint32
}

type Unk4Controller struct {
	ControllerHeader
	UnkVectors [2]mgl32.Vec3
}

type Unk6Controller struct {
	ControllerHeader
	UnkInt uint32
}

type Unk7Controller struct {
	ControllerHeader
	UnkInts [6]uint32
}

type BoxController struct {
	ControllerHeader
	Min mgl32.Vec3
	Max mgl32.Vec3
}

type CylinderController struct {
	ControllerHeader
	UnkInt     uint32
	MinRadial  float32
	MaxRadial  float32
	MinZ       float32
	MaxZ       float32
	MinAngular float32
	MaxAngular float32
}

type Graph struct {
	XValues [10]float32
	YValues [10]float32
}

type GradientGraph struct {
	XValues     [10]float32
	ColorValues [10]mgl32.Vec3
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
	NumComponents     uint32
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
}

type Particle struct {
	Header
	Variables       []Variable
	ParticleSystems []ParticleSystem
}
