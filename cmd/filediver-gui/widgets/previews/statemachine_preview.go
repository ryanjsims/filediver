package previews

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl32"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/textutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/state_machine"
)

type DisplayMode int32

const (
	AnimationList DisplayMode = iota
	States
	AnimationSelector
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=DisplayMode

type layerStateEntry struct {
	idx             uint32
	triggeringEvent stingray.ThinHash
	// How long should the highlight last when visible (in seconds)
	highlight float32
}

type StateMachinePreview struct {
	*state_machine.StateMachine
	queryBuf        string
	shownFiles      []stingray.Hash
	allFiles        []stingray.Hash
	availableEvents []stingray.ThinHash
	layerStates     []layerStateEntry
	displayMode     DisplayMode
	selection       stingray.Hash

	lookupHash     func(stingray.Hash) string
	lookupThinHash func(stingray.ThinHash) string
}

func NewStateMachinePreview(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) *StateMachinePreview {
	return &StateMachinePreview{
		queryBuf:    "",
		displayMode: AnimationList,

		lookupHash:     lookupHash,
		lookupThinHash: lookupThinHash,
	}
}

func (pv *StateMachinePreview) updateShownFiles() {
	queryMatcher := textutils.NewQuery(pv.queryBuf)
	pv.shownFiles = pv.shownFiles[:0]
	for _, item := range pv.allFiles {
		if queryMatcher.MatchesAny(pv.lookupHash(item), item.String()) {
			pv.shownFiles = append(pv.shownFiles, item)
		}
	}

	slices.SortFunc(pv.shownFiles, func(a, b stingray.Hash) int {
		return strings.Compare(pv.lookupHash(a), pv.lookupHash(b))
	})
}

func (pv *StateMachinePreview) drawAnimationsTable(idx int, state state_machine.State) {
	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsRowBg
	imgui.PushStyleColorVec4(imgui.ColTableRowBg, *imgui.StyleColorVec4(imgui.ColWindowBg))
	altBg := imgui.NewVec4(0.1, 0.1, 0.1, 0.75)
	imgui.PushStyleColorVec4(imgui.ColTableRowBgAlt, altBg)
	defer imgui.PopStyleColorV(2)
	if imgui.BeginTableV(fmt.Sprintf("##Animation subtable %v", idx), 2, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Weight", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("Animation File", imgui.TableColumnFlagsWidthStretch, 3, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		maxWeight := float32(1.0)
		if len(state.AnimationWeights) > 0 {
			maxWeight = state.AnimationWeights[len(state.AnimationWeights)-1]
		}
		for idx, animation := range state.AnimationHashes {
			animationWeight := float32(1.0)
			if idx < len(state.AnimationWeights) {
				animationWeight = state.AnimationWeights[idx]
				if idx > 0 {
					animationWeight -= state.AnimationWeights[idx-1]
				}
			}
			animationWeight /= maxWeight
			imgui.TableNextColumn()
			imutils.Textf("%4.1f", 100*animationWeight)
			imgui.TableNextColumn()
			imgui.PushTextWrapPos()
			widgets.GamefileLinkTextF(
				stingray.NewFileID(animation, stingray.Sum("animation")),
				"%v.animation",
				pv.lookupHash(animation),
			)
		}
		if len(state.AnimationHashes) == 0 {
			imgui.TableNextColumn()
			imgui.Text("N/A")
			imgui.TableNextColumn()
			imgui.Text("<none>")
		}
		imgui.EndTable()
	}
}

const highlightDuration = float32(2.0) // seconds

func lerp(a, b mgl32.Vec4, x float32) mgl32.Vec4 {
	return a.Mul(1.0 - x).Add(b.Mul(x))
}

func fromImgui(a imgui.Vec4) mgl32.Vec4 {
	return mgl32.Vec4{
		a.X,
		a.Y,
		a.Z,
		a.W,
	}
}

func toImgui(a mgl32.Vec4) imgui.Vec4 {
	return imgui.NewVec4(a.Elem())
}

func (pv *StateMachinePreview) drawLayer(idx int, layer state_machine.Layer) {
	state := layer.States[pv.layerStates[idx].idx]
	var childFlags = imgui.ChildFlagsFrameStyle | imgui.ChildFlagsAutoResizeY

	colorActive := fromImgui(*imgui.StyleColorVec4(imgui.ColFrameBgActive))
	colorInactive := fromImgui(*imgui.StyleColorVec4(imgui.ColFrameBg))
	result := *imgui.StyleColorVec4(imgui.ColFrameBg)
	if pv.layerStates[idx].highlight > 0.0 {
		factor := (highlightDuration - pv.layerStates[idx].highlight) / highlightDuration
		result = toImgui(imutils.AnimateToFrom(
			colorInactive, colorActive, factor,
			lerp,
			func(f float32) float32 {
				// Ease out quintic
				return 1.0 - float32(math.Pow(1.0-float64(f), 5.0))
			},
		))
	}
	imgui.PushStyleColorVec4(imgui.ColFrameBg, result)
	defer imgui.PopStyleColor()
	if imgui.BeginChildStrV(fmt.Sprintf("%v %v", state.Name.String(), idx), imgui.NewVec2(0.0, 0.0), childFlags, imgui.WindowFlagsNone) {
		imgui.SeparatorText(fmt.Sprintf("State: %v", pv.lookupHash(state.Name)))
		if pv.layerStates[idx].triggeringEvent.Value != 0x0 {
			imgui.Text("Triggered by")
			imgui.SameLine()
			imutils.CopyableTextf("%v", pv.lookupThinHash(pv.layerStates[idx].triggeringEvent))
		}
		imgui.BeginDisabled()
		imgui.Checkbox("Additive", &state.Additive)
		imgui.SameLine()
		imgui.Checkbox("Loop", &state.Loop)
		imgui.EndDisabled()

		imgui.Text("Animations")
		pv.drawAnimationsTable(idx, state)

		imgui.Text("End Event")
		imgui.SameLine()
		if state.EmitEndEvent.Value != 0x0 {
			imutils.CopyableTextf("%v", pv.lookupThinHash(state.EmitEndEvent))
		} else {
			imgui.Text("<none>")
		}

		imgui.Text("Group ID")
		imgui.SameLine()
		imutils.CopyableTextf("%v", pv.lookupThinHash(state.GroupID))

		if state.RagdollName.Value != 0x0 {
			imgui.Text("Ragdoll Name")
			imgui.SameLine()
			imutils.CopyableTextf("%v", pv.lookupThinHash(state.RagdollName))
		}
	}
	imgui.EndChild()
	if imgui.IsItemVisible() && pv.layerStates[idx].highlight > 0.0 {
		pv.layerStates[idx].highlight -= imgui.CurrentIO().DeltaTime()
	}
}

func (pv *StateMachinePreview) drawStates() {
	if imgui.Button("Reset") {
		pv.Reset()
	}
	imgui.SameLine()
	if imgui.BeginCombo("Send Animation Event", "Choose event...") {
		for _, event := range pv.availableEvents {
			if imgui.SelectableBool(pv.lookupThinHash(event)) {
				pv.sendEvent(event)
				pv.updateAvailableEvents()
				break
			}
		}
		imgui.EndCombo()
	}
	if imgui.BeginChildStr("##Layers Child") {
		for idx, layer := range pv.Layers {
			pv.drawLayer(idx, layer)
		}
	}
	imgui.EndChild()
}

func (pv *StateMachinePreview) drawList() {
	imutils.Textf("Animation Count: %v (%v/%v displayed)", len(pv.allFiles), len(pv.shownFiles), len(pv.allFiles))
	imgui.SetNextItemWidth(imgui.ContentRegionAvail().X)
	if imgui.InputTextWithHint("##Search Files", fnt.I.Search+" Filter by hash or string...", &pv.queryBuf, 0, nil) {
		pv.updateShownFiles()
	}

	imgui.Text("Files")
	if imgui.BeginListBoxV("##Files Preview", imgui.NewVec2(-1.0, -1.0)) {
		for _, id := range pv.shownFiles {
			isSelected := id == pv.selection
			if imgui.SelectableBoolV(pv.lookupHash(id), isSelected, imgui.SelectableFlagsNone, imgui.NewVec2(0.0, 0.0)) {
				pv.selection = id
			}
			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}
		imgui.EndListBox()
	}
}

func (pv *StateMachinePreview) Draw(name string) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	if pv.displayMode != AnimationSelector {
		imgui.SeparatorText("Display Mode")
		imgui.RadioButtonIntPtr("Animation file list", (*int32)(&pv.displayMode), int32(AnimationList))
		imgui.SameLine()
		imgui.RadioButtonIntPtr("State simulation", (*int32)(&pv.displayMode), int32(States))
	}

	if pv.StateMachine == nil {
		imgui.Text("<no state_machine>")
		return
	}

	switch pv.displayMode {
	case AnimationList, AnimationSelector:
		pv.drawList()
	case States:
		pv.drawStates()
	}
}

