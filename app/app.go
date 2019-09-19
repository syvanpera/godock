package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/syvanpera/godock/app/ui"
	"github.com/syvanpera/godock/flowdock"
	"golang.org/x/oauth2"
)

type App struct {
	*tview.Application

	Config         *Config
	Token          *oauth2.Token
	TokenCache     TokenCache
	FlowdockClient *flowdock.Client

	flows      []flowdock.Flow
	flowLookup map[string]*flowdock.Flow
	users      []flowdock.User
	userLookup map[int]*flowdock.User

	activeFlow flowdock.Flow

	pages *tview.Pages
	views map[string]tview.Primitive
}

func NewApp(config *Config) *App {
	a := &App{
		Application: tview.NewApplication(),
		Config:      config,
		TokenCache:  CacheFile("token-cache.json"),
		pages:       tview.NewPages(),
	}

	a.views = map[string]tview.Primitive{
		"logo":     ui.NewLogoView(),
		"flows":    ui.NewFlowsView(),
		"messages": ui.NewMessagesView(),
		"input":    ui.NewInputView(),
		"debug":    ui.NewDebugView(),
	}

	return a
}

func (a *App) Init() {
	var err error
	a.FlowdockClient, err = a.initFlowdockClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Flowdock client initialization failed")
	}

	flows, _, _ := a.FlowdockClient.Flows.List(false, nil)
	a.flows = flows
	a.flowLookup = make(map[string]*flowdock.Flow, len(a.flows))
	for i, f := range a.flows {
		a.flowLookup[*f.Id] = &a.flows[i]
	}

	users, _, _ := a.FlowdockClient.Users.All()
	a.users = users
	a.userLookup = make(map[int]*flowdock.User, len(a.users))
	for i, u := range a.users {
		a.userLookup[*u.Id] = &a.users[i]
	}

	a.views["flows"].(*ui.FlowsView).Init(a.flows)
	a.views["input"].(*ui.InputView).Init()
	a.views["debug"].(*ui.DebugView).Init(func() {
		a.Draw()
	})

	a.views["flows"].(*ui.FlowsView).SetSelectedFunc(func(row, col int) {
		if row < len(a.flows) {
			a.changeActiveFlow(&a.flows[row])
		}
	})

	grid := tview.NewGrid().SetRows(8, 0, 2, 15).SetColumns(48, 0).SetBorders(false)
	grid.AddItem(a.views["logo"], 0, 0, 1, 1, 1, 1, false)
	grid.AddItem(a.views["flows"], 1, 0, 2, 1, 1, 1, true)
	grid.AddItem(a.views["messages"], 0, 1, 2, 1, 1, 1, false)
	grid.AddItem(a.views["input"], 2, 1, 1, 1, 1, 1, false)
	grid.AddItem(a.views["debug"], 3, 0, 1, 2, 1, 1, false)

	a.pages.AddPage("main", grid, true, true)
	a.SetRoot(a.pages, true)
}

func (a *App) Run() {
	// Set the first flow as active
	if len(a.flows) > 0 {
		a.changeActiveFlow(&a.flows[0])
	}

	// Start listening to all subscribed flows
	subscriptions := make([]string, 0, len(a.flows))
	for _, f := range a.flows {
		path := fmt.Sprintf("%v/%v", *f.Organization.ParameterizedName, *f.ParameterizedName)
		subscriptions = append(subscriptions, path)
	}

	msgChan, es, _ := a.FlowdockClient.Messages.Stream(subscriptions, a.Token.AccessToken)
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

func (a *App) initFlowdockClient() (*flowdock.Client, error) {
	log.Info().Msg("Initializing Flowdock Client")

	conf := &oauth2.Config{
		ClientID:     a.Config.Auth.ClientID,
		ClientSecret: a.Config.Auth.ClientSecret,
		Scopes:       []string{"flow", "private", "profile"},
		RedirectURL:  a.Config.Auth.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.Config.Auth.AuthURL,
			TokenURL: a.Config.Auth.TokenURL,
		},
	}

	ctx := context.Background()

	token, err := a.TokenCache.Token()
	if err != nil {
		log.Debug().Msg("No cached token found, need authorization")
		var code string

		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		fmt.Printf("Visit the URL below for the auth dialog:\n%v\n", url)

		fmt.Printf("And input the authorization code here: ")
		if _, err := fmt.Scan(&code); err != nil {
			return nil, err
		}

		token, err = conf.Exchange(ctx, code)
		if err != nil {
			return nil, err
		}

		a.TokenCache.PutToken(token)
	}

	log.Debug().Interface("Token", token).Msg("Using token")
	a.Token = token

	tc := conf.Client(ctx, token)

	return flowdock.NewClient(tc), nil
}

func (a *App) changeActiveFlow(flow *flowdock.Flow) {
	log.Debug().Interface("FLOW", flow).Msg("Changing active flow")
	a.activeFlow = *flow
	f := tview.Escape(fmt.Sprintf("[%s]", *flow.Name))
	prompt := fmt.Sprintf(" [yellow]  %s: ", f)
	a.views["input"].(*ui.InputView).SetLabel(prompt)

	opts := flowdock.MessagesListOptions{
		Event: "message",
		Limit: 30,
	}
	messages, _, err := a.FlowdockClient.Messages.List(*a.activeFlow.Organization.ParameterizedName, *a.activeFlow.ParameterizedName, &opts)
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
	_, _, err := a.FlowdockClient.Messages.Create(&opt)

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
