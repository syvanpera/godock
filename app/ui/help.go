package ui

import (
	"github.com/derailed/tview"
)

type HelpView struct {
	*tview.TextView
}

func NewHelpView() *HelpView {
	text := tview.NewTextView()
	text.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("   Help ").
		SetTitleAlign(tview.AlignLeft)

	v := HelpView{
		TextView: text,
	}

	return &v
}
