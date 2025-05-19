.PHONY: build clean package install test

# Binary name
BINARY_NAME=optimize-hpc-nic
# Version
VERSION=1.0.0
# Architecture
ARCH=amd64

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f bin/$(BINARY_NAME)
	@rm -rf build/

test:
	@echo "Running tests..."
	@go test -v ./...

package: build
	@echo "Creating Debian package..."
	@bash ./scripts/build-deb.sh

install: build
	@echo "Installing $(BINARY_NAME)..."
	@install -m 755 bin/$(BINARY_NAME) /usr/local/bin/
	@install -m 644 $(BINARY_NAME).service /lib/systemd/system/
	@systemctl daemon-reload
	@systemctl enable $(BINARY_NAME).service
	@systemctl start $(BINARY_NAME).service
	@echo "Installation complete!"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@systemctl stop $(BINARY_NAME).service || true
	@systemctl disable $(BINARY_NAME).service || true
	@rm -f /usr/bin/$(BINARY_NAME)
	@rm -f /lib/systemd/system/$(BINARY_NAME).service
	@systemctl daemon-reload
	@echo "Uninstallation complete!"