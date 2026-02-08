package data

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"mchat/internal/auth_google"
	"mchat/internal/config"
	"mchat/internal/models"
	"mchat/internal/storage"
	"mchat/pkg/oxsmtp"
	"mchat/pkg/pop3"

	"golang.org/x/oauth2"
)

const mChatIdHeader = "X-MChat-Id"

type DataService struct {
	db              *sql.DB
	cfg             *config.Config
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
	err = svc.loadExistingMessages()
	if err != nil {
		log.Println(err)
	}

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
		if s.cfg.User != "" {
			log.Println("checking for updates..")
			err := s.fetchMessages()
			if err != nil {
				log.Println("error while fetching messages", err)
			}
		} else {
			log.Println("app not configured yet. skiping fetch")
		}
		time.Sleep(time.Second * 15)
	}
}

func (s *DataService) SendMessage(m *models.Message) error {
	// Sets From and Id fields - without err - and sends the message
	m.Id = fmt.Sprintf("<%d@mchat.mchat>", time.Now().UnixNano())
	m.From = s.cfg.User

	token, err := s.GetActiveToken()
	if err != nil {
		return err
	}
	smtpAuth := oxsmtp.Auth{User: s.cfg.User, Token: token}

	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", m.From)
	fmt.Fprintf(&b, "To: %s\r\n", m.ChatAddress)
	fmt.Fprintf(&b, "Date: %s\r\n", m.Date.Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Subject: Notification from MChat\r\n")
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=\"utf-8\"\r\n")
	fmt.Fprintf(&b, "%s: %s\r\n", mChatIdHeader, m.Id)
	fmt.Fprintf(&b, "\r\n")
	fmt.Fprint(&b, m.Content)

	err = smtp.SendMail("smtp.gmail.com:587", smtpAuth, s.cfg.User, []string{m.ChatAddress}, b.Bytes())
	if err != nil {
		return err
	}

	err = storage.SaveMessage(s.db, m)
	if err != nil {
		log.Println("error when saving the message", err)
	}
	return nil
}

func (s *DataService) SaveBasicConfig(user, pass string) {
	s.cfg = &config.Config{
		User:     user,
		Password: pass,
	}
	err := s.cfg.SaveConfig()
	if err != nil {
		log.Println(err)
	}
}

func (s *DataService) SaveGoogleConfig(user string, token *oauth2.Token) {
	s.cfg = &config.Config{
		User:  user,
		Token: *token,
	}
	err := s.cfg.SaveConfig()
	if err != nil {
		log.Println(err)
	}
}

func (s *DataService) GetActiveToken() (string, error) {
	authSvc := auth_google.NewGoogleAuthService()
	token, err := authSvc.GetActiveToken(&s.cfg.Token)
	if err != nil {
		return "", err
	}

	if s.cfg.Token.AccessToken != token.AccessToken {
		s.cfg.Token = *token
		err = s.cfg.SaveConfig()
		if err != nil {
			log.Println(err)
		}
	}
	return s.cfg.Token.AccessToken, nil
}

func (s *DataService) fetchMessages() error {
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
		return err
	}

	if google_auth {
		var token string
		token, err = s.GetActiveToken()
		if err != nil {
			log.Fatal(err)
		}
		err = conn.XOAuth2(s.cfg.User, token)
	} else {
		err = conn.Auth(s.cfg.User, s.cfg.Password)
	}

	if err != nil {
		return err
	}

	msginfos, err := conn.List()
	if err != nil {
		return err
	}

	for _, m := range msginfos {
		log.Printf("Retrieving msg %d of size %d ready\n", m.Id, m.Size)
		msg, err := conn.Retr(m.Id)
		if err != nil {
			log.Printf("error: %v", err)
		} else {
			m := s.processMessage(msg)
			if _, ok := s.existingMsgsIds[m.Id]; !ok {
				err := storage.SaveMessage(s.db, m)
				if err != nil {
					log.Println(err)
				}
				s.msgChan <- m
				s.existingMsgsIds[m.Id] = struct{}{}
			}
		}
	}

	err = conn.Quit()
	return err
}
