package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/syvanpera/godock/flowdock"
)

type OrganizationsView struct {
	*tview.Table

	Organizations []flowdock.Organization
}

func NewOrganizationsView() *OrganizationsView {
	t := tview.NewTable()
	t.
		SetSelectable(true, false).
		SetSelectedStyle(
			tcell.ColorBlack,
			tcell.ColorWhite,
			tcell.AttrNone,
		).
		SetBorder(true).
		SetTitle("   Organizations ").
		SetTitleAlign(tview.AlignLeft)

	v := OrganizationsView{
		Table: t,
	}

	return &v
}

func (v *OrganizationsView) Init(organizations []flowdock.Organization) {
	for i, org := range organizations {
		v.SetCell(i, 0,
			tview.NewTableCell(*org.Name).
				SetTextColor(tcell.ColorWhite).
				SetExpansion(1).
				SetAlign(tview.AlignLeft))
	}
}
