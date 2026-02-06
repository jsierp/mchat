BINARY_NAME=mchat

build:
	go build -ldflags="-X 'auth_google.clientID=$(GOOGLE_CLIENT_ID)' \
	                   -X 'auth_google.clientSecret=$(GOOGLE_CLIENT_SECRET)'" \
	         -o $(BINARY_NAME) ./cmd/mchat

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
