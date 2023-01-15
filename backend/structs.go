package main

import (
    "sync"
    "text/template"
)

type Config struct {
    XmppServer  string `json:"xmppServer"`
    JidLogin    string `json:"jidLogin"`
    JidPwd      string `json:"jidPwd"`
    JidNick     string `json:jidNick"`
    RedisURI    string `json:"redisURI"`
    GroupsURI   string `json:"groupsURI"`
}


type ServiceConfig struct {
    ChannelDefVer   uint32 `redis:"channel_version"`
    LastChannelId   uint32 `redis:"last_channel_id"`
    ModList         string `redis:"modules"`
    IsShutdown      bool   `redis:"shutdown"`
}

type PnFunc struct {
  numArgs   int
  function  RulePartFunc
}

type RuleParserCtx struct {
  partPtr *RulePart
  currArg  int
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
    argType    RulePartArgType
    argValue   interface{}
}

type RulePartFunc func(args []RulePartArg, alert Alert) interface{}

type RulePart struct {
    function     RulePartFunc
    arguments  []RulePartArg
}

type ChannelRule struct {
    RuleId       uint32 `json:"id"`
    SrcChId      interface{} `json:"src"`
    RuleStr      string `json:"cond"`
    CondLink     string `json:"condfrom"`
    Root         RulePart
}

type ChannelDef struct {
    Id              uint32 `json:"id"`
    Version         uint32 `json:"version"`
    Label           string `json:"label"`
    Group           string `json:"group"`
    MsgFormat       string `json:"format"`
    MsgTemplate    *template.Template
    Rules         []ChannelRule `json:"rules"`
    Sinks         []uint32 
}

type Alert map[string]interface{}

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

type XmppMsg struct {
    XmppGroup   string
    Message     string
}

const JOIN_STATUS_NOT_JOINED = 0
const JOIN_STATUS_JOINING = 1
const JOIN_STATUS_JOINED = 2
type XmppGroup struct {
    JoinStatus    uint32
    GroupName     string
    GroupServer   string
    GroupNick     string
    Deferred    []string
    GroupLock  sync.Mutex
}
