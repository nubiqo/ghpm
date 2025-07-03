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

## Upgrading to a New Version

To upgrade the application, it is recommended to first uninstall the previous version and then follow the installation instructions for the new release.

### Uninstalling

#### Linux (Debian/Ubuntu)

Use the following command to remove the installed package:
```sh
sudo dpkg -r github-profile-manager
```

#### Linux (Other distributions)

If you installed from a `.tar.xz` file, simply delete the directory where you extracted the application.

#### macOS

Drag the `GitHub Profile Manager` application from your `Applications` folder to the Trash.
