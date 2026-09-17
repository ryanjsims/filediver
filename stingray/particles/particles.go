package particles

import (
	"encoding/binary"
	"errors"
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
	Type VisualizerType `json:"type"`
}

func (v *VisualizerHeader) GetType() VisualizerType {
	return v.Type
}

func (v *VisualizerHeader) Simplify(_ func(stingray.Hash) string, _ func(stingray.ThinHash) string) (Visualizer, error) {
	return nil, errors.ErrUnsupported
}

type VisualizerMiddle struct {
	UnkInts [51]uint32 `json:"unk_ints"`
}

type VisualizerTrailer struct {
	UnkInt          uint32 `json:"unk_int"`
	ComponentCount  uint32 `json:"component_count"`
	ComponentOffset uint32 `json:"component_offset"`
	TotalSize       uint32 `json:"total_size"`
}

func (v *VisualizerTrailer) Size() int64 {
	return int64(v.TotalSize)
}
func (v *VisualizerTrailer) NumComponents() uint32 {
	return v.ComponentCount
}
func (v *VisualizerTrailer) ComponentsOffset() int64 {
	return int64(v.ComponentOffset)
}

type PhysicalVisualizerExtras struct {
	UnkInt      uint32 `json:"padding_int"`
	NamesOffset uint32 `json:"names_offset"`
}

func (v *PhysicalVisualizerExtras) GetNamesOffset() int64 {
	return int64(v.NamesOffset)
}

