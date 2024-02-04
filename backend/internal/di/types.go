package di


type Event map[string]interface{}
type ModLoadConfigHook      func(config CFConfig) (PluginConfig, error)
type ModReceiveEventHook    func() (Event, error)
type ModSendEventHook       func(Event) error
type ModProcessEventHook    func(Event) error
type ModHookTable struct {
    LoadConfigHook      ModLoadConfigHook
    ReceiveEventHook    ModReceiveEventHook
    SendEventHook       ModSendEventHook
    ProcessEventHook    ModProcessEventHook
}
type ModHooksFunc func()    (ModHookTable, error)

type ModulePtr *Module
type Module struct {
    Name    string
    Hooks   ModHookTable
}
type CFConfig map[string]interface{} // Config File config
type PluginConfig interface{} // Opaque, module-dependent 

type PluginPtr *Plugin

type Plugin struct {
	Name string
	Module Module
	Config PluginConfig
}

type ChannelPtr *Channel
type Channel struct {
    Id              string 
    Descr           string 
    MsgTemplate    	TemplatePtr
    Rules         []RulePtr
    Sinks         []ChannelPtr
    Plugins       []PluginPtr
}
type RulePtr *Rule
type Rule struct {
    RuleId       uint32 `json:"id"`
    SrcChId      interface{} `json:"src"`
    RuleStr      string `json:"cond"`
    CondLink     string `json:"condfrom"`
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
	Plugins			map[string]Plugin 
	Channels		map[string]Channel
	Templates		map[string]Template
}
