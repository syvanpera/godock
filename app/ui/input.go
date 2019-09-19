package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type InputView struct {
	*tview.InputField

	FlowName    string
	SendMessage func(s string) error
}

func NewInputView() *InputView {
	input := tview.NewInputField()
	input.
		SetFieldBackgroundColor(tcell.ColorGrey).
		SetFieldTextColor(tcell.ColorWhite)

	return &InputView{
		InputField: input,
	}
}

func (v *InputView) Init() {
	v.InputField.SetFinishedFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			msg := v.InputField.GetText()
			log.Debug().Str("MSG", msg).Msg("Input finished")
			if v.SendMessage != nil {
				if err := v.SendMessage(msg); err == nil {
					v.InputField.SetText("")
				}
			}
		}
	})
}

// func (r *InputView) Draw(screen tcell.Screen) {
// 	f := tview.Escape(fmt.Sprintf("[%s]", r.FlowName))
// 	prompt := fmt.Sprintf(" [yellow] %s: ", f)
// 	r.SetLabel(prompt)
// 	r.InputField.Draw(screen)
// 	// x, y, width, height := r.GetInnerRect()

// 	// tview.Print(screen, prompt, x, y+height/2, width, tview.AlignLeft, tcell.ColorYellow)

// 	// if r.HasFocus() {
// 	// 	screen.ShowCursor(x+tview.TaggedStringWidth(prompt), y+height/2)
// 	// }
// }

// InputHandler returns the handler for this primitive.
// func (r *InputView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
// 	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
// 		log.Debug().Interface("EVENT", event).Msg("InputView::InputHandler()")
// 	})
// }
