# CHANGELOG of Chemotion CLI tool

## Version 0.2.0-alpha

The main changes are as follows:

- `chemotion instance consoles` command that allows you to enter the console of a running instance.
- `chemotion instance ping` command that checks if the instance is up, running and available at the specified URL.
- `chemotion advanced update` command that updates the CLI tool itself. The tool now also checks (once a day) if an update to itself is available and displays a reminder if this is the case.
- `chemotion instance new` now uses the latest availble `docker-compose.yml` file, not a hard-coded one.
- The tool now copies the entries in the `instances:<instance_name>:environment` section to the `./instances/<instance_folder>/shared/pullin/.env` file whenever an instance is turned on (or restarted). The idea is that a user can edit the `chemotion-cli.yml` file (more) easily.
- Responsive menus: only actions that make sense are displayed e.g. the main menu shows an `off` option if the selected instance is already running.
- `Back` option: It is now (mostly) possible to go to the menu above using a `back` option. The tool exits once **a task** is completed -- this is an intended feature.
- The tool is (milliseconds) slower to start now precisely because i now checks on the status of the instance on launch.
- Bugfixes
- _Changes that effect the user's files are as follows.:_

> The first difference is formatting of the `chemotion-cli.yml` file.

- The global keys that handle state of the tool i.e. `selected`, `quiet` and `debug` have now been moved to `cli_state:selected`, `cli_state:quiet` and `cli_state:debug` respectively.
- The `instances:<instance_name>:address` and `instances:<instance_name>:protocol` keys have been removed. Instead, we have `instances:<instance_name>:accessaddress` which stores the full URL that is used to access the ELN instance.
- A new key called `instances:<instance_name>:environment` has been introduced. This is now used to create the `shared/pullin/.env` file **everytime** the instance is (re)started. **Please** move all your `key=value` pairs from this `.env` file to the `chemotion-cli.yml` in `key: value` format as sub-keys of the `instances:<instance_name>:environment` key.
- With these changes, the version of this YAML file has been changed from `"1.0"` to `"1.1"`.

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
instances:
  main:
    accessaddress: http://mynotebook.kit.edu
    environment:
      url_host: ifgs6.ifg.kit.edu
      url_protocol: http
    kind: Production
    name: main-ee5e5044
    port: 4000
  second:
    accessaddress: http://localhost:4100
    environment:
      url_host: localhost:4100
      url_protocol: http
      smtp_port: ...<key-value pairs from /shared/pullin/.env file>...
    kind: Production
    name: second-ff6f6535
    port: 4100
version: "1.1"
```

> The second difference is splitting of `docker-compose.yml` file into two files.

So far dockerized installations of Chemotion have relied on `docker-compose.yml` file from [here](https://github.com/ptrxyz/chemotion).

The CLI in version 0.1.x-alpha diverged from this by modifying the file to suit the needs of the CLI by

1. changing the `services:eln:ports` key
2. including this label on `networks`, `services` and `volumes`: `net.chemotion.cli.project: <instance_name>-<instance_uniqueID>
3. including names on the `volumes` so that they are named the following: `<instance_name>-<instance_uniqueID>_chemotion_<app|data|db|spectra>`.

Version 0.2.x onwards, we refrain from modifying the `docker-compose.yml` file, making only one change in it (Change 1. is still done.). Changes 2. and 3. are inlcuded in the configuration by adding a new file called `docker-compose.cli.yml` (that we use in addition to the `docker-compose.yml` file). The `docker compose` tool seamlessly merges the two files when reading them.
