package di
type DagMsg struct {
    Id      uint64
    Data    map[string]interface{} // actual message 
    Channel ChannelPtr             // channel to associate message with 
}
type DagMsgPtr *DagMsg

type PlugCommPtr *PlugComm
type PlugComm struct {
    TxChan chan DagMsgPtr
    CTS    bool // This plugin is clear to send 
    Buffer []DagMsgPtr
}

type ModLoadConfigHook      func(config interface{}, isGlobal bool, path string) (PluginConfig, error)
type ModReceiveMsgHook      func(chplct ChanPlugCtxPtr) (DagMsgPtr, error)
type ModSendMsgHook         func(dams DagMsgPtr, chplct ChanPlugCtxPtr) error 
type ModProcessMsgHook      func(dams DagMsgPtr, chplct ChanPlugCtxPtr) (DagMsgPtr, error) 
type ModHookTable struct {
    LoadConfigHook      ModLoadConfigHook
    ReceiveMsgHook      ModReceiveMsgHook
    SendMsgHook         ModSendMsgHook
    ProcessMsgHook      ModProcessMsgHook
}

type ModHooksFunc func()    (ModHookTable, error)

type ModulePtr *Module
type Module struct {
    Name    string
    Hooks   ModHookTable
}
type PluginConfig       interface{} // Opaque, module-dependent 
type PluginConfigCtx    interface{} // Opaque, module-dependent 

type PluginPtr *Plugin
type Plugin struct {
	Name   string
    Type   string
	Module Module
	Config PluginConfig
    Ctx    PluginConfigCtx
}

const (
    CKW_GLOBAL  = 1 // keyword allowed in global config
    CKW_CHANNEL = 2 // keyword allowed in per channel config
)
type ChanPlugCtxPtr *ChanPlugCtx
type ChanPlugCtx struct {
    Plugin  PluginPtr
    Config  PluginConfig        // interface{}
    Ctx     PluginConfigCtx     // interface{}
}

type ChannelPtr *Channel
type Channel struct {
    Name       string 
    Descr      string 
    Rules      []RulePtr
    Sinks      []ChannelPtr
    InPlugs    []ChanPlugCtxPtr
    OutPlugs   []ChanPlugCtxPtr
    ProcPlugs  []ChanPlugCtxPtr
}
type RulePtr *Rule
type Rule struct {
    RuleId       uint32 
    SrcChName    string
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

type RulePartFunc func(args []RulePartArg, event map[string]interface{}) interface{}

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
