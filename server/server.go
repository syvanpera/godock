package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/syvanpera/godock/flowdock"
	"golang.org/x/oauth2"
)

type Server struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string

	Token      *oauth2.Token
	TokenCache TokenCache

	FlowdockClient *flowdock.Client
}

func (s *Server) Init() error {
	client, err := s.initFlowdockClient()
	if err != nil {
		log.Debug().Err(err).Msg("Flowdock client initialization failed")
		return errors.New("Flowdock client initialization failed")
	}
	s.FlowdockClient = client

	return nil
}

func (s *Server) initFlowdockClient() (*flowdock.Client, error) {
	log.Info().Msg("Initializing Flowdock Client")

	conf := &oauth2.Config{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		Scopes:       []string{"flow", "private", "profile"},
		RedirectURL:  s.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.AuthURL,
			TokenURL: s.TokenURL,
		},
	}

	ctx := context.Background()

	token, err := s.TokenCache.Token()
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

		s.TokenCache.PutToken(token)
	}

	log.Debug().Interface("Token", token).Msg("Using token")
	s.Token = token

	tc := conf.Client(ctx, token)

	return flowdock.NewClient(tc), nil
}
