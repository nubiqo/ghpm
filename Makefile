# App Metadata
APP_NAME     = github-profile-manager
APP_ID       = com.ghpm.app
VERSION      = 0.0.11
LD_FLAGS     = -X github.com/huzaifanur/ghpm/pkg/version.Version=$(VERSION)

# Go
GOCMD        = go
GOBUILD      = $(GOCMD) build
GOCLEAN      = $(GOCMD) clean
GOTEST       = $(GOCMD) test
GOGET        = $(GOCMD) get

# Fyne
FYNE         = fyne

# Paths
BUILD_DIR    = build
DIST_DIR     = dist
DEB_DIR      = $(BUILD_DIR)/deb
BIN_PATH     = $(BUILD_DIR)/linux/$(APP_NAME)

.PHONY: all clean deps test build-linux build-darwin build-darwin-amd64 build-darwin-arm64 build package-deb package-tar package-darwin-amd64 package-darwin-arm64 package-linux package-darwin package release

all: deps test build

deps:
	go install fyne.io/tools/cmd/fyne@latest
	$(GOGET) -u ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR) $(DIST_DIR)

test:
	$(GOTEST) -v ./...

# Cross-compilation
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LD_FLAGS)" -o $(BIN_PATH) ./cmd/ghpm

build-darwin-amd64:
	@echo "Building for macOS Intel..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/darwin-amd64/$(APP_NAME) ./cmd/ghpm

build-darwin-arm64:
	@echo "Building for macOS Apple Silicon..."
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/darwin-arm64/$(APP_NAME) ./cmd/ghpm

build-darwin: build-darwin-amd64 build-darwin-arm64

build: build-linux

# Packaging
package-deb: build-linux
	@echo "Packaging for Ubuntu (.deb)..."
	@mkdir -p $(DIST_DIR)

	# .deb structure
	@mkdir -p $(DEB_DIR)/DEBIAN
	@mkdir -p $(DEB_DIR)/usr/bin
	@mkdir -p $(DEB_DIR)/usr/share/applications
	@mkdir -p $(DEB_DIR)/usr/share/icons/hicolor/256x256/apps

	# Copy files
	@cp $(BIN_PATH) $(DEB_DIR)/usr/bin/
	@cp logo.png $(DEB_DIR)/usr/share/icons/hicolor/256x256/apps/$(APP_NAME).png

	# Desktop entry
	@echo "[Desktop Entry]" > $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Type=Application" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Name=GitHub Profile Manager" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Comment=Manage multiple GitHub profiles with ease" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Icon=$(APP_NAME)" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Exec=/usr/bin/$(APP_NAME)" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Terminal=false" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop
	@echo "Categories=Development;Utility;" >> $(DEB_DIR)/usr/share/applications/$(APP_NAME).desktop

	# Control file (Multi-line Description, required fields)
	@echo "Package: $(APP_NAME)" > $(DEB_DIR)/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DEB_DIR)/DEBIAN/control
	@echo "Section: utils" >> $(DEB_DIR)/DEBIAN/control
	@echo "Priority: optional" >> $(DEB_DIR)/DEBIAN/control
	@echo "Architecture: amd64" >> $(DEB_DIR)/DEBIAN/control
	@echo "Maintainer: Your Name <your.email@example.com>" >> $(DEB_DIR)/DEBIAN/control
	@echo "Description: GitHub Profile Manager" >> $(DEB_DIR)/DEBIAN/control
	@echo " A desktop application to manage multiple GitHub profiles," >> $(DEB_DIR)/DEBIAN/control
	@echo " including git configuration and SSH keys." >> $(DEB_DIR)/DEBIAN/control

	# Build .deb
	@dpkg-deb --build $(DEB_DIR) $(DIST_DIR)/$(APP_NAME)_$(VERSION)_amd64.deb

package-tar: build-linux
	@echo "Creating Linux tarball..."
	@mkdir -p $(DIST_DIR)
	$(FYNE) package -os linux \
		--name "GitHub Profile Manager" \
		--app-id $(APP_ID) \
		--app-version $(VERSION) \
		--source-dir . \
		--executable $(BIN_PATH) \
		--icon logo.png \
		--release
	@mv *.tar.xz $(DIST_DIR)/GitHubProfileManager-$(VERSION)-linux-amd64.tar.xz 2>/dev/null || true

package-darwin-amd64: build-darwin-amd64
	@echo "Packaging macOS Intel with Fyne..."
	@mkdir -p $(DIST_DIR)
	$(FYNE) package -os darwin \
		--name "GitHub Profile Manager" \
		--app-id $(APP_ID) \
		--app-version $(VERSION) \
		--icon logo.png \
		--executable $(BUILD_DIR)/darwin-amd64/$(APP_NAME) \
		--release
	@mv *.dmg $(DIST_DIR)/GitHubProfileManager-$(VERSION)-darwin-amd64.dmg 2>/dev/null || true
	@zip -r $(DIST_DIR)/GitHubProfileManager-$(VERSION)-darwin-amd64.zip "GitHub Profile Manager.app" 2>/dev/null || true
	@rm -rf "GitHub Profile Manager.app" 2>/dev/null || true

package-darwin-arm64: build-darwin-arm64
	@echo "Packaging macOS ARM64 with Fyne..."
	@mkdir -p $(DIST_DIR)
	$(FYNE) package -os darwin \
		--name "GitHub Profile Manager" \
		--app-id $(APP_ID) \
		--app-version $(VERSION) \
		--icon logo.png \
		--executable $(BUILD_DIR)/darwin-arm64/$(APP_NAME) \
		--release
	@mv *.dmg $(DIST_DIR)/GitHubProfileManager-$(VERSION)-darwin-arm64.dmg 2>/dev/null || true
	@zip -r $(DIST_DIR)/GitHubProfileManager-$(VERSION)-darwin-arm64.zip "GitHub Profile Manager.app" 2>/dev/null || true
	@rm -rf "GitHub Profile Manager.app" 2>/dev/null || true

package-linux: package-deb package-tar

package-darwin: package-darwin-amd64 package-darwin-arm64

package: package-linux

release: clean package
	@echo "Release artifacts created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/
