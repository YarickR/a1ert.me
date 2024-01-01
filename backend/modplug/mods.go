package modplug

import (
  "dagproc/event"
)


type ModHooksFunc func()    ModHookTable
type ModLoadConfigHook      func(config []interface{}) error
type ModReceiveEventHook    func() (event.Ev, error)
type ModSendEventHook       func(event.Ev) error
type ModProcessEventHook    func(event.Ev) error
type ModHookTable struct {
    LoadConfigHook      ModLoadConfigHook
    ReceiveEventHook    ModReceiveEventHook
    SendEventHook       ModSendEventHook
    ProcessEventHook    ModProcessEventHook
}

type ModConfig []interface{}

type ModuleInfo struct {
    Name    string
    Config  ModConfig
    Hooks   ModHookTable
}
