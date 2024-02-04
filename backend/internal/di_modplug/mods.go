package di_modplug
import (
    "fmt"
    "strings"
    "errors"
    "dagproc/internal/di"
)
var ( 
    ModHookMap  map[string]di.ModHooksFunc // ModInfoMap is populated manually early in main()
    ModMap      map[string]di.Module  // ModMap is populated by reading ModInfoMap and calling ModHooksFunc's for each module
)


func ValidateHooks(hooks []string, module string) error {
    var v string 
    var m di.Module
    var ok bool
    m, ok = ModMap[module]
    if !ok {
        return errors.New(fmt.Sprintf("Unknown module %s", module)) 
    }
    ok = true
    for _, v = range hooks {
        switch strings.ToLower(v) {
            case "in":
                if (m.Hooks.ReceiveEventHook == nil) {
                    ok = false;
                }
            case "out":
                if (m.Hooks.SendEventHook == nil) {
                    ok = false;
                }
            case "process":
                if (m.Hooks.ProcessEventHook == nil) {
                    ok = false;
                }
        }
        if (!ok) {
            return errors.New(fmt.Sprintf("Module %s does not declare %s hook", module, v))
        }
    }
    return nil
}