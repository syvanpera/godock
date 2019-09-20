package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/syvanpera/godock/flowdock"
)

var (
	flowIcon   = "  "
	unreadIcon = "  "
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
	v.Clear()
	v.Flows = flows
	for i, flow := range flows {
		v.SetCell(i, 0,
			tview.NewTableCell(flowIcon).
				SetTextColor(tcell.ColorWhite).
				SetExpansion(0).
				SetAlign(tview.AlignCenter))

		v.SetCell(i, 1,
			tview.NewTableCell(*flow.Name).
				SetTextColor(tcell.ColorWhite).
				SetExpansion(1).
				SetAlign(tview.AlignLeft))

		v.SetCell(i, 2,
			tview.NewTableCell(" ").
				SetTextColor(tcell.ColorYellow).
				SetExpansion(0).
				SetAlign(tview.AlignCenter))
	}
}

func (v *FlowsView) MarkFlowUnread(flowID string, unread bool) {
	for i, f := range v.Flows {
		if *f.Id == flowID {
			icon := unreadIcon
			if !unread {
				icon = " "
			}
			v.SetCell(i, 2,
				tview.NewTableCell(icon).
					SetTextColor(tcell.ColorYellow).
					SetExpansion(0).
					SetAlign(tview.AlignCenter))
		}
	}
}
