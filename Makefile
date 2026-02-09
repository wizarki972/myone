BINARY_NAME = myone
INSTALL_DIR = /usr/local/bin
CACHE_DIR = ~/.cache/myone
BASE_DIR = ~/.local/share/myone
SCRIPTS_DIR = $(BASE_DIR)/bin
GO = go

all: build install

build:
	@echo "==> Building Binary..."
	@mkdir -p ./build
	$(GO) build -o ./build/$(BINARY_NAME)
	chmod +x ./build/$(BINARY_NAME)

build_cache:
	@echo "BUILDING..."
	mkdir -p $(CACHE_DIR)/build
	$(GO) build -o $(CACHE_DIR)/build/$(BINARY_NAME)
	chmod +x $(CACHE_DIR)/build/$(BINARY_NAME)

install: build_cache
	-killall -9 $(BINARY_NAME)

	@echo "INSTALLING..."
	sudo install -Dm755 $(CACHE_DIR)/build/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	mkdir -p $(SCRIPTS_DIR)
	sudo cp ./scripts/* $(SCRIPTS_DIR)
	sudo chmod +x $(SCRIPTS_DIR)

	@echo "STARTING SYSTEM PROCESSES..."
	-/usr/local/bin/myone --battery-monitor & disown
	-/usr/local/bin/myone --monitor-daemon & disown

start:
	@echo "STARTING SYSTEM PROCESSES..."
	-/usr/local/bin/myone --battery-monitor & disown
	-/usr/local/bin/myone --monitor-daemon & disown

clean:
	@echo "CLEANING UP..."
	-rm -rf $(CACHE_DIR)

run: build_cache
	@echo "RUNNING..."
	$(CACHE_DIR)/build/$(BINARY_NAME)

.PHONY: all build build_cache install clean run
