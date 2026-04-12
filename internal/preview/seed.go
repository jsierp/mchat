package preview

import (
	"database/sql"
	"fmt"
	"mchat/internal/models"
	"path/filepath"
	"time"
)

const PreviewUser = "preview@mchat.test"

func SeedDB(db *sql.DB) (int, error) {
	if _, err := db.Exec(`DELETE FROM messages`); err != nil {
		return 0, err
	}

	messages := sampleMessages()
	for _, msg := range messages {
		_, err := db.Exec(
			`INSERT INTO messages (id, from_addr, to_addr, contact, chat_address, content, sent_date) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			msg.Id, msg.From, msg.To, msg.Contact, msg.ChatAddress, msg.Content, msg.Date,
		)
		if err != nil {
			return 0, err
		}
	}

	return len(messages), nil
}

func ConfigJSON() ([]byte, error) {
	return []byte("{}"), nil
}

func PreviewPaths(dir string) (dbPath, configPath string) {
	return filepath.Join(dir, "mchat.db"), filepath.Join(dir, "config.json")
}

func sampleMessages() []*models.Message {
	base := time.Date(2026, time.April, 9, 9, 0, 0, 0, time.UTC)
	type chatSeed struct {
		name     string
		address  string
		messages []string
	}

	chats := []chatSeed{
		{
			name:    "Alex Rivera",
			address: "alex@example.test",
			messages: []string{
				"Morning. Are we still on for the product preview later?",
				"Yes. I just need a lightweight demo dataset first.",
				"I can keep the copy generic so nothing looks real.",
				"Perfect. A few short threads are enough for screenshots.",
			},
		},
		{
			name:    "Jordan Lee",
			address: "jordan@example.test",
			messages: []string{
				"I left a couple of placeholder notes in the draft.",
				"Seen them. I'll trim it down before the review.",
				"Keep one longer message too, just to show wrapping in the UI preview.",
				"Good call. That makes the layout feel a lot more realistic.",
			},
		},
		{
			name:    "Sam Patel",
			address: "sam@example.test",
			messages: []string{
				"Do we need fake config as well?",
				"Only if the preview flow depends on it. Otherwise the sample DB is enough.",
				"Works for me. Let's keep the app behavior unchanged.",
				"Exactly. Seed data should stay isolated from real usage.",
			},
		},
	}

	var out []*models.Message
	for chatIdx, chat := range chats {
		for msgIdx, content := range chat.messages {
			timestamp := base.Add(time.Duration(chatIdx*4+msgIdx) * 37 * time.Minute)
			from := chat.address
			to := PreviewUser
			if msgIdx%2 == 1 {
				from = PreviewUser
				to = chat.address
			}

			out = append(out, &models.Message{
				Id:          fmt.Sprintf("<preview-%d-%d@mchat.test>", chatIdx+1, msgIdx+1),
				From:        from,
				To:          to,
				Contact:     chat.name,
				ChatAddress: chat.address,
				Content:     content,
				Date:        timestamp,
			})
		}
	}

	return out
}
