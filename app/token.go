package app

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
)

type TokenCache interface {
	Token() (*oauth2.Token, error)
	PutToken(*oauth2.Token) error
}

type CacheFile string

func (f CacheFile) Token() (*oauth2.Token, error) {
	file, err := os.Open(string(f))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	tok := &oauth2.Token{}
	if err := json.NewDecoder(file).Decode(tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func (f CacheFile) PutToken(tok *oauth2.Token) error {
	file, err := os.OpenFile(string(f), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(file).Encode(tok); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
