# Chemotion-CLI

##

Chemotion CLI tool is there to help you manage installation(s) of Chemotion on a machine. The goal is to make installation, maintenance and upgradation of Chemotion as easy as possible.

## Installation

### Download the binary

The Chemotion CLI tool is a binary file and needs no installation. The only prerequisite is that you install [Docker Desktop](https://www.docker.com/products/docker-desktop/) (and, on Windows, [WSL](https://docs.microsoft.com/en-us/windows/wsl/install)). Depending on your OS, you can download the lastest release of the CLI from here:

- [Linux, amd64](https://github.com/harivyasi/chemotion/releases/download/latest/chemotion)
- [Windows, amd64](https://github.com/harivyasi/chemotion/releases/download/latest/chemotion.exe); remember to turn on [Docker integration with WSL](https://docs.docker.com/desktop/windows/wsl/).
- [macOS, apple-silicon](https://github.com/harivyasi/chemotion/releases/download/latest/chemotion.arm.x)
- [macOS, amd64](https://github.com/harivyasi/chemotion/releases/download/latest/chemotion.amd.x)

### Make it an executable

On Linux, make this file executable by doing: `chmod u+x chemotion`.

On Windows, the file should be executable by default, i.e. do nothing.

On macOS, make this file executable by doing: `chmod u+x chemotion.amd.x` or `chmod u+x chemotion.arm.x`. If the there is a security pop-up when running the command, please also `Allow` the executable in `System Preferences > Security & Privacy`.

### Important Note:

All commands here, and in the documentation, use term `chemotion` to refer to the executable. Depending on your configuration, you may have to use any one of the following:

- `./chemotion`
- `.\chemotion.exe`
- `./chemotion.arm.x`
- `./chemotion.amd.x`

### First run

#### Make a dedicated folder

Make a folder where you want to store installation(s) of Chemotion. Ideally this folder should be in the largest drive (in terms of free space) of your system. Remember that Chemotion also uses space via Docker (docker containers, volumes etc.) and therefore you need to make sure that your system partition has abundant free space.

#### Install

To begin with installation, execute: `chemotion install` and follow the prompt. The first installation can take really long time (15-30 minutes depending on your download and processor speeds).

This will create the first (production-grade) `instance` of Chemotion on your system. Generally, this is suffice if you want to use Chemotion in a single scientific group/lab. By default

- this first instance will be available on port 4000
- this first instance will be the `chosen` instance (more on this below)

#### Start and Stop Chemotion

To turn on, or off, the `chosen` instance, issue the commands:

- `chemotion on`, please wait for a minute before the instance becomes fully active
- `chemotion off`.

## Uninstallation

> Usual warning of "be sure about what you want to do" applies!

You can uninstall everything created by the CLI tool by running: `chemotion advanced uninstall`. Last you can simply delete the downloaded binary itself.

# Planned concept for CLI

Following features are planned/thought of:

- Installation & Deployment: we plan to implement `chemotion instance install` to install a Chemotion instance
- Upgrade: use `chemotion instance upgrade` to upgrade an existing Chemotion instance
- Backups: `chemotion snapshot create|restore` to savely store your data somewhere
- Instance life cycle commands, such as `chemotion instance start|stop|pause|restart|status`
- Manage Settings: `chemotion settings import|export` to import/export you settings and `chemotion instance configure` to run configuration wizards that help you to create configuration stubs
- Frequently asked for features for the Chemotion Administrator: `chemotion user show|add|delete|password-reset`, `chemotion system info|rails-shell|shell`

We plan to follow one of the following layouts, depending on which one proves to be more handy in every day use.

```
general: cli-executable  <resource>  <command>  <argument>  <flags>
         └─────┬──────┘  └───┬────┘  └───┬───┘  └───┬────┘  └──┬──┘
example:    chemotion     instance    restart   MyInstance  --force
```

```
general: cli-executable  <command>  <resource>  <argument>  <flags>
         └─────┬──────┘  └───┬───┘  └───┬────┘  └───┬────┘  └──┬──┘
example:    chemotion     restart    instance   MyInstance  --force
```

# Known limitations and bugs

- The following flags cannot be specified in the configuration (`chemotion-cli.yml`) file:
  - `--config`: because that creates a circular dependency
  - `chemotion off`: does not lead to exit of containers with exit code 0.
- Everything happens in the folder (and subfolders) of where `chemotion` is executed. All files and folders are expected to be there; otherwise failures can happen.
