package unit

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray/animation"
)

type unitPreviewTranslationKeyframe struct {
	time  float32
	value mgl32.Vec3
}

type unitPreviewRotationKeyframe struct {
	time  float32
	value mgl32.Quat
}

type unitPreviewScaleKeyframe struct {
	time  float32
	value mgl32.Vec3
}

type translationChannel struct {
	idx         uint32
	translation []unitPreviewTranslationKeyframe
}

func (t *translationChannel) Sample(time float32) mgl32.Vec3 {
	if time < t.translation[t.idx].time {
		t.idx = 0
	}
	for int(t.idx+1) < len(t.translation) && t.translation[t.idx+1].time <= time {
		t.idx = t.idx + 1
	}
	if int(t.idx+1) >= len(t.translation) {
		if int(t.idx) < len(t.translation) {
			return t.translation[t.idx].value
		}
		return mgl32.Vec3{0.0, 0.0, 0.0}
	}

	// lerp
	factor := (time - t.translation[t.idx].time) / (t.translation[t.idx+1].time - t.translation[t.idx].time)
	return t.translation[t.idx].value.Mul(1.0 - factor).Add(t.translation[t.idx+1].value.Mul(factor))
}

type rotationChannel struct {
	idx      uint32
	rotation []unitPreviewRotationKeyframe
}

func (t *rotationChannel) Sample(time float32) mgl32.Quat {
	if time < t.rotation[t.idx].time {
		t.idx = 0
	}
	for int(t.idx+1) < len(t.rotation) && t.rotation[t.idx+1].time <= time {
		t.idx = t.idx + 1
	}
	if int(t.idx+1) >= len(t.rotation) {
		if int(t.idx) < len(t.rotation) {
			return t.rotation[t.idx].value
		}
		return mgl32.QuatIdent()
	}

	// lerp
	factor := (time - t.rotation[t.idx].time) / (t.rotation[t.idx+1].time - t.rotation[t.idx].time)
	return mgl32.QuatSlerp(t.rotation[t.idx].value, t.rotation[t.idx+1].value, factor)
}

type UnitPreviewAnimation struct {
	BoneCount    uint32
	duration     float32
	translations map[uint32]translationChannel
	rotations    map[uint32]rotationChannel
	scales       map[uint32]translationChannel
}

func (a *UnitPreviewAnimation) Sample(bone uint32, time float32) mgl32.Mat4 {
	time = float32(math.Mod(float64(time), float64(a.duration)))
	translation := mgl32.Vec3{0.0, 0.0, 0.0}
	scale := mgl32.Vec3{1.0, 1.0, 1.0}
	rotation := mgl32.QuatIdent()

	if channel, ok := a.translations[bone]; ok {
		translation = channel.Sample(time)
	}
	if channel, ok := a.scales[bone]; ok {
		scale = channel.Sample(time)
	}
	if channel, ok := a.rotations[bone]; ok {
		rotation = channel.Sample(time)
	}

	return mgl32.Translate3D(translation.Elem()).Mul4(rotation.Mat4()).Mul4(mgl32.Scale3D(scale.Elem()))
}

func (a *UnitPreviewAnimation) Load(info *animation.Animation) {
	a.translations = make(map[uint32]translationChannel)
	a.rotations = make(map[uint32]rotationChannel)
	a.scales = make(map[uint32]translationChannel)
	a.BoneCount = info.Header.BoneCount
	a.duration = info.Header.AnimationLength
	for idx, keyframe := range info.Header.InitialTransforms {
		a.translations[uint32(idx)] = translationChannel{
			idx: uint32(idx),
			translation: []unitPreviewTranslationKeyframe{{
				time:  0.0,
				value: keyframe.Position(),
			}},
		}
		a.rotations[uint32(idx)] = rotationChannel{
			idx: uint32(idx),
			rotation: []unitPreviewRotationKeyframe{{
				time:  0.0,
				value: keyframe.Rotation(),
			}},
		}
		a.scales[uint32(idx)] = translationChannel{
			idx: uint32(idx),
			translation: []unitPreviewTranslationKeyframe{{
				time:  0.0,
				value: keyframe.Scale(),
			}},
		}
	}

	for _, keyframe := range info.Entries {
		translationChannel, _ := a.translations[uint32(keyframe.Header.Bone())]
		rotationChannel, _ := a.rotations[uint32(keyframe.Header.Bone())]
		scaleChannel, _ := a.scales[uint32(keyframe.Header.Bone())]
		if position, err := keyframe.Position(); err == nil {
			translationChannel.translation = append(translationChannel.translation, unitPreviewTranslationKeyframe{
				time:  float32(keyframe.Header.TimeMS()) / 1000.0,
				value: position,
			})
			a.translations[uint32(keyframe.Header.Bone())] = translationChannel
		} else if rotation, err := keyframe.Rotation(); err == nil {
			rotationChannel.rotation = append(rotationChannel.rotation, unitPreviewRotationKeyframe{
				time:  float32(keyframe.Header.TimeMS()) / 1000.0,
				value: rotation,
			})
			a.rotations[uint32(keyframe.Header.Bone())] = rotationChannel
		} else if scale, err := keyframe.Scale(); err == nil {
			scaleChannel.translation = append(scaleChannel.translation, unitPreviewTranslationKeyframe{
				time:  float32(keyframe.Header.TimeMS()) / 1000.0,
				value: scale,
			})
			a.scales[uint32(keyframe.Header.Bone())] = scaleChannel
		}
	}
}
