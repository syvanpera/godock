package ui

import (
	"github.com/derailed/tview"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DebugView struct {
	*tview.TextView
}

func NewDebugView() *DebugView {
	text := tview.NewTextView()
	text.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("   Debug ").
		SetTitleAlign(tview.AlignLeft)

	v := DebugView{
		TextView: text,
	}

	return &v
}

func (v *DebugView) Init(changedFn func()) {
	v.SetChangedFunc(changedFn)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: tview.ANSIWriter(v, "black", "white"), TimeFormat: "15:04:05"})
}
