# App Metadata
APP_NAME     = github-profile-manager
APP_ID       = com.ghpm.app
VERSION      = 0.0.10

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

.PHONY: all clean deps test build-linux build-darwin build-darwin-amd64 build-darwin-arm64 build package-deb package-tar package-dmg package-dmg-amd64 package-dmg-arm64 package-zip-amd64 package-zip-arm64 package-linux package-darwin package release

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
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BIN_PATH) ./cmd/ghpm

build-darwin-amd64:
	@echo "Building for macOS Intel..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/darwin-amd64/$(APP_NAME) ./cmd/ghpm

build-darwin-arm64:
	@echo "Building for macOS Apple Silicon..."
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/darwin-arm64/$(APP_NAME) ./cmd/ghpm

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

package-dmg-amd64: 
	@echo "Creating macOS DMG for Intel..."
	@if [ -d "$(BUILD_DIR)/darwin-amd64" ]; then \
		mkdir -p $(DIST_DIR); \
		echo "Contents of $(BUILD_DIR)/darwin-amd64:"; \
		ls -la $(BUILD_DIR)/darwin-amd64/; \
		cd $(BUILD_DIR)/darwin-amd64 && \
		hdiutil create -volname "$(APP_NAME)" -srcfolder . -ov -format UDZO ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.dmg; \
	else \
		echo "macOS Intel build not found. Skipping DMG creation."; \
	fi

package-dmg-arm64:
	@echo "Creating macOS DMG for Apple Silicon..."
	@if [ -d "$(BUILD_DIR)/darwin-arm64" ]; then \
		mkdir -p $(DIST_DIR); \
		echo "Contents of $(BUILD_DIR)/darwin-arm64:"; \
		ls -la $(BUILD_DIR)/darwin-arm64/; \
		cd $(BUILD_DIR)/darwin-arm64 && \
		hdiutil create -volname "$(APP_NAME)" -srcfolder . -ov -format UDZO ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.dmg; \
	else \
		echo "macOS ARM64 build not found. Skipping DMG creation."; \
	fi

package-dmg: package-dmg-amd64 package-dmg-arm64

package-zip-amd64:
	@echo "Creating macOS ZIP for Intel..."
	@mkdir -p $(DIST_DIR)
	@if [ -d "$(BUILD_DIR)/darwin-amd64" ]; then \
		echo "Contents of $(BUILD_DIR)/darwin-amd64:"; \
		ls -la $(BUILD_DIR)/darwin-amd64/; \
		cd $(BUILD_DIR)/darwin-amd64 && zip -r ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.zip .; \
	elif [ -f "$(BUILD_DIR)/darwin-amd64/$(APP_NAME)" ]; then \
		echo "Creating ZIP from binary only"; \
		cd $(BUILD_DIR)/darwin-amd64 && zip ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.zip $(APP_NAME); \
	else \
		echo "macOS Intel build not found. Creating empty marker file."; \
		echo "macOS Intel build failed" > $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.txt; \
	fi

package-zip-arm64:
	@echo "Creating macOS ZIP for Apple Silicon..."
	@mkdir -p $(DIST_DIR)
	@if [ -d "$(BUILD_DIR)/darwin-arm64" ]; then \
		echo "Contents of $(BUILD_DIR)/darwin-arm64:"; \
		ls -la $(BUILD_DIR)/darwin-arm64/; \
		cd $(BUILD_DIR)/darwin-arm64 && zip -r ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.zip .; \
	elif [ -f "$(BUILD_DIR)/darwin-arm64/$(APP_NAME)" ]; then \
		echo "Creating ZIP from binary only"; \
		cd $(BUILD_DIR)/darwin-arm64 && zip ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.zip $(APP_NAME); \
	else \
		echo "macOS ARM64 build not found. Creating empty marker file."; \
		echo "macOS ARM64 build failed" > $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.txt; \
	fi

package-darwin-amd64: 
	@echo "Packaging macOS Intel..."
	@mkdir -p $(DIST_DIR)
	@echo "Build directory contents:"
	@ls -la $(BUILD_DIR)/ || echo "Build directory not found"
	@$(MAKE) package-dmg-amd64
	@$(MAKE) package-zip-amd64
	@echo "Final dist directory contents:"
	@ls -la $(DIST_DIR)/ || echo "Dist directory not found"

package-darwin-arm64: 
	@echo "Packaging macOS ARM..."
	@mkdir -p $(DIST_DIR)
	@echo "Build directory contents:"
	@ls -la $(BUILD_DIR)/ || echo "Build directory not found"
	@$(MAKE) package-dmg-arm64
	@$(MAKE) package-zip-arm64
	@echo "Final dist directory contents:"
	@ls -la $(DIST_DIR)/ || echo "Dist directory not found"

package-linux: package-deb package-tar

package-darwin: package-dmg package-zip-amd64 package-zip-arm64

package: package-linux

release: clean package
	@echo "Release artifacts created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/