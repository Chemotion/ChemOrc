# CHANGELOG of Chemotion CLI tool

## Version 0.2

The main changes are as follows:

- `./chemCLI instance consoles` command that allows you to enter the console of a running instance.
- `./chemCLI instance ping` command that checks if the instance is up, running and available at the specified URL.
- `./chemCLI instance restore` command to restore your data into a new instance of Chemotion ELN.
- `./chemCLI advanced update` command that updates the CLI tool itself. The tool now also checks (once a day) if an update to itself is available and displays a reminder if this is the case.
- `./chemCLI instance new` now uses the latest available `docker-compose.yml` file, not a hard-coded one.
- `./chemCLI instance status` is now merged into `./chemCLI instance stats`.
- Responsive menus: only actions that make sense are displayed e.g. the main menu shows an `off` option if the selected instance is already running.
- `Back` option: It is now possible to go to the menu above using a `back` option. The tool exits once **a task** is completed -- this is an intended feature.
- The tool is (milliseconds) slower to start now precisely because it now checks on the status of the instance on launch.
- Bugfixes
- _Changes that effect the user's files are as follows.:_

> `chemotion-cli.yml` is now called `chem_cli.yml`. `chemotion-cli.log` is called `chem_cli.log`. The executable is called `chemCLI` instead of `chemotion`.

> The second difference is formatting of the `chem_cli.yml` file.

- The global keys that handle state of the tool i.e. `selected`, `quiet` and `debug` have now been moved to `cli_state.selected`, `cli_state.quiet` and `cli_state.debug` respectively.
- The `instances.<instance_name>.address` and `instances.<instance_name>.protocol` keys have been removed. Instead, we have `instances.<instance_name>.accessaddress` which stores the full URL that is used to access the ELN instance.

- We have new keys: `cli_state.version`, `cli_state.version_checked_on`
- With these changes, the version of this YAML file has been changed from `"1.0"` to `"2.0"`.

Therefore, if your file looked as follows:

```yaml
instances:
  main:
    address: mynotebook.kit.edu
    debug: false
    kind: Production
    name: main-ee5e5424
    port: 4000
    protocol: http
    quiet: false
  second:
    address: localhost
    debug: false
    kind: Production
    name: second-ff6f6535
    port: 4100
    protocol: http
    quiet: false
selected: main
version: "1.0"
```

It should now look as follows:

```yaml
cli_state:
  debug: false
  quiet: false
  selected: main
  version: 0.2.2
  version_checked_on: 2023-04-06T10:54:40.818796948+02:00
instances:
  main:
    accessaddress: http://mynotebook.kit.edu
    image: ptrxyz/eln-1.3.1p220712
    kind: Production
    name: main-ee5e5044
    port: 4000
  second:
    accessaddress: http://localhost:4100
    image: ptrxyz/chemotion:eln-1.5.1
    kind: Production
    name: second-ff6f6535
    port: 4100
version: "2.0"
```

> The third difference is splitting of `docker-compose.yml` file into two files.

So far dockerized installations of Chemotion have relied on `docker-compose.yml` file from [here](https://github.com/ptrxyz/chemotion).

The CLI in version 0.1.x-alpha diverged from this by modifying the file to suit the needs of the CLI by

1. changing the `services:eln:ports` key
2. including this label on `networks`, `services` and `volumes`: `net.chemotion.cli.project: <instance_name>-<instance_uniqueID>
3. including names on the `volumes` so that they are named the following: `<instance_name>-<instance_uniqueID>_chemotion_<app|data|db|spectra>`.

Version 0.2.x onwards, we refrain from modifying the `docker-compose.yml` file, making only one change in it (Change 1. is still done.). Changes 2. and 3. are included in the configuration by adding a new file called `docker-compose.cli.yml` (that we use in addition to the `docker-compose.yml` file). The `docker compose` tool seamlessly merges the two files when reading them.

The extended file also includes an extra service called `executor` to ease running commands in the `eln` (primary service) without switching on all other services.
