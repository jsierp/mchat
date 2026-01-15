package config

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuthService struct {
	cfg oauth2.Config
}

func NewGoogleAuthService() *GoogleAuthService {
	c := oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Scopes:       []string{"https://mail.google.com/"},
		RedirectURL:  "http://localhost:8080",
		Endpoint:     google.Endpoint,
	}
	return &GoogleAuthService{cfg: c}
}

func (svc *GoogleAuthService) GetGoogleUrl() string {
	return svc.cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (svc *GoogleAuthService) WaitForAuthCode() string {
	codeChan := make(chan string)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		w.Write([]byte("Return to your terminal app"))
		codeChan <- code
	})
	server := &http.Server{Addr: ":8080", Handler: mux}
	go server.ListenAndServe()

	code := <-codeChan

	server.Close()

	return code
}

func (svc *GoogleAuthService) ExchangeCode(code string) (*Config, error) {
	token, err := svc.cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	return &Config{
		Google:       true,
		Login:        "recent:mchatgolang@gmail.com",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}, nil
}
