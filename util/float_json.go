package util

import (
	"encoding/json"
	"fmt"
	"math"
)

// A float type that doesn't error on marshal when it contains NaN/Inf
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

type Vec3JSON [3]FloatJSON
