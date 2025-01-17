package gioui

import (
	"fmt"
	"image"
	"math"

	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/vsariola/sointu/tracker"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type UnitEditor struct {
	sliderList     *DragList
	searchList     *DragList
	Parameters     []*ParameterWidget
	DeleteUnitBtn  *ActionClickable
	CopyUnitBtn    *TipClickable
	ClearUnitBtn   *ActionClickable
	DisableUnitBtn *BoolClickable
	SelectTypeBtn  *widget.Clickable
	tag            bool
	caser          cases.Caser
}

func NewUnitEditor(m *tracker.Model) *UnitEditor {
	ret := &UnitEditor{
		DeleteUnitBtn:  NewActionClickable(m.DeleteUnit()),
		ClearUnitBtn:   NewActionClickable(m.ClearUnit()),
		DisableUnitBtn: NewBoolClickable(m.UnitDisabled().Bool()),
		CopyUnitBtn:    new(TipClickable),
		SelectTypeBtn:  new(widget.Clickable),
		sliderList:     NewDragList(m.Params().List(), layout.Vertical),
		searchList:     NewDragList(m.SearchResults().List(), layout.Vertical),
	}
	ret.caser = cases.Title(language.English)
	return ret
}

func (pe *UnitEditor) Layout(gtx C, t *Tracker) D {
	for _, e := range gtx.Events(&pe.tag) {
		switch e := e.(type) {
		case key.Event:
			if e.State == key.Press {
				pe.command(e, t)
			}
		}
	}
	defer op.Offset(image.Point{}).Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)).Push(gtx.Ops).Pop()
	key.InputOp{Tag: &pe.tag, Keys: "←|Shift-←|→|Shift-→|⎋"}.Add(gtx.Ops)

	editorFunc := pe.layoutSliders

	if t.UnitSearching().Value() || pe.sliderList.TrackerList.Count() == 0 {
		editorFunc = pe.layoutUnitTypeChooser
	}
	return Surface{Gray: 24, Focus: t.InstrumentEditor.wasFocused}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return editorFunc(gtx, t)
			}),
			layout.Rigid(func(gtx C) D {
				return pe.layoutFooter(gtx, t)
			}),
		)
	})
}

func (pe *UnitEditor) layoutSliders(gtx C, t *Tracker) D {
	numItems := pe.sliderList.TrackerList.Count()

	for len(pe.Parameters) < numItems {
		pe.Parameters = append(pe.Parameters, new(ParameterWidget))
	}

	index := 0
	t.Model.Params().Iterate(func(param tracker.Parameter) {
		pe.Parameters[index].Parameter = param
		index++
	})

	element := func(gtx C, index int) D {
		if index < 0 || index >= numItems {
			return D{}
		}
		paramStyle := t.ParamStyle(t.Theme, pe.Parameters[index])
		paramStyle.Focus = pe.sliderList.TrackerList.Selected() == index
		dims := paramStyle.Layout(gtx)
		return D{Size: image.Pt(gtx.Constraints.Max.X, dims.Size.Y)}
	}

	fdl := FilledDragList(t.Theme, pe.sliderList, element, nil)
	dims := fdl.Layout(gtx)
	gtx.Constraints = layout.Exact(dims.Size)
	fdl.LayoutScrollBar(gtx)
	return dims
}

