# Dev Notes

## Tool Initialization: Flow of Control

1. `func main()` [main.go]
2. `init()` [root.go]
   1. `initLog()` [initialize.go]
   2. `initFlags()` [initialize.go]
   3. `cobra.OnInitialize()` [root.go]
      1. `initConf()` [initialize.go]
      2. `bindFlags()` [initialize.go]
   4. `rootCmd.SetVersionTemplate()` [root.go]
3. `func Execute()` [root.go]
4. `func rootCmd.Execute()` [root.go]
   1. `logwhere()`
   2. `confirmVirtualizer()`
   3. `instanceValidate()`
