BINARY     := moon-readwise
CLIENT     := moon-cli
GO         := go
GOFLAGS    := -ldflags="-s -w"
PREFIX     := /usr/local/bin
UNIT_DIR   := /etc/systemd/system
ENV_DIR    := /etc/moon-readwise
ENV_FILE   := moon-readwise.env

.PHONY: all build server client run test clean fmt lint vet install uninstall

all: server client

build: all

server:
	$(GO) build $(GOFLAGS) -o $(BINARY) ./cmd/server/main.go

client:
	$(GO) build $(GOFLAGS) -o $(CLIENT) ./cmd/client/main.go

run: server
	./$(BINARY) -debug

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

lint: fmt vet

vet:
	$(GO) vet ./...

clean:
	rm -f $(BINARY) $(CLIENT)

install: all
	install -m 755 $(BINARY) $(PREFIX)/$(BINARY)
	install -m 755 $(CLIENT) $(PREFIX)/$(CLIENT)
	install -m 644 deploy/moon-readwise.service $(UNIT_DIR)/moon-readwise.service
	mkdir -p -m 700 $(ENV_DIR)
	@test -f $(ENV_DIR)/$(ENV_FILE) || install -m 600 deploy/$(ENV_FILE) $(ENV_DIR)/$(ENV_FILE)
	systemctl daemon-reload

uninstall:
	rm -f $(PREFIX)/$(BINARY) $(PREFIX)/$(CLIENT)
	systemctl disable moon-readwise 2>/dev/null || true
	systemctl stop moon-readwise 2>/dev/null || true
	rm -f $(UNIT_DIR)/moon-readwise.service
	rm -rf $(ENV_DIR)
	systemctl daemon-reload
