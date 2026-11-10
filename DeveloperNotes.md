# Dev Notes

## Tool Initialization: Flow of Control

1. `init()` [root.go], run before `main()`
   1. `initLog()` [initialize.go]: reads `-d`/`-q` directly from `os.Args`, because logging starts before the configuration is read
   2. `initFlags()` [initialize.go]
   3. `cobra.OnInitialize(initConf, bindFlags)` [root.go]: registers the two functions; cobra runs them later, see step 3
   4. `rootCmd.SetVersionTemplate()` [root.go]
   5. `determineShell()` [helper-run.go]
2. `func main()` [main.go] calls `cli.Execute()` [root.go], which calls `rootCmd.Execute()`
3. cobra runs the registered initializers
   1. `initConf()` [initialize.go]: reads `chem_cli.yml` and the selected instance
   2. `bindFlags()` [initialize.go]
4. `rootCmd.PersistentPreRun` [root.go], before every command
   1. debug mode only: write `chem_cli.debug.yml`, then `logwhere()`
   2. `confirmVirtualizer()` [virtualize.go]
   3. `instanceValidate()` on the selected instance [prompt.go]
   4. `autoCheckUpdate()` [root-advanced-update.go]
   5. `upgradeRequired()` [root-instance-upgrade.go]
5. the `Run` of the command that was called; with no arguments, the interactive menu in `rootCmd.Run` [root.go]

Note: a subcommand that defines its own `PersistentPreRun` (e.g. `instance` [root-instance.go]) replaces the root's one, as cobra runs only the closest `PersistentPreRun`.
