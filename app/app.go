package app

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/syvanpera/godock/app/ui"
	"github.com/syvanpera/godock/flowdock"
	"github.com/syvanpera/godock/server"
)

type App struct {
	*tview.Application

	Server *server.Server

	organizations      []flowdock.Organization
	organizationLookup map[string]*flowdock.Organization
	flows              []flowdock.Flow
	flowLookup         map[string]*flowdock.Flow
	users              []flowdock.User
	userLookup         map[int]*flowdock.User

	activeOrganization flowdock.Organization
	activeFlow         flowdock.Flow

	pages *tview.Pages
	views map[string]tview.Primitive
}

func NewApp(server *server.Server) *App {
	a := &App{
		Application: tview.NewApplication(),
		Server:      server,
		pages:       tview.NewPages(),
	}

	a.views = map[string]tview.Primitive{
		"logo":          ui.NewLogoView(),
		"organizations": ui.NewOrganizationsView(),
		"flows":         ui.NewFlowsView(),
		"messages":      ui.NewMessagesView(),
		"input":         ui.NewInputView(),
		"debug":         ui.NewDebugView(),
	}

	return a
}

func (a *App) Init() {
	a.Server.Authenticate()

	organizations, _, _ := a.Server.FlowdockClient.Organizations.All()
	a.organizations = organizations
	a.organizationLookup = make(map[string]*flowdock.Organization, len(a.organizations))
	for _, o := range organizations {
		a.organizationLookup[*o.ParameterizedName] = &o
	}

	flows, _, _ := a.Server.FlowdockClient.Flows.List(false, nil)
	a.flows = make([]flowdock.Flow, len(a.flows))
	a.flowLookup = make(map[string]*flowdock.Flow, len(a.flows))
	for _, f := range flows {
		if *f.Open {
			a.flows = append(a.flows, f)
			a.flowLookup[*f.Id] = &f
		}
	}

	// Set the first organization as active
	if len(a.organizations) > 0 {
		a.changeActiveOrganization(&a.organizations[0])
	}

	users, _, _ := a.Server.FlowdockClient.Users.All()
	a.users = users
	a.userLookup = make(map[int]*flowdock.User, len(a.users))
	for i, u := range a.users {
		a.userLookup[*u.Id] = &a.users[i]
	}

	a.views["flows"].(*ui.FlowsView).Init()
	a.views["organizations"].(*ui.OrganizationsView).Init(a.organizations)
	a.views["input"].(*ui.InputView).Init()
	a.views["debug"].(*ui.DebugView).Init(func() {
		a.Draw()
	})

	a.views["organizations"].(*ui.OrganizationsView).SetSelectedFunc(func(row, col int) {
		if row < len(a.organizations) {
			a.changeActiveOrganization(&a.organizations[row])
		}
	})

	a.views["flows"].(*ui.FlowsView).SetSelectedFunc(func(row, col int) {
		if row < len(a.flows) {
			a.changeActiveFlow(&a.flows[row])
		}
	})

	grid := tview.NewGrid().SetRows(7, 4, 0, 1, 15).SetColumns(48, 0).SetBorders(false)
	grid.AddItem(a.views["logo"], 0, 0, 1, 1, 1, 1, false)
	grid.AddItem(a.views["organizations"], 1, 0, 1, 1, 1, 1, true)
	grid.AddItem(a.views["flows"], 2, 0, 1, 1, 1, 1, true)
	grid.AddItem(a.views["messages"], 0, 1, 3, 1, 1, 1, false)
	grid.AddItem(a.views["input"], 3, 0, 1, 2, 1, 1, false)
	grid.AddItem(a.views["debug"], 4, 0, 1, 2, 1, 1, false)

	a.pages.AddPage("main", grid, true, true)
	a.SetRoot(a.pages, true)
}

func (a *App) Run() {
	// Start listening to all subscribed flows
	subscriptions := make([]string, 0, len(a.flows))
	for _, f := range a.flows {
		path := fmt.Sprintf("%v/%v", *f.Organization.ParameterizedName, *f.ParameterizedName)
		subscriptions = append(subscriptions, path)
	}

	msgChan, es, _ := a.Server.FlowdockClient.Messages.Stream(subscriptions, a.Server.Token.AccessToken)
	defer es.Close()

	go func() {
		for msg := range msgChan {
			log.Debug().Interface("MSG", msg).Msg("Message from stream")
			if *msg.Event == "message" && *msg.FlowID == *a.activeFlow.Id {
				m := a.convertMessageToFlowMessage(msg)
				a.views["messages"].(*ui.MessagesView).NewMessage(m)
			}
		}
	}()

	a.views["input"].(*ui.InputView).SendMessage = a.sendMessage

	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyTab:
			if a.views["input"].(*ui.InputView).HasFocus() {
				a.SetFocus(a.views["flows"])
			} else {
				a.SetFocus(a.views["input"])
			}
		case tcell.KeyBacktab:
			a.SetFocus(a.views["flows"])
		case tcell.KeyRune:
			if event.Modifiers()&tcell.ModAlt > 0 {
				switch event.Rune() {
				case 'm':
					a.SetFocus(a.views["messages"])
				case 'o':
					a.SetFocus(a.views["organizations"])
				case 'f':
					a.SetFocus(a.views["flows"])
				case 'd':
					a.SetFocus(a.views["debug"])
				case 'i':
					a.SetFocus(a.views["input"])
				}
			} else if event.Key() == tcell.KeyRune {
				switch event.Rune() {
				case 'q':
					if !a.views["input"].(*ui.InputView).HasFocus() {
						a.Stop()
					}
				case '?':
					a.showHelp()
				}
			}
		}
		return event
	})

	if err := a.Application.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	a.Application.Stop()
}