func (pv *StateMachinePreview) sendEvent(event stingray.ThinHash) {
	for idx, layer := range pv.Layers {
		state := layer.States[pv.layerStates[idx].idx]
		if link, contains := state.StateTransitions[event]; contains {
			pv.layerStates[idx].idx = link.Index
			pv.layerStates[idx].triggeringEvent = event
			pv.layerStates[idx].highlight = highlightDuration
		}
	}
}

func (pv *StateMachinePreview) updateAvailableEvents() {
	pv.availableEvents = pv.availableEvents[:0]
	for idx, layer := range pv.Layers {
		for event := range layer.States[pv.layerStates[idx].idx].StateTransitions {
			pv.availableEvents = append(pv.availableEvents, event)
		}
	}

	pv.availableEvents = slices.Compact(slices.SortedFunc(slices.Values(pv.availableEvents), func(a, b stingray.ThinHash) int {
		return strings.Compare(pv.lookupThinHash(a), pv.lookupThinHash(b))
	}))
}

func (pv *StateMachinePreview) Reset() {
	for idx, layer := range pv.Layers {
		pv.layerStates[idx] = layerStateEntry{
			idx:             layer.DefaultState,
			triggeringEvent: stingray.ThinHash{Value: 0x0},
		}
	}
	pv.updateAvailableEvents()
}

func (pv *StateMachinePreview) Clear() {
	pv.StateMachine = nil
	pv.allFiles = make([]stingray.Hash, 0)
	pv.shownFiles = make([]stingray.Hash, 0)
	pv.availableEvents = make([]stingray.ThinHash, 0)
	pv.selection = stingray.Hash{}
}

func (pv *StateMachinePreview) Load(data *state_machine.StateMachine, name stingray.FileID) {
	pv.Clear()
	pv.StateMachine = data
	pv.layerStates = make([]layerStateEntry, len(pv.Layers))
	pv.Reset()
	for _, layer := range pv.Layers {
		for _, state := range layer.States {
			for _, animation := range state.AnimationHashes {
				pv.allFiles = append(pv.allFiles, animation)
			}
		}
	}
	pv.allFiles = slices.Compact(slices.SortedFunc(slices.Values(pv.allFiles), stingray.Hash.Cmp))
	pv.updateShownFiles()
}

func (pv *StateMachinePreview) SetMode(mode DisplayMode) {
	pv.displayMode = mode
}

func (pv *StateMachinePreview) GetSelection() stingray.Hash {
	return pv.selection
}
