package ui

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type LogoView struct {
	*tview.Box
}

func NewLogoView() *LogoView {
	return &LogoView{
		Box: tview.NewBox(),
	}
}

func (r *LogoView) Draw(screen tcell.Screen) {
	// log.Debug().Msg("LogoView::Draw()")
	r.Box.Draw(screen)
	x, y, width, height := r.GetInnerRect()

	logo := `

 [green] █████╗  █████╗ [blue]█████╗  █████╗  █████╗██╗  ██╗
 [green]██╔═══╝ ██╔══██╗[blue]██╔═██╗██╔══██╗██╔═══╝██║ ██╔╝
 [green]██║ ███╗██║  ██║[blue]██║ ██║██║  ██║██║    █████╔╝
 [green]██║  ██║██║  ██║[blue]██║ ██║██║  ██║██║    ██╔═██╗
 [green]╚█████╔╝╚█████╔╝[blue]█████╔╝╚█████╔╝╚█████╗██║  ██╗
 [green] ╚════╝  ╚════╝ [blue]╚════╝  ╚════╝  ╚════╝╚═╝  ╚═╝
`

	for i, l := range strings.Split(logo, "\n") {
		if i >= height {
			break
		}
		tview.Print(screen, l, x, y+i-1, width, tview.AlignLeft, tcell.ColorYellow)
	}
}

// InputHandler returns the handler for this primitive.
func (r *LogoView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		log.Debug().Interface("EVENT", event).Msg("LogoView::InputHandler()")
	})
}
