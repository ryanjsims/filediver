package particles

import (
	"encoding/binary"
	"fmt"
	"io"
)

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
