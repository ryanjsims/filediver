package previews

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
	"github.com/xypwn/filediver/stingray/shading_environment"
)

type EntityPreview struct {
	*entity.Entity

	name             stingray.FileID
	lookupHash       func(stingray.Hash) string
	lookupThinHash   func(stingray.ThinHash) string
	entityVarMapping shading_environment.ShadingEnvironmentEntityToShaderMapping
}

func NewEntityPreview(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string, entityVarMapping shading_environment.ShadingEnvironmentEntityToShaderMapping) *EntityPreview {
	return &EntityPreview{
		lookupHash:       lookupHash,
		lookupThinHash:   lookupThinHash,
		entityVarMapping: entityVarMapping,
	}
}

func (pv *EntityPreview) drawComponentMap(component entity.Component) {
	var categoryKey string
	for _, key := range component.CategoryNames {
		categoryKey += key.String()
	}
	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsRowBg
	if imgui.BeginTableV("##Category Names "+categoryKey, 1, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Category", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		for _, key := range component.CategoryNames {
			imgui.PushIDInt(int32(key.Value))
			imgui.TableNextColumn()
			imutils.CopyableTextf("%v##key", pv.lookupThinHash(key))
			imgui.PopID()
		}
		imgui.EndTable()
	}

	if imgui.BeginTableV("##Settings "+categoryKey, 3, tableFlags|imgui.TableFlagsScrollY, imgui.NewVec2(0, 200.0), 0) {
		imgui.TableSetupColumnV("Setting", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("Shader Name", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("Value", imgui.TableColumnFlagsWidthStretch, 2, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		for idx, setting := range component.Settings {
			key := component.SettingNames[idx]
			shaderMapping, ok := pv.entityVarMapping[key]
			var shaderName string = "<none>"
			if ok {
				shaderName = pv.lookupThinHash(shaderMapping.ShaderName)
			}
			imgui.PushIDInt(int32(key.Value))
			imgui.TableNextColumn()
			imutils.CopyableTextf("%v##key", pv.lookupThinHash(key))
			imgui.TableNextColumn()
			imgui.PushTextWrapPos()
			imutils.CopyableTextf("%v##shaderName", shaderName)
			imgui.TableNextColumn()
			imgui.PushTextWrapPos()
			imutils.CopyableTextf("%v##value", setting.Data)
			imgui.PopID()
		}
		imgui.EndTable()
	}
}

func tagSize(text string) (labelSize, tagSize imgui.Vec2) {
	style := imgui.CurrentStyle()
	labelSize = imgui.CalcTextSize(text)
	tagSize = imgui.InternalCalcItemSize(imgui.Vec2{}, labelSize.X+2.0*style.FramePadding().X, labelSize.Y+2.0*style.FramePadding().Y)
	return
}

func drawTag(text string) {
	imgui.BeginDisabled()
	imgui.Button(text)
	imgui.EndDisabled()
}

func (pv *EntityPreview) drawInfoMap(info entity.Info) {
	imutils.CopyableTextfV(imutils.CopyableTextOptions{
		TooltipHovered: "Entity Maps store shader settings that can be applied\n" +
			"per planet. They are used to change environment colors,\n" +
			"cloud settings, and other planetary effects",
	}, "Entity Map %v", fnt.I.Info)
	imgui.Text("Tags")
	totalSize := float32(0)
	for i, key := range info.ComponentThinHashes {
		text := fmt.Sprintf("%v", pv.lookupThinHash(key))
		_, size := tagSize(text)
		totalSize += size.X + 1.0
		fmt.Println(imgui.ContentRegionAvail())
		if i != 0 && totalSize < imgui.ContentRegionAvail().X {
			imgui.SameLine()
		}
		fmt.Println(imgui.CursorPos())
		if totalSize >= imgui.ContentRegionAvail().X {
			totalSize = 0
		}
		drawTag(text)
	}
	for _, component := range info.Components {
		pv.drawComponentMap(component)
	}
}

func (pv *EntityPreview) Draw(name string) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	for _, info := range pv.Entity.Infos {
		switch info.InfoType {
		case entity.InfoTypeMap1, entity.InfoTypeMap2:
			pv.drawInfoMap(info)
		case entity.InfoTypeString:
		}
	}
}

func (pv *EntityPreview) Load(data *entity.Entity, name stingray.FileID) {
	pv.Entity = data
	pv.name = name
}