func (pe *UnitEditor) layoutFooter(gtx C, t *Tracker) D {
	for pe.CopyUnitBtn.Clickable.Clicked() {
		if contents, ok := t.Units().List().CopyElements(); ok {
			clipboard.WriteOp{Text: string(contents)}.Add(gtx.Ops)
			t.Alerts().Add("Unit copied to clipboard", tracker.Info)
		}
	}
	copyUnitBtnStyle := TipIcon(t.Theme, pe.CopyUnitBtn, icons.ContentContentCopy, "Copy unit (Ctrl+C)")
	deleteUnitBtnStyle := ActionIcon(t.Theme, pe.DeleteUnitBtn, icons.ActionDelete, "Delete unit (Ctrl+Backspace)")
	disableUnitBtnStyle := ToggleIcon(t.Theme, pe.DisableUnitBtn, icons.AVVolumeUp, icons.AVVolumeOff, "Disable unit (Ctrl-D)", "Enable unit (Ctrl-D)")
	text := t.Units().SelectedType()
	if text == "" {
		text = "Choose unit type"
	} else {
		text = pe.caser.String(text)
	}
	hintText := Label(text, white, t.Theme.Shaper)
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(deleteUnitBtnStyle.Layout),
		layout.Rigid(copyUnitBtnStyle.Layout),
		layout.Rigid(disableUnitBtnStyle.Layout),
		layout.Rigid(func(gtx C) D {
			var dims D
			if t.Units().SelectedType() != "" {
				clearUnitBtnStyle := ActionIcon(t.Theme, pe.ClearUnitBtn, icons.ContentClear, "Clear unit")
				dims = clearUnitBtnStyle.Layout(gtx)
			}
			return D{Size: image.Pt(gtx.Dp(unit.Dp(48)), dims.Size.Y)}
		}),
		layout.Flexed(1, hintText),
	)
}

func (pe *UnitEditor) layoutUnitTypeChooser(gtx C, t *Tracker) D {
	var names [256]string
	index := 0
	t.Model.SearchResults().Iterate(func(item string) (ok bool) {
		names[index] = item
		index++
		return index <= 256
	})
	element := func(gtx C, i int) D {
		w := LabelStyle{Text: names[i], ShadeColor: black, Color: white, Font: labelDefaultFont, FontSize: unit.Sp(12), Shaper: t.Theme.Shaper}
		if i == pe.searchList.TrackerList.Selected() {
			for pe.SelectTypeBtn.Clicked() {
				t.Units().SetSelectedType(names[i])
			}
			return pe.SelectTypeBtn.Layout(gtx, w.Layout)
		}
		return w.Layout(gtx)
	}
	fdl := FilledDragList(t.Theme, pe.searchList, element, nil)
	dims := fdl.Layout(gtx)
	gtx.Constraints = layout.Exact(dims.Size)
	fdl.LayoutScrollBar(gtx)
	return dims
}

func (pe *UnitEditor) command(e key.Event, t *Tracker) {
	params := (*tracker.Params)(t.Model)
	switch e.State {
	case key.Press:
		switch e.Name {
		case key.NameLeftArrow:
			sel := params.SelectedItem()
			if sel == nil {
				return
			}
			i := (&tracker.Int{IntData: sel})
			if e.Modifiers.Contain(key.ModShift) {
				i.Set(i.Value() - sel.LargeStep())
			} else {
				i.Set(i.Value() - 1)
			}
		case key.NameRightArrow:
			sel := params.SelectedItem()
			if sel == nil {
				return
			}
			i := (&tracker.Int{IntData: sel})
			if e.Modifiers.Contain(key.ModShift) {
				i.Set(i.Value() + sel.LargeStep())
			} else {
				i.Set(i.Value() + 1)
			}
		case key.NameEscape:
			t.InstrumentEditor.unitDragList.Focus()
		}
	}
}

type ParameterWidget struct {
	floatWidget widget.Float
	boolWidget  widget.Bool
	instrBtn    widget.Clickable
	instrMenu   Menu
	unitBtn     widget.Clickable
	unitMenu    Menu
	Parameter   tracker.Parameter
}

type ParameterStyle struct {
	tracker *Tracker
	w       *ParameterWidget
	Theme   *material.Theme
	Focus   bool
}

func (t *Tracker) ParamStyle(th *material.Theme, paramWidget *ParameterWidget) ParameterStyle {
	return ParameterStyle{
		tracker: t, // TODO: we need this to pull the instrument names for ID style parameters, find out another way
		Theme:   th,
		w:       paramWidget,
	}
}

