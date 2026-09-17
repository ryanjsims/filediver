package particles

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

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
