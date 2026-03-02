package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type SprayWeaponStateInfo struct {
	Shake             WeaponCameraShakeInfo   // Settings for local and in-world camera shakes to play in this state.
	MuzzleEffect      stingray.Hash           // [particles]Particle effect to play from the muzzle in this state.
	CalloutEventStart stingray.ThinHash       // [string]Audio event to player when entering this state.
	AudioEventStart   stingray.ThinHash       // [string]Audio event to player when entering this state.
	AudioEventStop    stingray.ThinHash       // [string]Audio event to player when exiting this state.
	WindTemplate      enum.WindEffectTemplate // Audio event to player when exiting this state.
}

type SprayWeaponComponent struct {
	WarmupTime           float32           // The time between starting to fire a cold weapon and the spray.
	CooldownTime         float32           // The time between stopping fire until the weapon is considered cold again (and re-firing is not instant).
	UnkFloat             float32           // Name length 19
	AudioNode            stingray.ThinHash // [string]Node we play warmup, fire, and cooldown audio from.
	SprayCapsuleRadius   float32           // If greater than 0, it's uses physics overlap test instead of relying on the projectiles hitting. Experimental!
	_                    [4]uint8
	Warmup               SprayWeaponStateInfo     // Settings for warmup state.
	Firing               SprayWeaponStateInfo     // Settings for firing state.
	Cooldown             SprayWeaponStateInfo     // Settings for cooldown state.
	AmmoCost             uint32                   // How much ammo is taken away per spray blob
	ProjectileType       enum.ProjectileType      // If not set to None, this weapon will use projectiles.
	DamageInfoType       enum.DamageInfoType      // If not set to None (and projectiles are None), this will use collision events.
	EffectType           enum.HitEffectDamageType // When using collision events, this dictates what kind of damage effect to play on hits.
	UnkBool              uint8                    // Unknown, name length 32
	_                    [3]uint8
	LastDropOwnerVoEvent stingray.ThinHash // [string]VO event to play on the owner of the wepaon when the last bullet has been fired.
	DryFireAudioEvent    stingray.ThinHash // [string]A sound event to play when the ammo is 0 and the player is trying to shoot.
	FireTemplate         enum.FireTemplate // fire template for the weapon
}

type SimpleSprayWeaponStateInfo struct {
	Shake             SimpleWeaponCameraShakeInfo `json:"shake"`
	MuzzleEffect      string                      `json:"muzzle_effect"`
	CalloutEventStart string                      `json:"callout_event_start"`
	AudioEventStart   string                      `json:"audio_event_start"`
	AudioEventStop    string                      `json:"audio_event_stop"`
	WindTemplate      enum.WindEffectTemplate     `json:"wind_template"`
}

func (s SprayWeaponStateInfo) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleSprayWeaponStateInfo {
	return SimpleSprayWeaponStateInfo{
		Shake:             s.Shake.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		MuzzleEffect:      lookupHash(s.MuzzleEffect),
		CalloutEventStart: lookupThinHash(s.CalloutEventStart),
		AudioEventStart:   lookupThinHash(s.AudioEventStart),
		AudioEventStop:    lookupThinHash(s.AudioEventStop),
		WindTemplate:      s.WindTemplate,
	}
}

type SimpleSprayWeaponComponent struct {
	WarmupTime           float32                    `json:"warmup_time"`
	CooldownTime         float32                    `json:"cooldown_time"`
	UnkFloat             float32                    `json:"unk_float"`
	AudioNode            string                     `json:"audio_node"`
	SprayCapsuleRadius   float32                    `json:"spray_capsule_radius"`
	Warmup               SimpleSprayWeaponStateInfo `json:"warmup"`
	Firing               SimpleSprayWeaponStateInfo `json:"firing"`
	Cooldown             SimpleSprayWeaponStateInfo `json:"cooldown"`
	AmmoCost             uint32                     `json:"ammo_cost"`
	ProjectileType       enum.ProjectileType        `json:"projectile_type"`
	DamageInfoType       enum.DamageInfoType        `json:"damage_info_type"`
	EffectType           enum.HitEffectDamageType   `json:"effect_type"`
	UnkBool              bool                       `json:"unk_bool"`
	LastDropOwnerVoEvent string                     `json:"last_drop_owner_vo_event"`
	DryFireAudioEvent    string                     `json:"dry_fire_audio_event"`
	FireTemplate         enum.FireTemplate          `json:"fire_template"`
}

