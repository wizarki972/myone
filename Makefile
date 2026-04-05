MAKEFLAGS += --no-print-directory
export GO111MODULE=on
GOPROXY ?= https://proxy.golang.org,direct
export GOPROXY

BIN := myone
GO := go
DESTDIR :=
PREFIX := /usr/local

VERSION := 0.8.36
BUILD := beta

FLAGS ?= -trimpath -mod=readonly -modcacherw
LDFLAGS := -X "github.com/wizarki972/myone/internal/common.VERSION=${VERSION}" -X "github.com/wizarki972/myone/internal/common.BUILD=${BUILD}" -s -w
# For PIE Security features, I don't think this is needed... is it? 
EXTRA_FLAGS ?= -buildmode=pie

BASE_DIRECTORY = ~/.local/share/myone

.PHONY: all
all: install

.PHONY: dep
dep:
	@echo "CHECKING FOR DEPENDENCIES..."
	-@sudo pacman -S --needed --noconfirm starship go hyprland wireplumber blueman waybar rofi brightnessctl wiremix nwg-displays nwg-look nautilus wl-clipboard kitty swaync swayosd flameshot awww wlogout zsh > /dev/null 2>&1

.PHONY: build
build:
	@echo "BUILDING BINARY..."
	@mkdir -p ./build
	$(GO) build $(FLAGS) -ldflags '$(LDFLAGS)' -o ./build/$(BIN)
	chmod +x ./build/$(BIN)

.PHONY: install
install: build
	@echo "INSTALLING..."
	sudo install -Dm755 ./build/$(BIN) $(DESTDIR)$(PREFIX)/bin/$(BIN)

.PHONY: start
start:
	@echo "KILLING OLD PROCESSES..."
	@-killall -9 $(BIN)

	@echo "STARTING SYSTEM PROCESSES..."
	-/usr/local/bin/myone --battery-monitor > /dev/null 2>&1 & disown
	-/usr/local/bin/myone --monitor-daemon > /dev/null 2>&1 & disown

.PHONY: clean
clean:
	@echo "CLEANING UP..."
	$(GO) clean $(FLAGS) -i ./...
	-rm -rf ./build

.PHONY: full_install
full_install: dep install start clean