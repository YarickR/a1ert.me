package di_modplug
import (
    "fmt"
    "strings"
    "errors"
    "dagproc/internal/di"
)

func ValidatePluginType(pType string, module string) error {
    var 
    (
        m di.Module
        ok,ret bool
        hookD map[string]bool
    )
    m, ok = di.ModMap[module]
    if !ok {
        return errors.New(fmt.Sprintf("Unknown module %s", module)) 
    }
    hookD = map[string]bool { 
            "in":   m.Hooks.ReceiveMsgHook != nil, 
            "out":  m.Hooks.SendMsgHook    != nil,
            "proc": m.Hooks.ProcessMsgHook != nil,
    }
    ret, ok = hookD[strings.ToLower(pType)]
    if !ok {
        return fmt.Errorf("Unknown hook type %s", pType)
    }
    if !ret {
        return fmt.Errorf("Module %s missing hook type %s", module, pType)
    }
    return nil
}