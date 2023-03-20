# Dev Notes

# Tool Initialization: Flow of Control

1. func main() [main.go]
2. func Execute() [root.go]
3. init() [root.go]
   1. initLog()
   2. initFlags()
4. cobra.OnInitialize
   1. initConf()
   2. bindFlags()
5. rootCmd.SetVersionTemplate
6. func rootCmd.Execute()
