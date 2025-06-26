# App Metadata
APP_NAME     = github-profile-manager
APP_ID       = com.ghpm.app
VERSION      = 0.0.2

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

.PHONY: all clean deps test build-linux package-deb release

all: deps test build-linux

deps:
	go install fyne.io/tools/cmd/fyne@latest
	$(GOGET) -u ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR) $(DIST_DIR)

test:
	$(GOTEST) -v ./...

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BIN_PATH) .

package-deb: build-linux
	@echo "Packaging for Ubuntu (.deb)..."
	@mkdir -p $(DIST_DIR)

	# Fyne package
	$(FYNE) package -os linux \
		--name "GitHub Profile Manager" \
		--app-id $(APP_ID) \
		--app-version $(VERSION) \
		--source-dir . \
		--executable $(BIN_PATH) \
		--icon logo.png \
		--release

	@mv *.tar.xz $(DIST_DIR)/GitHubProfileManager-$(VERSION)-linux-amd64.tar.xz 2>/dev/null || true

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

release: clean package-deb
	@echo "Release artifacts created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/