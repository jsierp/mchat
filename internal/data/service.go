package data

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"sync"
	"time"

	"mchat/internal/auth_google"
	"mchat/internal/config"
	"mchat/internal/models"
	"mchat/internal/storage"
	"mchat/pkg/oxsmtp"
	"mchat/pkg/pop3"

	"golang.org/x/oauth2"
)

type DataService struct {
	db              *sql.DB
	cfg             *config.Config
	cfgMutex        sync.RWMutex
	msgChan         chan<- *models.Message
	existingMsgsIds map[string]struct{}
}

func NewDataService(msgChan chan<- *models.Message) (*DataService, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, err
	}

	svc := &DataService{db: db, cfg: cfg, msgChan: msgChan, existingMsgsIds: make(map[string]struct{})}
	svc.loadExistingMessages()

	go svc.startPolling()

	return svc, nil
}

func (s *DataService) loadExistingMessages() error {
	msgs, err := storage.GetMessages(s.db)
	if err != nil {
		return err
	}
	for _, m := range msgs {
		s.existingMsgsIds[m.Id] = struct{}{}
		s.msgChan <- m
	}
	return nil
}

func (s *DataService) startPolling() {
	for {
		log.Println("checking for updates..")
		s.fetchMessages()
		time.Sleep(time.Second * 15)
	}
}

func (s *DataService) SendMessage(chat *models.Chat, msg string) (*models.Message, error) {
	token, err := s.GetActiveToken()
	if err != nil {
		return nil, err
	}
	smtpAuth := oxsmtp.Auth{User: s.cfg.User, Token: token}
	msgId := fmt.Sprintf("<%d@mchat.mchat>", time.Now().UnixNano())
	date := time.Now()

	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", s.cfg.User)
	fmt.Fprintf(&b, "From: %s\r\n", s.cfg.User)
	fmt.Fprintf(&b, "To: %s\r\n", chat.Address)
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Subject: Notification from MChat\r\n")
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=\"utf-8\"\r\n")
	fmt.Fprintf(&b, "X-MCHAT-ID: %s\r\n", msgId)
	fmt.Fprintf(&b, "\r\n")
	fmt.Fprint(&b, msg)

	err = smtp.SendMail("smtp.gmail.com:587", smtpAuth, s.cfg.User, []string{chat.Address}, b.Bytes())
	if err != nil {
		return nil, err
	}

	m := &models.Message{
		Id:          msgId,
		From:        s.cfg.User,
		To:          chat.Address,
		Contact:     chat.Address,
		ChatAddress: chat.Address,
		Content:     msg,
		Date:        date.Format(time.DateTime),
	}
	err = storage.SaveMessage(s.db, m)
	return m, nil
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

func (s *DataService) fetchMessages() {
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
		err = conn.XOAuth2(s.cfg.User, token)
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

	for _, m := range msginfos {
		log.Printf("Retrieving msg %d of size %d ready\n", m.Id, m.Size)
		msg, err := conn.Retr(m.Id)
		if err != nil {
			log.Printf("error: %v", err)
		} else {
			m := s.processMessage(msg)
			if _, ok := s.existingMsgsIds[m.Id]; !ok {
				storage.SaveMessage(s.db, m)
				s.msgChan <- m
				s.existingMsgsIds[m.Id] = struct{}{}
			}
		}
	}
}
