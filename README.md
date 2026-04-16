# mchat

`mchat` is a terminal email client that presents email threads like chats. It includes its own POP3 and SMTP implementations in Go.

## Run

Seed fake data into the app's normal database/config locations:

```bash
go run ./cmd/mchat-seed --force
```

Start the app:

```bash
go run ./cmd/mchat
```

[Watch the demo video](./mchat.webm)

## Use

- Open the chats list on startup.
- Press `Enter` to open a conversation.
- Press `Tab` to move into the message box.
- Press `c` to open config.
- Press `?` for help.
- Press `q` to quit.