func (p ParameterStyle) Layout(gtx C) D {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(110))
			return layout.E.Layout(gtx, Label(p.w.Parameter.Name(), white, p.tracker.Theme.Shaper))
		}),
		layout.Rigid(func(gtx C) D {
			switch p.w.Parameter.Type() {
			case tracker.IntegerParameter:
				for _, e := range gtx.Events(&p.w.floatWidget) {
					if ev, ok := e.(pointer.Event); ok && ev.Type == pointer.Scroll {
						delta := math.Min(math.Max(float64(ev.Scroll.Y), -1), 1)
						tracker.Int{IntData: p.w.Parameter}.Add(-int(delta))
					}
				}
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(200))
				gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(40))
				if !p.w.floatWidget.Dragging() {
					p.w.floatWidget.Value = float32(p.w.Parameter.Value())
				}
				ra := p.w.Parameter.Range()
				sliderStyle := material.Slider(p.Theme, &p.w.floatWidget, float32(ra.Min), float32(ra.Max))
				sliderStyle.Color = p.Theme.Fg
				r := image.Rectangle{Max: gtx.Constraints.Min}
				area := clip.Rect(r).Push(gtx.Ops)
				if p.Focus {
					pointer.InputOp{Tag: &p.w.floatWidget, Types: pointer.Scroll, ScrollBounds: image.Rectangle{Min: image.Pt(0, -1e6), Max: image.Pt(0, 1e6)}}.Add(gtx.Ops)
				}
				dims := sliderStyle.Layout(gtx)
				area.Pop()
				tracker.Int{IntData: p.w.Parameter}.Set(int(p.w.floatWidget.Value + 0.5))
				return dims
			case tracker.BoolParameter:
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(60))
				gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(40))
				ra := p.w.Parameter.Range()
				p.w.boolWidget.Value = p.w.Parameter.Value() > ra.Min
				boolStyle := material.Switch(p.Theme, &p.w.boolWidget, "Toggle boolean parameter")
				boolStyle.Color.Disabled = p.Theme.Fg
				boolStyle.Color.Enabled = white
				dims := layout.Center.Layout(gtx, boolStyle.Layout)
				if p.w.boolWidget.Value {
					tracker.Int{IntData: p.w.Parameter}.Set(ra.Max)
				} else {
					tracker.Int{IntData: p.w.Parameter}.Set(ra.Min)
				}
				return dims
			case tracker.IDParameter:
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(200))
				gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(40))
				instrItems := make([]MenuItem, p.tracker.Instruments().Count())
				for i := range instrItems {
					i := i
					name, _, _ := p.tracker.Instruments().Item(i)
					instrItems[i].Text = name
					instrItems[i].IconBytes = icons.NavigationChevronRight
					instrItems[i].Doer = tracker.Allow(func() {
						if id, ok := p.tracker.Instruments().FirstID(i); ok {
							tracker.Int{IntData: p.w.Parameter}.Set(id)
						}
					})
				}
				var unitItems []MenuItem
				instrName := "<instr>"
				unitName := "<unit>"
				targetI, targetU, err := p.tracker.FindUnit(p.w.Parameter.Value())
				if err == nil {
					targetInstrument := p.tracker.Instrument(targetI)
					instrName = targetInstrument.Name
					units := targetInstrument.Units
					unitName = fmt.Sprintf("%v: %v", targetU, units[targetU].Type)
					unitItems = make([]MenuItem, len(units))
					for j, unit := range units {
						id := unit.ID
						unitItems[j].Text = fmt.Sprintf("%v: %v", j, unit.Type)
						unitItems[j].IconBytes = icons.NavigationChevronRight
						unitItems[j].Doer = tracker.Allow(func() {
							tracker.Int{IntData: p.w.Parameter}.Set(id)
						})
					}
				}
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(p.tracker.layoutMenu(instrName, &p.w.instrBtn, &p.w.instrMenu, unit.Dp(200),
						instrItems...,
					)),
					layout.Rigid(p.tracker.layoutMenu(unitName, &p.w.unitBtn, &p.w.unitMenu, unit.Dp(200),
						unitItems...,
					)),
				)
			}
			return D{}
		}),
		layout.Rigid(func(gtx C) D {
			if p.w.Parameter.Type() != tracker.IDParameter {
				return Label(p.w.Parameter.Hint(), white, p.tracker.Theme.Shaper)(gtx)
			}
			return D{}
		}),
	)
}
