# ChemCLI

ChemCLI, short for Chemotion CLI, is a tool to help you manage Chemotion ELN on a machine. The goal is to make installation, maintenance and upgradation of (multiple instances of) Chemotion as easy as possible.

## Compatibility with Chemotion ELN

ChemCLI tool supports the following versions of Chemotion ELN:

| ELN Version                                                            | `docker-compose.yml` file                                                                                                   | Supported by chemCLI version                                           |
| ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| [v1.5.1](https://github.com/ComPlat/chemotion_ELN/releases/tag/v1.5.1) | [eln-1.5.1](https://raw.githubusercontent.com/ptrxyz/chemotion/44399a444948a99ddd50ecbea0fbcb84d159f2cc/docker-compose.yml) | [0.2.x](https://github.com/Chemotion/ChemCLI/releases/tag/0.2.2)       |
| [v1.3.1](https://github.com/ComPlat/chemotion_ELN/releases/tag/v1.3.1) | [eln-1.3.1p220712](https://github.com/ptrxyz/chemotion/blob/cb906ce8d37e0a173ae66b3696a9f039f540eace/docker-compose.yml)    | [0.1.x](https://github.com/Chemotion/ChemCLI/releases/tag/0.1.6-alpha) |

> Chemotion ELN version [1.4.x](https://github.com/ComPlat/chemotion_ELN/releases/tag/v1.4.1-3) is **not supported** because it requires [manual changes](https://chemotion.net/docs/eln/install_configure/manual_install#only-when-installing-or-upgrading-to-version-141) to the installation after downloading the [`docker-compose.yml`](https://raw.githubusercontent.com/ptrxyz/chemotion/ba4b4620ab2aaa6be32df78189c29970335b1989/docker-compose.yml) file.

## Concept for chemCLI

The commands have the following general layout:

```
general: cli-executable  <resource>  <command>  <flags>
         └─────┬──────┘  └───┬────┘  └───┬───┘  └──┬──┘
example:    chemCLI       instance     logs     --all
```

Following features have been implemented:

- ✔ Installation & Deployment: `./chemCLI` > `install` installs a production instance that is ready to use.
- ✔ Instance life cycle commands: `./chemCLI` > `on|off|restart` and `./chemCLI` > `instance` > `stats|ping|logs`.
- ✔ Upgrade: `./chemCLI` > `instance` > `upgrade` to upgrade an existing Chemotion instance.
- ✔ Backups: `./chemCLI` > `instance` > `backup` to save the data associated with an instance.
- ✔ Restore: `./chemCLI` > `instance` > `restore` to save the data associated with an instance into a new instance.
- ✔ Multiple instances: `./chemCLI` > `instance` > `new|list|switch|remove` can be used to manage multiple instances.
- ✔ Administrative consoles: `./chemCLI` > `instance` > `consoles` > `<console_name>` to drop into `shell`, `postgreSQL` and `rails` consoles.

Following features are under consideration planned:

- Manage Settings: `./chemCLI instance settings` to configure settings of an instance and to assist with auto-configuring wizards.
- A GUI interface to monitor instances being run.
- Have another feature in mind: [Open an issue](https://github.com/Chemotion/ChemCLI/issues) or contact our [helpdesk](https://chemotion.net/helpdesk)!

## Download

### Getting the tool

The ChemCLI tool is a binary file called `chemCLI` and needs no installation. The only prerequisite is that you install [Docker Desktop](https://www.docker.com/products/docker-desktop/) (and, on Windows, [WSL](https://docs.microsoft.com/en-us/windows/wsl/install)). Depending on your OS, you can download the latest release of the CLI from [here](https://github.com/Chemotion/ChemCLI/releases/). Builds for the following systems are available:

- Linux, amd64
- Windows, amd64; remember to turn on [Docker integration with WSL](https://docs.docker.com/desktop/windows/wsl/)
- macOS, apple-silicon
- macOS, amd64

Please be sure that you have both, `docker` and `docker compose` commands. This should be the case if you install Docker Desktop following the instructions [here](https://docs.docker.com/desktop/#download-and-install). If you choose to install only Docker Engine, then please make sure that you **also** have `docker compose` as a command (as opposed to `docker-compose`). On Linux, you might have to install the [`docker-compose-plugin`](https://docs.docker.com/compose/install/linux/#install-using-the-repository) to achieve this.

These binary builds rely on libraries of the underlying operating system: if they do not work on your system for some reason, please create an [issue here](https://github.com/Chemotion/ChemCLI/issues) and we will try to provide you a binary build as soon as possible. If you feel like it, you can always compile the `go` source code on your own.

### Making it an executable

| OS                       | How to make it an executable | How to run the executable |
| ------------------------ | :--------------------------: | ------------------------- |
| Linux (Ubuntu, WSL etc.) |     `chmod u+x chemCLI`      | `./chemCLI`               |
| Windows (Powershell)     |       (nothing to do)        | `.\chemCLI.exe`           |
| macOS (intel/amd64)^     | `chmod u+x chemCLI.amd.osx`  | `./chemCLI.amd.osx`       |
| macOS (apple-silicon)^   | `chmod u+x chemCLI.arm.osx`  | `./chemCLI.arm.osx`       |

^On macOS, if the there is a security pop-up when running the command, please also `Allow` the executable in `System Preferences > Security & Privacy`.

#### Important Notes:

- All commands here, and all the documentation of the tool, use `./chemCLI` when describing how to run the executable. However, your specific command to run the executable is given in the table above.
- If possible, do not rename the executable, or rename/remove files and folders created by it. All reasonable operations can be done using chemCLI; manual operations might break the chemCLI's ability to understand how things are laid out on your machine.

## First run

### Make a dedicated folder

Make a folder where you want to store installation(s) of Chemotion ELN. Ideally this folder should be in the largest drive (in terms of free space) of your system. Remember that Chemotion also uses space via Docker (docker containers, volumes etc.) and therefore you need to make sure that your system partition has abundant free space.

### Install

To begin with installation, run the executable (`./chemCLI`) and follow the prompt. The first installation can take really long time (15-30 minutes depending on your download and processor speeds). Please be aware that instance names must be lowercase and cannot container periods (`.`).

This will create the first (production-grade) `instance` of Chemotion on your system. Generally, this is suffice if you want to use Chemotion in a single scientific group/lab. By default

- this first instance will be available on port 4000
- this first instance will be the `selected` instance.

> :warning: **chem_cli.yml**: Installation also creates a file called `chem_cli.yml`. This file is critical as it contains information regarding existing installations. Removing the file will render chemCLI clueless about existing installations and it will behave as if Chemotion was never installed. Please do not remove the file. Ideally there should be no need for you to modify it manually.

### The `selected` instance

Once you install multiple instances of Chemotion, the actions of chemCLI pertain to only one of them i.e. you actively operate only one of them when you run a command in the `./chemCLI` > `instance` section. This instance is referred to as the `selected` instance and its name is stored in a local file (`chem_cli.yml`). You can do `./chemCLI instance switch` to switch to another instance if you have more than one instance.

You can also select an instance _temporarily_ by giving its name to the CLI as a flag e.g. `./chemCLI instance status -i my-other-instance`.

### Start and Stop Chemotion

To turn on, and off, the `selected` instance, issue the commands:

- `./chemCLI on`, and
- `./chemCLI off`.

### Upgrading an instance (for ELN versions 1.3 and above)

As long as you installed an instance of Chemotion using this tool, the upgrade process is quite straightforward:

- First make sure that you have the [latest version of this tool](#updating-the-tool).
- Prepare for update by running `./chemCLI` > `instance` > `upgrade` > `pull image only`. This will download the latest chemotion image from the internet if not already present on the system. Doing this saves you time later (during scheduled downtime).
- Schedule a downtime of at least 15 minutes; more if you have a lot of data that needs to backed up. During the downtime, run `./chemCLI` > `instance` > `all actions` to backup your data followed by an upgrade of the instance.

> When upgrading from ELN version 1.3, please create a backup using chemCLI version 0.2.2 or above. This is because the data backup script provided inside the container is broken and this is fixed by chemCLI tool.

### Uninstallation

> :warning: be sure about what you want to do!

You can uninstall everything created by chemCLI by running: `./chemCLI` > `advanced` > `uninstall`. Last you can simply delete the downloaded binary itself.

## Updating the Tool

chemCLI (version 0.2 onwards) is configured to check for new releases of itself every 24 hours with the servers of GitHub. If a new version is available, it will inform you of it. You can then update the tool by going to `./chemCLI` > `advanced` > `update - chemCLI`.

> Respecting your privacy: You can disable this automatic checking for next 100 years by running `./chemCLI advanced update --disable-autocheck`.

### Updating from version 0.1.x

If you are using chemCLI version 0.1 (then called `chemotion CLI`), please do the following to update to the latest version:

- [Download](#download) the latest version and [make it an executable](#making-it-an-executable).
- Create copy of the file `chemotion-cli.yml` as `chem_cli.yml` by executing: `cp chemotion-cli.yml chem_cli.yml`.
- Run the new executable (`./chemCLI`). It should guide you through an automated update. After a successful update, it is safe to remove `chemotion-cli.yml` and `chemotion-cli.log` files.
- This update does not upgrade the instances

> Please note that support for version 0.1.x will be completely deprecated on 31.12.2023.

## Advanced Usage

### Using flags

chemCLI has a lot to offer, with a few advanced features available exclusively via flags. Feel free to use `--help` option at the end of the command and its subcommands to explore more.

A particular construct worth noting is using the `./chemCLI instance restore` command to create the **first** instance while restoring data from a previous instance into it. This can be useful for moving from non-docker and/or non-chemCLI based installations to a chemCLI managed installation.

```bash
./chemCLI instance restore --name first-instance --data /absolute/path/to/backup.data.tar.gz --db /absolute/path/to/backup.sql.gz --address https://chemotion.myuni.de
```

### Silent and Debug Use

Almost all features of chemCLI can be used in silent mode i.e. without any input/interaction from user as long as all required pieces of information have been provided using flags. In silent mode, most of the output from the CLI (but not that of docker) is logged only in the log file, and not put on screen.

To use chemCLI in silent mode, add the flag `-q`/`--quiet` to your command. The CLI will then use default values and other flags to try and accomplish the action. Examples:

```bash
./chemCLI -q instance new --name second-instance --address https://myuni.de:3000 --use ../my/path/docker-compose.yml
```

```bash
./chemCLI -q instance backup # for running backups silently
```

Similarly, the CLI can be run in Debug mode when you encounter an error. This produces a very detailed log file containing a trace of actions you undertake. Telling us about the error and sending us the log file can help us a lot when it comes to helping you. Please audit the debug file for any personal information (such as username etc.) before sending it to us.

## Known limitations and bugs

- `./chemCLI off`: does not lead to exit of containers with exit code 0.
- Everything happens in the folder (and subfolders) where `./chemCLI` is executed. All files and folders are expected to be there; otherwise failures can happen.
