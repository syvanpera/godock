package ui

import (
	"html/template"
	"time"

	"github.com/rivo/tview"
)

const (
	messageTemplate = `[gray]{{.Sent.Format "02.01.06 15:04:05"}} [darkcyan][{{.Flow}}] - [yellow]{{.Nick}}:
{{if .Tags}}[blue]{{.Tags}}
{{end}}[white]{{.Content}}

`
)

type FlowMessage struct {
	Nick    string
	Flow    string
	Sent    time.Time
	Content string
	Tags    []string
}

type MessagesView struct {
	*tview.TextView

	tpl *template.Template
}

func NewMessagesView() *MessagesView {
	view := tview.NewTextView()
	view.
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("   Messages ").
		SetTitleAlign(tview.AlignLeft)

	v := MessagesView{
		TextView: view,
	}

	t, err := template.New("msg").Parse(messageTemplate)
	if err != nil {
		panic(err)
	}
	v.tpl = t

	return &v
}

func (v *MessagesView) SetMessages(msgs []FlowMessage) {
	v.Clear()
	for _, m := range msgs {
		err := v.tpl.Execute(v, m)
		if err != nil {
			panic(err)
		}
	}
	v.ScrollToEnd()
}

func (v *MessagesView) NewMessage(m FlowMessage) {
	// tags := strings.Join(m.Tags, ", ")
	// v.Write([]byte(fmt.Sprintf("[gray]%s [blue][%s] [yellow]%s:\n[white]%s\n[darkcyan]%s\n", m.Sent.Format("02.01.06 15:04:05"), m.Flow, m.Nick, m.Content, tags)))
	err := v.tpl.Execute(v, m)
	if err != nil {
		panic(err)
	}
}
