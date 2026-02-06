package auth_google

import (
	"context"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var clientID, clientSecret string

type GoogleAuthService struct {
	cfg      oauth2.Config
	verifier string
}

func NewGoogleAuthService() *GoogleAuthService {
	c := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://mail.google.com/"},
		RedirectURL:  "http://localhost:8080",
		Endpoint:     google.Endpoint,
	}
	return &GoogleAuthService{cfg: c}
}

func (s *GoogleAuthService) GetActiveToken(t *oauth2.Token) (*oauth2.Token, error) {
	ts := s.cfg.TokenSource(context.Background(), t)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func (s *GoogleAuthService) GetGoogleUrl() string {
	s.verifier = oauth2.GenerateVerifier()
	return s.cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(s.verifier))
}

func (s *GoogleAuthService) WaitForAuthCode() string {
	codeChan := make(chan string)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		_, err := w.Write([]byte("Return to your terminal app"))
		if err != nil {
			log.Println(err)
		}
		codeChan <- code
	})
	server := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		err := server.ListenAndServe()
		log.Println(err)
	}()
	defer server.Close()

	code := <-codeChan

	return code
}

func (s *GoogleAuthService) ExchangeCode(code string) (*oauth2.Token, error) {
	token, err := s.cfg.Exchange(context.Background(), code, oauth2.VerifierOption(s.verifier))
	if err != nil {
		return nil, err
	}
	return token, nil
}
