package data

import (
	"cmp"
	"fmt"
	"log"
	"net/smtp"
	"slices"

	"mchat/internal/auth_google"
	"mchat/internal/config"
	"mchat/internal/models"
	"mchat/pkg/oxsmtp"
	"mchat/pkg/pop3"

	"golang.org/x/oauth2"
)

type DataService struct {
	cfg *config.Config
}

func NewDataService() (*DataService, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &DataService{cfg: cfg}, nil
}

func (s *DataService) SendMessage(chat *models.Chat, msg string) error {
	token, err := s.GetActiveToken()
	if err != nil {
		return err
	}
	smtpAuth := oxsmtp.Auth{User: s.cfg.User, Token: token}

	header := make(map[string]string)
	header["From"] = s.cfg.User
	header["To"] = chat.Contact.Address
	header["Subject"] = "Notification from MChat"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg

	return smtp.SendMail("smtp.gmail.com:587", smtpAuth, s.cfg.User, []string{chat.Contact.Address}, []byte(msg))
}

func (s *DataService) GetChats() []*models.Chat {
	msgs := s.getMessages()
	chats := make(map[string]*models.Chat)

	for _, msg := range msgs {
		from := msg.Contact.Address
		chat, ok := chats[from]
		if ok {
			chat.Messages = append(chat.Messages, msg)
		} else {
			chats[from] = &models.Chat{
				Contact:  msg.Contact,
				Messages: []*models.Message{msg},
			}
		}
	}

	var ordChats []*models.Chat
	for _, chat := range chats {
		ordChats = append(ordChats, chat)
		slices.SortFunc(chat.Messages, func(a, b *models.Message) int {
			return cmp.Compare(a.Date, b.Date)
		})
	}

	return ordChats
}

func (s *DataService) IsConfigured() bool {
	return s.cfg.User != ""
}

func (s *DataService) SaveBasicConfig(user, pass string) {
	s.cfg = &config.Config{
		User:     user,
		Password: pass,
	}
	s.cfg.SaveConfig()
}

func (s *DataService) SaveGoogleConfig(user string, token *oauth2.Token) {
	s.cfg = &config.Config{
		User:  user,
		Token: *token,
	}
	s.cfg.SaveConfig()
}

func (s *DataService) GetActiveToken() (string, error) {
	authSvc := auth_google.NewGoogleAuthService()
	token, err := authSvc.GetActiveToken(&s.cfg.Token)
	if err != nil {
		return "", err
	}

	if s.cfg.Token.AccessToken != token.AccessToken {
		s.cfg.Token = *token
		s.cfg.SaveConfig()
	}
	return s.cfg.Token.AccessToken, nil
}

func (s *DataService) getMessages() []*models.Message {
	var p pop3.Pop3
	var conn *pop3.Connection
	var err error
	google_auth := s.cfg.Token.AccessToken != ""

	if google_auth {
		p = pop3.New("pop.gmail.com", "995")
		conn, err = p.Conn(true)
	} else {
		p = pop3.New("localhost", "1110")
		conn, err = p.Conn(false)
	}
	if err != nil {
		log.Fatal(err)
	}

	if google_auth {
		token, err := s.GetActiveToken()
		if err != nil {
			log.Fatal(err)
		}
		err = conn.XOAuth2(fmt.Sprintf("recent:%s", s.cfg.User), token)
	} else {
		err = conn.Auth(s.cfg.User, s.cfg.Password)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Quit()

	msginfos, err := conn.List()
	if err != nil {
		log.Fatal(err)
	}

	var messages []*models.Message
	for _, m := range msginfos {
		log.Printf("Retrieving msg %d of size %d ready\n", m.Id, m.Size)
		msg, err := conn.Retr(m.Id)
		if err != nil {
			log.Printf("error: %v", err)
		} else {
			messages = append(messages, processMessage(msg))
		}
	}

	return messages
}
