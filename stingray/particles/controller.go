package particles

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/util"
)

type ControllerFormat uint32

const (
	ControllerFormat_Unk0 ControllerFormat = iota
	ControllerFormat_MinMax
	ControllerFormat_Cone
	ControllerFormat_Sphere
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
	ControllerType_Position  ControllerType = 0
	ControllerType_Velocity  ControllerType = 16
	ControllerType_Life      ControllerType = 36
	ControllerType_Luminance ControllerType = 40
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

type SphereController struct {
	ControllerHeader `json:"header"`
	MinRadius        util.FloatJSON `json:"min_radius"`
	MaxRadius        util.FloatJSON `json:"max_radius"`
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

type Unk13Controller struct {
	ControllerHeader `json:"header"`
	MinRadial        util.FloatJSON `json:"min_radial"`
	MaxRadial        util.FloatJSON `json:"max_radial"`
	MinZ             util.FloatJSON `json:"min_z"`
	MaxZ             util.FloatJSON `json:"max_z"`
	MinAngular       util.FloatJSON `json:"min_angular"`
	MaxAngular       util.FloatJSON `json:"max_angular"`
	UnkInts          [3]int32       `json:"unk_ints"`
	UnkFloat         util.FloatJSON `json:"unk_float"`
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
	case ControllerFormat_Sphere:
		var controller SphereController
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
	case ControllerFormat_Unk13:
		var controller Unk13Controller
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
