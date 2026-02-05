package storage

import (
	"database/sql"
	"mchat/internal/models"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	_ "modernc.org/sqlite"
)

func getPath() (string, error) {
	path := filepath.Join(xdg.DataHome, "mchat", "mchat.db")
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	return path, err
}

func initDb(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT,
		from_addr TEXT,
		to_addr TEXT,
		contact TEXT,
		chat_address TEXT,
		content TEXT,
		sent_date DATETIME
    );
	`
	_, err := db.Exec(schema)
	return err
}

func GetDB() (*sql.DB, error) {
	path, err := getPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	err = initDb(db)
	return db, err
}

func GetMessages(db *sql.DB) ([]*models.Message, error) {
	rows, err := db.Query(`SELECT * FROM messages`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []*models.Message

	for rows.Next() {
		var msg models.Message
		err = rows.Scan(&msg.Id, &msg.From, &msg.To, &msg.Contact, &msg.ChatAddress, &msg.Content, &msg.Date)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}

func SaveMessage(db *sql.DB, msg *models.Message) error {
	_, err := db.Exec(
		`INSERT INTO messages (id, from_addr, to_addr, contact, chat_address, content, sent_date) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.Id, msg.From, msg.To, msg.Contact, msg.ChatAddress, msg.Content, msg.Date,
	)
	return err
}
