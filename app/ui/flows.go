package ui

import (
	"math/rand"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/syvanpera/godock/flowdock"
)

type FlowsView struct {
	*tview.Table

	Flows []flowdock.Flow
}

func NewFlowsView() *FlowsView {
	t := tview.NewTable()
	t.
		SetSelectable(true, false).
		SetSelectedStyle(
			tcell.ColorBlack,
			tcell.ColorWhite,
			tcell.AttrNone,
		).
		SetBorder(true).
		SetTitle("   Flows ").
		SetTitleAlign(tview.AlignLeft)

	v := FlowsView{
		Table: t,
	}

	return &v
}

func (v *FlowsView) Init() {
}

func (v *FlowsView) SetFlows(flows []flowdock.Flow) {
	for i, flow := range flows {
		icon := "  "
		color := tcell.ColorWhite
		if rand.Intn(100) > 60 {
			// These flows are shown in the messages view
			color = tcell.ColorGreen
		}
		unread := "  "

		if rand.Intn(100) > 20 {
			unread = " "
		}

		v.SetCell(i, 0,
			tview.NewTableCell(icon).
				SetTextColor(color).
				SetExpansion(0).
				SetAlign(tview.AlignCenter))

		v.SetCell(i, 1,
			tview.NewTableCell(*flow.Name).
				SetTextColor(color).
				SetExpansion(1).
				SetAlign(tview.AlignLeft))

		v.SetCell(i, 2,
			tview.NewTableCell(unread).
				SetTextColor(tcell.ColorYellow).
				SetExpansion(0).
				SetAlign(tview.AlignCenter))
	}
}
