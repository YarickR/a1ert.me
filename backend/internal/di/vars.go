package di

var (
	GCfg GlobalConfig
	ModHookMap  map[string]ModHooksFunc // ModInfoMap is populated manually early in main()
  ModMap      map[string]Module  // ModMap is populated by reading ModInfoMap and calling ModHooksFunc's for each module
)