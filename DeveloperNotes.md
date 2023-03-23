# Dev Notes

## Tool Initialization: Flow of Control

1. `func main()` [main.go]
2. `init()`
   1. `initLog()`
   2. `initFlags()`
   3. `cobra.OnInitialize()`
      1. `initConf()`
      2. `bindFlags()`
3. `rootCmd.SetVersionTemplate()`
4. `func Execute()`
5. `func rootCmd.Execute()`
6. `logwhere()`
7. `confirmVirtualizer()`
8. `instanceValidate()`
