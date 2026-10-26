package materials

import (
	"slices"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the rock fragment shader
var rockTextureNames = []string{
	"base_normal_ao_dirt",
	"detail_mask_",
	"weathering_albedo",
	"weathering_nar",
	"triplanar_detail_albedo",
	"triplanar_detail_data",
	"ibl_brdf_lut",
}

var rockShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/rock.frag",
}

var rockMaterials = []stingray.Hash{
	stingray.Sum("content/art_shared/base_shaders/rock_default"),
	stingray.Sum("content/art_shared/base_shaders/rock_emissive"),
	stingray.Sum("content/art_shared/base_shaders/rock_snow"),
	stingray.Sum("content/art_shared/base_shaders/rock_ice"),
}

func isRockMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return slices.Contains(rockMaterials, mat.BaseMaterial)
}

type RockMaterial struct {
	*BaseMaterial
}

func NewRockMaterial(id stingray.FileID, name string, shown bool) *RockMaterial {
	return &RockMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *RockMaterial) Generate() error {
	return mat.generate(
		rockShaderSources,
		0,
		append(baseUniforms, rockTextureNames...),
	)
}

var _ Material = &RockMaterial{}

func (ml *MaterialManager) updateRockMaterial(material *RockMaterial, mat *material.Material) {
	ml.addAssetGradingLut(material.BaseMaterial)
	ml.setupMaterialCommon(rockTextureNames, material, mat)
	ml.handleGlobalBlocks(material.BaseMaterial)
}