func (a *App) changeActiveOrganization(org *flowdock.Organization) {
	if *org == a.activeOrganization {
		log.Debug().Msg("Selected organization already active")
		return
	}
	log.Debug().Interface("ORG", *org.ParameterizedName).Msg("Changing active organization")
	a.activeOrganization = *org

	flows := []flowdock.Flow{}
	for _, f := range a.flows {
		if *f.Organization.Id == *a.activeOrganization.Id {
			flows = append(flows, f)
		}
	}

	// Set the first flow of the organization as active
	if len(flows) > 0 {
		a.changeActiveFlow(&flows[0])
	}

	a.views["flows"].(*ui.FlowsView).SetFlows(flows)
}

func (a *App) changeActiveFlow(flow *flowdock.Flow) {
	if *flow == a.activeFlow {
		log.Debug().Msg("Selected flow already active")
		return
	}
	log.Debug().Interface("FLOW", flow).Msg("Changing active flow")
	a.activeFlow = *flow
	f := tview.Escape(fmt.Sprintf("[%s]", *flow.Name))
	prompt := fmt.Sprintf(" [yellow]  %s: ", f)
	a.views["input"].(*ui.InputView).SetLabel(prompt)

	opts := flowdock.MessagesListOptions{
		Event: "message",
		Limit: 30,
	}
	messages, _, err := a.Server.FlowdockClient.Messages.List(*a.activeFlow.Organization.ParameterizedName, *a.activeFlow.ParameterizedName, &opts)
	if err != nil {
		log.Error().Err(err).Interface("FLOW", a.activeFlow).Msg("Error while loading messages")
	}

	msgs := []ui.FlowMessage{}
	for _, m := range messages {
		msg := a.convertMessageToFlowMessage(m)
		msgs = append(msgs, msg)
	}

	a.views["messages"].(*ui.MessagesView).SetMessages(msgs)
}

func (a *App) sendMessage(msg string) error {
	log.Debug().Str("MSG", msg).Str("FLOW", *a.activeFlow.Name).Msg("Sending message")
	opt := flowdock.MessagesCreateOptions{
		FlowID:  *a.activeFlow.Id,
		Event:   "message",
		Content: msg,
		Tags:    []string{"#godock", "#golang"},
	}
	_, _, err := a.Server.FlowdockClient.Messages.Create(&opt)

	return err
}

func (a *App) showHelp() {
	helpText := `
<q> - quit             <alt-f> - Focus flows
<?> - show this help   <alt-i> - Focus input
`
	a.showModal(" HELP", helpText)
}

func (a *App) showModal(title, msg string) {
	log.Debug().Str("MSG", msg).Msg("showModal")
	m := tview.NewModal().
		AddButtons([]string{"Close"}).
		SetTextColor(tcell.ColorBlack).
		SetText(msg).
		SetDoneFunc(func(_ int, _ string) {
			a.dismissModal()
		})
	m.SetTitle(fmt.Sprintf(" %s ", title))
	a.pages.AddPage("modal", m, false, false)
	a.pages.ShowPage("modal")
}

func (a *App) dismissModal() {
	a.pages.RemovePage("modal")
	a.pages.SwitchToPage("main")
}

func (a *App) convertMessageToFlowMessage(msg flowdock.Message) ui.FlowMessage {
	userId, err := strconv.Atoi(*msg.UserID)
	var user *flowdock.User
	if err == nil {
		user = a.userLookup[userId]
	}
	nick := "John Doe"
	if user != nil {
		nick = *user.Nick
	}

	flow := a.flowLookup[*msg.FlowID]
	m := ui.FlowMessage{
		Nick:    nick,
		Flow:    *flow.Name,
		Sent:    msg.Sent.Time,
		Content: msg.Content().String(),
		Tags:    *msg.Tags,
	}

	return m
}