func (component SprayWeaponComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleSprayWeaponComponent{
		WarmupTime:           component.WarmupTime,
		CooldownTime:         component.CooldownTime,
		UnkFloat:             component.UnkFloat,
		AudioNode:            lookupThinHash(component.AudioNode),
		SprayCapsuleRadius:   component.SprayCapsuleRadius,
		Warmup:               component.Warmup.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		Firing:               component.Firing.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		Cooldown:             component.Cooldown.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		AmmoCost:             component.AmmoCost,
		ProjectileType:       component.ProjectileType,
		DamageInfoType:       component.DamageInfoType,
		EffectType:           component.EffectType,
		UnkBool:              component.UnkBool != 0,
		LastDropOwnerVoEvent: lookupThinHash(component.LastDropOwnerVoEvent),
		DryFireAudioEvent:    lookupThinHash(component.DryFireAudioEvent),
		FireTemplate:         component.FireTemplate,
	}
}

func getSprayWeaponComponentData() ([]byte, error) {
	sprayWeaponHash := Sum("SprayWeaponComponentData")
	sprayWeaponHashData := make([]byte, 4)
	if _, err := binary.Encode(sprayWeaponHashData, binary.LittleEndian, sprayWeaponHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, sprayWeaponHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getSprayWeaponComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	sprayWeaponCmpDataHash := Sum("SprayWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var sprayWeaponCmpDataType DLTypeDesc
	var ok bool
	sprayWeaponCmpDataType, ok = typelib.Types[sprayWeaponCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find SprayWeaponComponentData hash in dl_library")
	}

	if len(sprayWeaponCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(sprayWeaponCmpDataType.Members))
	}

	if sprayWeaponCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if sprayWeaponCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if sprayWeaponCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if sprayWeaponCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (data storage was not struct)")
	}

	if sprayWeaponCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if sprayWeaponCmpDataType.Members[1].TypeID != Sum("SprayWeaponComponent") {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (data type was not SprayWeaponComponent)")
	}

	sprayWeaponComponentData, err := getSprayWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get spray weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(sprayWeaponComponentData)

	hashmap := make([]ComponentIndexData, sprayWeaponCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	var index int32 = -1
	for _, entry := range hashmap {
		if entry.Resource == hash {
			index = int32(entry.Index)
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("%v not found in spray weapon component data", hash.String())
	}

	var sprayWeaponComponentType DLTypeDesc
	sprayWeaponComponentType, ok = typelib.Types[Sum("SprayWeaponComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find SprayWeaponComponent hash in dl_library")
	}

	componentData := make([]byte, sprayWeaponComponentType.Size)
	if _, err := r.Seek(int64(sprayWeaponComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseSprayWeaponComponents() (map[stingray.Hash]SprayWeaponComponent, error) {
	sprayWeaponHash := Sum("SprayWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var sprayWeaponType DLTypeDesc
	var ok bool
	sprayWeaponType, ok = typelib.Types[sprayWeaponHash]
	if !ok {
		return nil, fmt.Errorf("could not find SprayWeaponComponentData hash in dl_library")
	}

	if len(sprayWeaponType.Members) != 2 {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(sprayWeaponType.Members))
	}

	if sprayWeaponType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if sprayWeaponType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if sprayWeaponType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if sprayWeaponType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (data storage was not struct)")
	}

	if sprayWeaponType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if sprayWeaponType.Members[1].TypeID != Sum("SprayWeaponComponent") {
		return nil, fmt.Errorf("SprayWeaponComponentData unexpected format (data type was not SprayWeaponComponent)")
	}

	sprayWeaponComponentData, err := getSprayWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get spray weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(sprayWeaponComponentData)

	hashmap := make([]ComponentIndexData, sprayWeaponType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]SprayWeaponComponent, sprayWeaponType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]SprayWeaponComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
