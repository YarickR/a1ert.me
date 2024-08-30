package di
import (
    "sync"
)
type DagMsg struct {
    Data    map[string]interface{}
    Channel ChannelPtr
}
type DagMsgPtr *DagMsg

type PlugCommPtr *PlugComm
type PlugComm struct {
    TxChan chan DagMsg
    CTS    bool // This plugin is clear to send 
    Buffer []DagMsg
}

type ModLoadConfigHook      func(config interface{}, isGlobal bool, path string) (PluginConfig, error)
type ModReceiveMsgHook      func(ChPCtx ChanPlugCtx) (DagMsg, error)
type ModSendMsgHook         func(msg DagMsg, ChPCtx ChanPlugCtx) error 
type ModHookTable struct {
    LoadConfigHook      ModLoadConfigHook
    ReceiveMsgHook      ModReceiveMsgHook
    SendMsgHook         ModSendMsgHook
//    ProcessEventHook    ModProcessEventHook
 //   InGoroHook          ModInGoroHook 
 //   OutGoroHook         ModOutGoroHook 
}

type ModHooksFunc func()    (ModHookTable, error)

type ModulePtr *Module
type Module struct {
    Name    string
    Hooks   ModHookTable
}
type PluginConfig interface{} // Opaque, module-dependent 

type PluginPtr *Plugin
type Plugin struct {
	Name   string
    Type   int
	Module Module
	Config PluginConfig
}

const (
    PT_IN       = 1 // Plugin type
    PT_OUT      = 2
    PT_PROC     = 4
)

const (
    CKW_GLOBAL  = 1 // keyword allowed in global config
    CKW_CHANNEL = 2 // keyword allowed in per channel config
)
type ChanPlugCtx struct {
    Plugin  PluginPtr
    Config  PluginConfig
}

type ChannelPtr *Channel
type Channel struct {
    Name       string 
    Descr      string 
    Rules      []RulePtr
    Sinks      []ChannelPtr
    InPlugs    []ChanPlugCtx
    OutPlugs   []ChanPlugCtx
    ProcPlugs  []ChanPlugCtx
}
type RulePtr *Rule
type Rule struct {
    RuleId       uint32 
    SrcChId      string
    RuleStr      string 
    CondLink     string 
    Root         RulePart
}
type PnFunc struct {
  NumArgs   int
  Function  RulePartFunc
}

type RuleParserCtx struct {
  PartPtr *RulePart
  CurrArg  int
}
const RULE_PARSER_MF_MASK       = 0xF
const RULE_PARSER_MF_IN_QUOTE   = 0x10
const RULE_PARSER_MF_IN_BQ      = 0x20

type RuleParserStateType uint
const  (
    RuleParserInvalidState RuleParserStateType = 0

    RuleParserInArgKey          = 1
    RuleParserInArgStr          = 2
    RuleParserInArgNum          = 3
    RuleParserInArgFunc         = 4
    RuleParserInArgSet          = 5
    RuleParserAfterArg          = 6
    RuleParserLookingForArg     = 7
)

type RulePartArgType uint
const (
    InvalidPartArgType RulePartArgType = 0
    PartArgTypeKey  = 1
    PartArgTypeStr  = 2
    PartArgTypeNum  = 3
    PartArgTypeFunc = 4
    PartArgTypeSet  = 5
)

type RulePartArg struct {
    ArgType    RulePartArgType
    ArgValue   interface{}
}

type RulePartFunc func(args []RulePartArg, event Event) interface{}

type RulePart struct {
    Function     RulePartFunc
    Arguments  []RulePartArg
}

type TemplatePtr *Template
type Template struct {  
    Contents string 
}
type MsgContents struct {
    Receiver        string          `json:"receiver"`
    Status          string          `json:"status"`
    Alerts        []interface{}     `json:"alerts"`
}

type AlertMsg struct {
    ReceivedTS  float64     `json:"received"`
    Message     MsgContents `json:"message"`
    Matches     uint32
}

type GlobalConfig struct {
	Plugins			map[string]PluginPtr 
	Channels		map[string]ChannelPtr
	Templates		map[string]TemplatePtr
}

type GlobalChans struct {
    InChan chan DagMsg
    OutChan chan DagMsg
}
