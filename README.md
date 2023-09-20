# ChemCLI

ChemCLI, short for Chemotion CLI, is a tool to help you manage Chemotion ELN on a machine. The goal is to make installation, maintenance and upgradation of (multiple instances of) Chemotion as easy as possible.

## Important Announcement

Have a bug that looks as follows when trying to update?

```sh
pg_dump: error: connection to server at "db" (172.21.0.2), port 5432 failed: FATAL:  database "chemotion" does not exist
```

Then please have a look at the `docker-compose.cli.yml` files in the `instances` folder. The section `services.executor.image` should be changed so that it matches `services.eln.image` in the `docker-compose.yml` file. Otherwise, you will likely bug that looks as follows when trying to upgrade:

## Note

If you are using ChemCLI versions 0.2.0 to 0.2.3, you will have to run the following command to (force) run the auto-update feature: `./chemCLI advanced update --force`. You can also (always) [download a new executable of ChemCLI](#download) to manually update the tool. (Apologies for the bug!)

> Please note that support for version 0.1.x will be completely deprecated on 31.12.2023. The codes that help you migrate from version 0.1 to 0.2 will be removed in releases after 31.12.2023.

## Compatibility with Chemotion ELN

ChemCLI tool supports the following versions of Chemotion ELN:

| ELN Version                                                            | `docker-compose.yml` file                                                                                                            | Supported by chemCLI version                                  |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------- |
| [v1.7.x](https://github.com/ComPlat/chemotion_ELN/releases/tag/v1.7.3) | [eln-1.7.3](https://raw.githubusercontent.com/Chemotion/ChemCLI/446e23aac22b4477887d474e1c79c744f415e1f5/payload/docker-compose.yml) | [0.2.x](https://github.com/Chemotion/ChemCLI/releases/latest) |
| [v1.6.x](https://github.com/ComPlat/chemotion_ELN/releases/tag/v1.6.2) | [eln-1.6.2](https://raw.githubusercontent.com/Chemotion/ChemCLI/e577832edaba14fa21ee9aa9288e4b00052729c8/payload/docker-compose.yml) | [0.2.x](https://github.com/Chemotion/ChemCLI/releases/latest) |
| [v1.5.x](https://github.com/ComPlat/chemotion_ELN/releases/tag/v1.5.4) | [eln-1.5.4](https://raw.githubusercontent.com/Chemotion/ChemCLI/548ead617a552307f30d5051e72c01d95e99b30f/payload/docker-compose.yml) | [0.2.x](https://github.com/Chemotion/ChemCLI/releases/latest) |

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
- ✔ Users: `./chemCLI` > `instance` > `users` > `create|list|update|describe|delete` to manage users in an instance. Particularly, `./chemCLI` > `instance` > `users` > `update` can be used to modify password for users who forget theirs.

## Download

### Getting the tool

The ChemCLI tool is a binary file called `chemCLI` and needs no installation. The only prerequisite is that you install [Docker Desktop](https://www.docker.com/products/docker-desktop/) (and, on Windows, [WSL](https://docs.microsoft.com/en-us/windows/wsl/install)). Depending on your OS, you can download the latest release of the CLI from [here](https://github.com/Chemotion/ChemCLI/releases/). Builds for the following systems are available:

- Linux, amd64
- Windows, amd64; remember to turn on [Docker integration with WSL](https://docs.docker.com/desktop/windows/wsl/)
- macOS, apple-silicon
- macOS, amd64

Please be sure that you have both, `docker` and `docker compose` commands. This should be the case if you install Docker Desktop following the instructions [here](https://docs.docker.com/desktop/#download-and-install). If you choose to install only Docker Engine, then please make sure that you **also** have `docker compose` as a command (as opposed to `docker-compose`). On Linux, you might have to install the [`docker-compose-plugin`](https://docs.docker.com/compose/install/linux/#install-using-the-repository) to achieve this.

These binary builds should not rely on libraries of the underlying operating system: if they still do not work on your system for some reason, please create an [issue here](https://github.com/Chemotion/ChemCLI/issues) and we will try to provide you a binary build as soon as possible. If you feel like it, you can always compile the `go` source code on your own.

### Making it an executable

| OS                       | How to make it an executable | How to run the executable |
| ------------------------ | :--------------------------: | ------------------------- |
| Linux (Ubuntu, WSL etc.) |     `chmod u+x chemCLI`      | `./chemCLI`               |
| Windows (Powershell)\*   |       (nothing to do)        | `.\chemCLI.exe`           |
| macOS (intel/amd64)^     | `chmod u+x chemCLI.amd.osx`  | `./chemCLI.amd.osx`       |
| macOS (apple-silicon)^   | `chmod u+x chemCLI.arm.osx`  | `./chemCLI.arm.osx`       |

\*On Windows, it is recommended to use [Powershell 7](https://learn.microsoft.com/en-us/powershell/scripting/install/installing-powershell) instead of the one provided natively ([confusingly called Windows Powershell](https://learn.microsoft.com/en-us/powershell/scripting/install/installing-powershell)). In any case, it is necessary to have `pwsh` in your `$PATH`.

^On macOS, if the there is a security pop-up when running the command, please also `Allow` the executable in `System Preferences > Security & Privacy`.

#### Important Notes:

- All commands here, and all the documentation of the tool, use `./chemCLI` when describing how to run the executable. However, your specific command to run the executable is given in the table above.
- If possible, do not rename the executable, or rename/remove files and folders created by it. All reasonable operations can be done using chemCLI; manual operations might break the chemCLI's ability to understand how things are laid out on your machine.

## First run

### Make a dedicated folder

Make a folder where you want to store installation(s) of Chemotion ELN. Ideally this folder should be in the largest drive (in terms of free space) of your system. Remember that Chemotion also uses space via Docker (docker containers, volumes etc.) and therefore you need to make sure that your system partition has abundant free space.

### Install

To begin with installation, run the executable (`./chemCLI`) and follow the prompt. The first installation can take really long time (15-30 minutes depending on your download and processor speeds). Please be aware that instance names must be lowercase and cannot contain periods (`.`).

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

### Backup an instance

:::caution WARNING
This backup process does not backup data that reside out docker e.g. `services` and `shared` folders. These folder need to be backed up by you separately.
:::

You can backup an instance of the ELN, installed using the CLI, by running the following command:
`./chemCLI instance backup  -i <name_of_instance> -q`. This command must be run individually for every instance of Chemotion ELN that you have. Two new backup files are created _for each execution_ of the command inside the `instances/<name_of_instance-xxxxxxx>/shared/backup` folder. (These two files can be used to restore data into a new instance using the `./chemCLI instance restore` command.)

If running this as a [cron](https://en.wikipedia.org/wiki/Cron) job, remember to change into the folder where the ChemCLI executable exists. To do this, include `cd path/to/the/folder` in your job script. Therefore an example crontab file that runs at [03:00 on every day-of-week from Tuesday through Saturday](https://crontab.guru/#0_3_*_*_2-6) would look like as follows:

```cron
0 3 * * 2-6 cd /home/admin/installations/chemotion_ELN && ./chemCLI instance backup -i prodinstance -q
```

Please also refer to the notes [here](manual_install#backing-up-and-restoring-your-data) for a better understanding of the backup process.

### Upgrading an instance (for ELN versions 1.3 and above)

As long as you installed an instance of Chemotion using this tool, the upgrade process is quite straightforward:

- First make sure that you have the [latest version of this tool](#updating-the-tool).
- Prepare for update by running `./chemCLI` > `instance` > `upgrade` > `pull image only`. This will download the latest Chemotion image from the internet if not already present on the system. Doing this saves you time later (during scheduled downtime).
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

For example:

```bash
./chemCLI instance restore --name first-instance --data /absolute/path/to/backup.data.tar.gz --db /absolute/path/to/backup.sql.gz --suffix 4cfcfd0c --address https://chemotion.myuni.de --use https://github.com/Chemotion/ChemCLI/releases/download/0.2.7/docker-compose.yml
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

- Everything happens in the folder (and subfolders) where `./chemCLI` is executed. All files and folders are expected to be there; otherwise failures can happen. The user executing ChemCLI is expected to have all file permissions for this folder.
- Have a bug that looks as follows when trying to update?

```sh
pg_dump: error: connection to server at "db" (172.21.0.2), port 5432 failed: FATAL:  database "chemotion" does not exist
```

Then please have a look at the `docker-compose.cli.yml` files in the `instances` folder. The section `services.executor.image` should be changed so that it matches `services.eln.image` in the `docker-compose.yml` file. Otherwise, you will likely bug that looks as follows when trying to upgrade:

- If you are using ChemCLI versions 0.2.0 to 0.2.3, you will have to run the following command to (force) run the auto-update feature: `./chemCLI advanced update --force`. You can also (always) [download a new executable of ChemCLI](#download) to manually update the tool.

## Acknowledgments

This project has been funded by the **[DFG](https://www.dfg.de/)**.

![DFG Logo](https://www.dfg.de/zentralablage/bilder/service/logos_corporate_design/logo_negativ_267.png)

Funded by the [Deutsche Forschungsgemeinschaft (DFG, German Research Foundation)](https://www.dfg.de/) under the [National Research Data Infrastructure – NFDI4Chem](https://nfdi4chem.de/) project – Projektnummer **441958208** – since 2020.
