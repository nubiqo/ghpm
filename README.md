# GitHub Profile Manager (ghpm)

A desktop application to manage multiple GitHub profiles with ease, including git configuration and SSH keys.

## Installation

You can download the latest version of the application from the [GitHub Releases page](https://github.com/your-repo/ghpm/releases).

### Linux

#### Debian/Ubuntu (.deb)

1.  Download the `github-profile-manager_<version>_amd64.deb` package from the releases page.
2.  Install the package using the following command:

    ```sh
    sudo dpkg -i /path/to/github-profile-manager_<version>_amd64.deb
    ```

#### Other Linux Distributions (.tar.xz)

1.  Download the `GitHubProfileManager-<version>-linux-amd64.tar.xz` archive from the releases page.
2.  Extract the archive:

    ```sh
    tar -xf /path/to/GitHubProfileManager-<version>-linux-amd64.tar.xz
    ```

3.  Run the application executable from the extracted directory.

### macOS

#### .dmg (Recommended)

1.  Download the `.dmg` file for your architecture (Intel or Apple Silicon) from the releases page.
2.  Open the `.dmg` file.
3.  Drag the `GitHub Profile Manager` application into your `Applications` folder.

#### .zip

1.  Download the `.zip` file for your architecture.
2.  Unzip the file.
3.  Move the `GitHub Profile Manager` application to your `Applications` folder.

## Upgrading / Updating

Good news: your profiles live in `~/.ghpm`, so updating the app will not touch your saved profiles.

### Linux

#### Debian/Ubuntu (.deb)

Option A — In‑place upgrade (recommended):

```sh
cd /path/to/downloads
sudo apt install -y ./github-profile-manager_<version>_amd64.deb
```

This will upgrade or install as needed, resolving dependencies automatically.

Option B — Using `dpkg`:

```sh
sudo dpkg -i /path/to/github-profile-manager_<version>_amd64.deb
```

If `dpkg` reports missing dependencies, run `sudo apt -f install`.

#### Other Linux (.tar.xz)

1) Extract the new archive and replace the old binary/directory you were using:

```sh
tar -xf /path/to/GitHubProfileManager-<version>-linux-amd64.tar.xz -C /opt
sudo ln -snf /opt/GitHubProfileManager-<version>/github-profile-manager /usr/local/bin/github-profile-manager
```

2) If you previously created a desktop entry manually, update it to point to the new path.

### macOS

#### DMG

1) Open the new `.dmg`.
2) Drag `GitHub Profile Manager.app` into `Applications`.
3) When prompted, choose “Replace”.

#### ZIP

1) Unzip the new archive.
2) Move `GitHub Profile Manager.app` to `Applications`, replacing the existing one.

### Verifying the Update

- Launch the app and check the footer — it shows the version as: `© huzaifa • vX.Y.Z`.
- Your profiles should still appear. If they don’t, ensure `~/.ghpm` exists and contains your JSON profiles.

### Uninstalling (if needed)

#### Debian/Ubuntu

```sh
sudo dpkg -r github-profile-manager
```

#### Other Linux (.tar.xz)

Remove the extracted directory and any symlinks you created (for example in `/usr/local/bin`).

#### macOS

Delete `GitHub Profile Manager.app` from your `Applications` folder.