type BillboardVisualizer struct {
	VisualizerHeader
	UnkInt1  uint32
	UnkInt2  uint32
	Material stingray.Hash
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

type SimpleBillboardVisualizer struct {
	VisualizerHeader
	UnkInt1  uint32 `json:"unk_int1"`
	UnkInt2  uint32 `json:"unk_int2"`
	Material string `json:"material"`
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

func (v *BillboardVisualizer) Simplify(lookupHash func(stingray.Hash) string, _ func(stingray.ThinHash) string) (Visualizer, error) {
	return &SimpleBillboardVisualizer{
		VisualizerHeader:         v.VisualizerHeader,
		UnkInt1:                  v.UnkInt1,
		UnkInt2:                  v.UnkInt2,
		Material:                 lookupHash(v.Material),
		VisualizerMiddle:         v.VisualizerMiddle,
		VisualizerTrailer:        v.VisualizerTrailer,
		PhysicalVisualizerExtras: v.PhysicalVisualizerExtras,
	}, nil
}

type MeshVisualizer struct {
	VisualizerHeader
	Unit     stingray.Hash
	Mesh     stingray.ThinHash
	_        [4]uint8
	Material stingray.Hash
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

type SimpleMeshVisualizer struct {
	VisualizerHeader
	Unit     string `json:"unit"`
	Mesh     string `json:"mesh"`
	Material string `json:"material"`
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

func (v *MeshVisualizer) Simplify(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) (Visualizer, error) {
	return &SimpleMeshVisualizer{
		VisualizerHeader:         v.VisualizerHeader,
		Unit:                     lookupHash(v.Unit),
		Mesh:                     lookupThinHash(v.Mesh),
		Material:                 lookupHash(v.Material),
		VisualizerMiddle:         v.VisualizerMiddle,
		VisualizerTrailer:        v.VisualizerTrailer,
		PhysicalVisualizerExtras: v.PhysicalVisualizerExtras,
	}, nil
}

type Unk3Visualizer struct {
	VisualizerHeader
	UnkInt1  uint32
	UnkInt2  uint32
	Material stingray.Hash
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

type SimpleUnk3Visualizer struct {
	VisualizerHeader
	UnkInt1  uint32 `json:"unk_int1"`
	UnkInt2  uint32 `json:"unk_int2"`
	Material string `json:"material"`
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

func (v *Unk3Visualizer) Simplify(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) (Visualizer, error) {
	return &SimpleUnk3Visualizer{
		VisualizerHeader:         v.VisualizerHeader,
		UnkInt1:                  v.UnkInt1,
		UnkInt2:                  v.UnkInt2,
		Material:                 lookupHash(v.Material),
		VisualizerMiddle:         v.VisualizerMiddle,
		VisualizerTrailer:        v.VisualizerTrailer,
		PhysicalVisualizerExtras: v.PhysicalVisualizerExtras,
	}, nil
}

type TrailVisualizer struct {
	VisualizerHeader
	Material stingray.Hash
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

type SimpleTrailVisualizer struct {
	VisualizerHeader
	Material string `json:"trail_material"`
	VisualizerMiddle
	VisualizerTrailer
	PhysicalVisualizerExtras
}

func (v *TrailVisualizer) Simplify(lookupHash func(stingray.Hash) string, _ func(stingray.ThinHash) string) (Visualizer, error) {
	return &SimpleTrailVisualizer{
		VisualizerHeader:         v.VisualizerHeader,
		Material:                 lookupHash(v.Material),
		VisualizerMiddle:         v.VisualizerMiddle,
		VisualizerTrailer:        v.VisualizerTrailer,
		PhysicalVisualizerExtras: v.PhysicalVisualizerExtras,
	}, nil
}

type LightVisualizer struct {
	VisualizerHeader
	VisualizerMiddle
	UnkInt1   uint32            `json:"unk_int1"`
	UnkInt2   uint32            `json:"unk_int2"`
	UnkFloats [6]util.FloatJSON `json:"unk_floats"`
	VisualizerTrailer
}

func (v *LightVisualizer) GetNamesOffset() int64 {
	return -1
}

type UnimplementedVisualizer struct {
	VisualizerHeader
	UnkData [58]uint32 `json:"unk_data"` // The trailer should have at least this much data
}

func (v *UnimplementedVisualizer) Size() int64 {
	return -1
}
func (v *UnimplementedVisualizer) NumComponents() uint32 {
	return 0
}
func (v *UnimplementedVisualizer) ComponentsOffset() int64 {
	return -1
}
func (v *UnimplementedVisualizer) GetNamesOffset() int64 {
	return -1
}

func ReadVisualizer(r io.ReadSeeker) (Visualizer, int64, error) {
	var header VisualizerHeader
	var base int64
	var err error
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, -1, fmt.Errorf("Reading visualizer header: %v", err)
	}
	if base, err = r.Seek(0, io.SeekCurrent); err != nil {
		return nil, -1, fmt.Errorf("getting base: %v", err)
	}
	if _, err = r.Seek(-int64(binary.Size(header)), io.SeekCurrent); err != nil {
		return nil, -1, fmt.Errorf("Seeking visualizer start: %v", err)
	}
	switch header.GetType() {
	case VisualizerType_Billboard:
		var visualizer BillboardVisualizer
		if err := binary.Read(r, binary.LittleEndian, &visualizer); err != nil {
			return nil, -1, fmt.Errorf("Reading %v visualizer: %v", header.GetType().String(), err)
		}
		return &visualizer, base, nil
	case VisualizerType_Mesh:
		var visualizer MeshVisualizer
		if err := binary.Read(r, binary.LittleEndian, &visualizer); err != nil {
			return nil, -1, fmt.Errorf("Reading %v visualizer: %v", header.GetType().String(), err)
		}
		return &visualizer, base, nil
	case VisualizerType_Light:
		var visualizer LightVisualizer
		if err := binary.Read(r, binary.LittleEndian, &visualizer); err != nil {
			return nil, -1, fmt.Errorf("Reading %v visualizer: %v", header.GetType().String(), err)
		}
		return &visualizer, base, nil
	case VisualizerType_Unk3:
		var visualizer Unk3Visualizer
		if err := binary.Read(r, binary.LittleEndian, &visualizer); err != nil {
			return nil, -1, fmt.Errorf("Reading %v visualizer: %v", header.GetType().String(), err)
		}
		return &visualizer, base, nil
	case VisualizerType_Trail:
		var visualizer TrailVisualizer
		if err := binary.Read(r, binary.LittleEndian, &visualizer); err != nil {
			return nil, -1, fmt.Errorf("Reading %v visualizer: %v", header.GetType().String(), err)
		}
		return &visualizer, base, nil
	default:
		var visualizer UnimplementedVisualizer
		if err := binary.Read(r, binary.LittleEndian, &visualizer); err != nil {
			return nil, -1, fmt.Errorf("Reading %v visualizer: %v", header.GetType().String(), err)
		}
		return &visualizer, base, nil
	}
}

type ComponentsVisualizer struct {
	Visualizer
	Components []Component
	Names      []stingray.ThinHash
}

type SimpleComponentsVisualizer struct {
	Visualizer `json:"visualizer"`
	Components []Component `json:"components"`
	Names      []string    `json:"names"`
}

func (v ComponentsVisualizer) Simplify(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) (Visualizer, error) {
	visualizer, err := v.Visualizer.Simplify(lookupHash, lookupThinHash)
	if err != nil {
		return nil, err
	}
	var names []string
	if v.Names != nil {
		names = make([]string, 0)
		for _, thinhash := range v.Names {
			names = append(names, lookupThinHash(thinhash))
		}
	}
	return &SimpleComponentsVisualizer{
		Visualizer: visualizer,
		Components: v.Components,
		Names:      names,
	}, nil
}

type ComponentFormat uint32

const (
	ComponentFormat_Unk0 ComponentFormat = iota
	ComponentFormat_Unk1
	ComponentFormat_Unk2
	ComponentFormat_Unk3
	ComponentFormat_Unk4
	ComponentFormat_Unk5
	ComponentFormat_Unk6
	ComponentFormat_Unk7
	ComponentFormat_Unk8
	ComponentFormat_Unk9
	ComponentFormat_Unk10
	ComponentFormat_Unk11
	ComponentFormat_Unk12
)

func (c ComponentFormat) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ComponentFormat

type ComponentHeader struct {
	Format ComponentFormat `json:"format"`
}

func (c *ComponentHeader) GetFormat() ComponentFormat {
	return c.Format
}

type Unk4Component struct {
	ComponentHeader
	UnkInts  [4]uint32 `json:"unk_ints"`
	Graphs   [2]Graph  `json:"graphs"`
	UnkInts2 [2]uint32 `json:"unk_ints2"`
}

type Unk5Component struct {
	ComponentHeader
	UnkInts [2]uint32     `json:"unk_ints"`
	Graphs  [4]Graph      `json:"graphs"`
	Color   GradientGraph `json:"color_graph"`
	UnkInt  int32         `json:"unk_int"`
}

type Unk12Component struct {
	ComponentHeader
	UnkInts [7]uint32 `json:"unk_ints"`
}

type UnimplementedComponent struct {
	ComponentHeader
	//UnkData [1]uint32 `json:"unk_data"`
}

func ReadComponent(r io.ReadSeeker) (Component, error) {
	var format ComponentFormat
	if err := binary.Read(r, binary.LittleEndian, &format); err != nil {
		return nil, fmt.Errorf("Reading component format: %v", err)
	}
	if _, err := r.Seek(-int64(binary.Size(format)), io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("Seeking base: %v", err)
	}
	switch format {
	case ComponentFormat_Unk4:
		var component Unk4Component
		if err := binary.Read(r, binary.LittleEndian, &component); err != nil {
			return nil, fmt.Errorf("Reading %v component: %v", format, err)
		}
		return &component, nil
	case ComponentFormat_Unk5:
		var component Unk5Component
		if err := binary.Read(r, binary.LittleEndian, &component); err != nil {
			return nil, fmt.Errorf("Reading %v component: %v", format, err)
		}
		return &component, nil
	case ComponentFormat_Unk12:
		var component Unk12Component
		if err := binary.Read(r, binary.LittleEndian, &component); err != nil {
			return nil, fmt.Errorf("Reading %v component: %v", format, err)
		}
		return &component, nil
	default:
		var component UnimplementedComponent
		if err := binary.Read(r, binary.LittleEndian, &component); err != nil {
			return nil, fmt.Errorf("Reading %v component: %v", format, err)
		}
		return &component, nil
	}
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
	Size() int64
	NumComponents() uint32
	ComponentsOffset() int64
	GetNamesOffset() int64
	Simplify(func(stingray.Hash) string, func(stingray.ThinHash) string) (Visualizer, error)
}

type Component interface {
	GetFormat() ComponentFormat
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
		if _, err := r.Seek(systemBase+int64(systemHeader.VisualizerOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking system visualizers: %v", err)
		}
		for range systemHeader.VisualizerCount {
			visualizer, base, err := ReadVisualizer(r)
			if err != nil {
				return nil, err
			}
			if _, err := r.Seek(base+visualizer.ComponentsOffset(), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking visualizer components: %v", err)
			}
			components := make([]Component, 0)
			for range visualizer.NumComponents() {
				component, err := ReadComponent(r)
				if err != nil {
					return nil, fmt.Errorf("reading component: %v", err)
				}
				components = append(components, component)
			}
			var names []stingray.ThinHash
			if visualizer.GetNamesOffset() > visualizer.ComponentsOffset() {
				if _, err := r.Seek(base+visualizer.GetNamesOffset(), io.SeekStart); err != nil {
					return nil, fmt.Errorf("seeking visualizer component names: %v", err)
				}
				names = make([]stingray.ThinHash, visualizer.NumComponents())
				if err := binary.Read(r, binary.LittleEndian, names); err != nil {
					//return nil, fmt.Errorf("reading component names: %v", err)
					names = nil
				}
			}
			componentVisualizer := ComponentsVisualizer{
				Visualizer: visualizer,
				Components: components,
				Names:      names,
			}
			visualizers = append(visualizers, componentVisualizer)
			if visualizer.Size() == -1 {
				break
			}
			if _, err := r.Seek(base+visualizer.Size(), io.SeekStart); err != nil {
				return nil, fmt.Errorf("seeking next visualizer: %v", err)
			}
		}

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
