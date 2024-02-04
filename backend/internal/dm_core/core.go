package dm_core

import (
	"fmt"
    "os"
	"regexp"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
    "dagproc/internal/di"
)

var (
    mLog zerolog.Logger
    channels map[string]di.Channel
    templates map[string]di.Template
)

func ModInit() (di.ModHookTable, error) {
    mLog = log.With().Str("module", "core").Caller().Logger()
    mLog.Debug().Msg("ModInit")
    return di.ModHookTable{ 
        LoadConfigHook:     coreLoadConfig,
        ReceiveEventHook:   nil, 
        SendEventHook:      nil,
        ProcessEventHook:   coreProcessEvent, 
    }, nil
}

func coreLoadConfig(config di.CFConfig) (di.PluginConfig, error) {
    // No config for the core right now
    mLog.Debug().Msg("coreLoadConfig")
    return make(map[string]interface{}), nil
}

func coreProcessEvent (ev di.Event) error {
    mLog.Debug().Msg("coreProcessEvent")
    return nil
}

func ruleFuncTrue(args []di.RulePartArg, event di.Event) interface{} {
    return true
}

func ruleFuncFalse(args []di.RulePartArg, event di.Event) interface{} {
    return false
}

func ruleFuncValue(args []di.RulePartArg, event di.Event) interface{} {
    if len(args) != 1 {
        mLog.Printf("value() takes one argument, %d given", len(args))
        return nil
    }
    switch (args[0].ArgType) {
        case di.PartArgTypeStr, di.PartArgTypeNum, di.PartArgTypeSet:
            return args[0].ArgValue
        case di.PartArgTypeKey:
            return EventGetKeyValue(event, args[0].ArgValue.(string))
        default:
            mLog.Printf("value() takes string, number, key or set argument, %d given", args[0].ArgType)
    }
    return nil
}

func ChannelGetBool(arg di.RulePartArg, event di.Event) bool {
    var argVal interface{}
    if (arg.ArgType != di.PartArgTypeFunc) {
        mLog.Printf("Argument %+v should be a function for ChannelGetBool", arg)
        return false
    }
    var part di.RulePart
    part = arg.ArgValue.(di.RulePart)
    argVal = part.Function(part.Arguments, event)
    if (argVal == nil) {
        return false
    }
    switch argVal.(type) {
        case bool:
            return argVal.(bool)
        case string:
            var argStr string
            argStr = argVal.(string)
            if (argStr == "true") || (argStr == "1") || (argStr == "True") || (argStr == "TRUE") {
                return true;
            }
            return false
        case float64:
            return argVal.(bool)
    }
    mLog.Printf("Cannot convert arg %+v to boolean, returning true", argVal)
    return true
}

func ChannelGetNum(arg di.RulePartArg, event di.Event) float64 {
    var argVal interface{}
    if (arg.ArgType != di.PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetNum", arg)
        return 0
    }
    var part di.RulePart
    part = arg.ArgValue.(di.RulePart)
    argVal = part.Function(part.Arguments, event)
    if (argVal == nil) {
        return 0
    }
    switch argVal.(type) {
        case bool, float64:
            return argVal.(float64)
        case string:
            var err error
            var ret float64
            ret, err = strconv.ParseFloat(argVal.(string), 64)
            if (err != nil) {
                log.Printf("Error %s converting %s to number", err,  argVal.(string))
                return 0
            }
            return ret
    }
    log.Printf("Cannot convert arg %+v to number, returning 0", argVal)
    return 0
}

func ChannelGetStr(arg di.RulePartArg, event di.Event) string {
    var argVal interface{}
    if (arg.ArgType != di.PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetStr", arg)
        return ""
    }
    var part di.RulePart
    part = arg.ArgValue.(di.RulePart)
    argVal = part.Function(part.Arguments, event)
    if (argVal == nil) {
        return ""
    }
    switch argVal.(type) {
        case bool:
            return fmt.Sprintf("%t", argVal.(bool))
        case float64:
            return fmt.Sprintf("%f", argVal.(float64))
        case string:
            return argVal.(string)
    }
    log.Printf("Cannot convert arg %+v to string, returning empty string", argVal)
    return ""
}

func ChannelGetRaw(arg di.RulePartArg, event di.Event) interface{} {
    if (arg.ArgType != di.PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetRaw", arg)
        return ""
    }
    var part di.RulePart
    part = arg.ArgValue.(di.RulePart)
    return part.Function(part.Arguments, event)
}

func channelCheckArgs(args []di.RulePartArg, numArgs int, argTypes ...di.RulePartArgType) bool {
    var goodCnt int
    goodCnt = 0
    if (len(args) != numArgs) {
        log.Printf("Must be %d args, %d given", numArgs, len(args))
        return false
    }
    for argPos, argType := range argTypes {
        if (args[argPos].ArgType != argType) {
            log.Printf("Argument %d has type %d, expected %d", argPos, args[argPos].ArgType, argType);
            break;
        }
        goodCnt++;
    }
    return goodCnt == len(args);
}

func ruleFuncEq(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], event) ==  ChannelGetNum(args[1], event)
}

func ruleFuncNe(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], event) !=  ChannelGetNum(args[1], event)
}

func ruleFuncLt(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], event) <  ChannelGetNum(args[1], event)
}

func ruleFuncLe(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], event) <= ChannelGetNum(args[1], event)
}

func ruleFuncGt(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], event) >  ChannelGetNum(args[1], event)
}

func ruleFuncGe(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], event) >= ChannelGetNum(args[1], event)
}

func ruleFuncAnd(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetBool(args[0], event) && ChannelGetBool(args[1], event)
}

func ruleFuncOr(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetBool(args[0], event) || ChannelGetBool(args[1], event)
}

func ruleFuncNot(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 1, di.PartArgTypeFunc)) {
        return false;
    }
    return !ChannelGetBool(args[0], event)
}

func ruleFuncRegex(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    var pattern, s string
    pattern = ChannelGetStr(args[0], event) // Pattern
    s = ChannelGetStr(args[1], event) // String to match against the pattern
    var err error
    var ret bool
    ret, err = regexp.MatchString(pattern, s)
    if (err != nil) {
        log.Printf("Error matching %s against %s", pattern, s)
        return false
    }
    return ret
}

func ruleFuncHas(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    var slice interface{}
    var item interface{}
    slice = ChannelGetRaw(args[0], event)
    item = ChannelGetRaw(args[1], event)
    switch slice.(type) {
        case []interface{}:
            for idx, _ := range slice.([]interface{}) {
                if (slice.([]interface{})[idx] == item) {
                    return true;
                }
            }
        case map[string]interface{}:
            for idx, _ := range slice.(map[string]interface{}) {
                if (slice.(map[string]interface{})[idx] == item) {
                    return true;
                }
            }
    }
    return false
}

func ruleFuncSince(args []di.RulePartArg, event di.Event) interface{} {
    var err error
  /* "Application level" function */
    if (!channelCheckArgs(args, 1, di.PartArgTypeFunc)) {
        return false;
    }
    var rawDate interface{}
    var parsedDate time.Time
    rawDate = ChannelGetRaw(args[0], event)
    switch rawDate.(type) {
        case string:
            /* assuming Go-style date RFC3339Nano */
            parsedDate, err = time.Parse(time.RFC3339Nano, rawDate.(string));
            if (err != nil) {
                log.Printf("Unable to parse %s as RFC3339Nano date", rawDate.(string))
                return 0
            }
        case float64:
            /* assuming unix timestamp */
            parsedDate = time.Unix(int64(rawDate.(float64)), 0);
        default:
            log.Printf("Can not process %+v as date", rawDate)
            return 0
    }
    return time.Since(parsedDate)
}

func ruleFuncJoin(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    var collection interface{}
    var joinSym string
    var temp []string

    collection = ChannelGetRaw(args[0], event)
    joinSym = ChannelGetStr(args[1], event)
    switch collection.(type) {
        case []interface{}:
            temp = make([]string, len(collection.([]interface{})))
            for idx, item := range collection.([]interface{}) {
                temp[idx] = fmt.Sprintf("%v", item)
            }
        case map[string]interface{}:
            temp = make([]string, len(collection.(map[string]interface{})))
            var idx = 0
            for _, item := range collection.(map[string]interface{}) {
                temp[idx] = fmt.Sprintf("%v", item)
                idx++
            }
        default:
            temp = make([]string, 1)
            temp[0] = fmt.Sprintf("%v", collection)
    }
    return strings.Join(temp, joinSym)
}

func ruleFuncHasany(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    var slice interface{}
    var itemSet []di.RulePartArg
    itemSet = ChannelGetRaw(args[0], event).([]di.RulePartArg)
    slice = ChannelGetRaw(args[1], event)
    log.Printf("Itemset is %v", itemSet)
    for _, item := range itemSet {
	    switch slice.(type) {
	        case []interface{}:
	            for idx, _ := range slice.([]interface{}) {
	                if (slice.([]interface{})[idx] == item.ArgValue) {
	                    return true;
	                }
	            }
	        case map[string]interface{}:
	            for idx, _ := range slice.(map[string]interface{}) {
	                if (slice.(map[string]interface{})[idx] == item.ArgValue) {
	                    return true;
	                }
	            }
	    }
    }
    return false;
}

func ruleFuncHasall(args []di.RulePartArg, event di.Event) interface{} {
    if (!channelCheckArgs(args, 2, di.PartArgTypeFunc, di.PartArgTypeFunc)) {
        return false;
    }
    var slice interface{}
    var itemSet []di.RulePartArg
    itemSet = ChannelGetRaw(args[0], event).([]di.RulePartArg)
    slice = ChannelGetRaw(args[1], event)
    var itemCount, matchCount int = 0, 0
    for _, item := range itemSet {
    	itemCount++
	    switch slice.(type) {
	        case []interface{}:
	            for idx, _ := range slice.([]interface{}) {
	                if (slice.([]interface{})[idx] == item.ArgValue) {
	                    matchCount++
	                    break
	                }
	            }
	        case map[string]interface{}:
	            for idx, _ := range slice.(map[string]interface{}) {
	                if (slice.(map[string]interface{})[idx] == item.ArgValue) {
	                    matchCount++
	                    break
	                }
	            }
	    }
    }
    return itemCount == matchCount
}

var PnFuncDispatcher map[string]di.PnFunc = map[string]di.PnFunc {
  "value":  { 1, ruleFuncValue },
  "true":	{ 0, ruleFuncTrue },
  "false":  { 0, ruleFuncFalse},
  "eq": 	{ 2, ruleFuncEq },
  "ne": 	{ 2, ruleFuncNe },
  "lt": 	{ 2, ruleFuncLt },
  "le": 	{ 2, ruleFuncLe },
  "gt": 	{ 2, ruleFuncGt },
  "ge": 	{ 2, ruleFuncGe },
  "and": 	{ 2, ruleFuncAnd },
  "or": 	{ 2, ruleFuncOr },
  "not": 	{ 1, ruleFuncNot },
  "regex": 	{ 2, ruleFuncRegex },
  "has": 	{ 2, ruleFuncHas },
  "hasany":	{ 2, ruleFuncHasany },
  "hasall":	{ 2, ruleFuncHasall },
  "since": 	{ 1, ruleFuncSince },
  "join":   { 2, ruleFuncJoin },
}

func channelRuleParserRuneCB(rRule []rune, rPos int, currState di.RuleParserStateType) (bool, di.RuleParserStateType) {
    /* returns whether or not to copy current rune to part buffer and new parser state */
    const validKeySymbols string = "._0123456789abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    const validKeyFirstSymbol string = "."
    const validFuncSymbols string = "_0123456789abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    const validFuncFirstSymbol string = "_abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    const validNumSymbols string = "0123456789abcdefxABCDEFX-"
    const validNumFirstSymbol string = "0123456789-"
    if (rPos >= len(rRule)) {
        return false, di.RuleParserAfterArg;
    }

    var runeStr string = string(string(rRule[rPos]))

    switch (currState & di.RULE_PARSER_MF_MASK) {
        case di.RuleParserLookingForArg:
            if unicode.IsSpace(rRule[rPos]) {
                return false, di.RuleParserLookingForArg
            }
            if ((rPos > 0) && !unicode.IsSpace(rRule[rPos-1])) {
                /* We were not at the beginning of the line, and previous symbol was not a space
                      (meaning tokens were not delimited by at least one space)
                */
                log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
                return false, di.RuleParserInvalidState
            }
            if (rRule[rPos] == '"') {
                /* Opening quote of a string argument */
                return false, (di.RuleParserInArgStr | di.RULE_PARSER_MF_IN_QUOTE)
            }
            if (rRule[rPos] == '[') {
                /* Set */
                return false, di.RuleParserInArgSet
            }
            if (strings.ContainsAny(runeStr, validKeyFirstSymbol)) {
                /* Key */
                return true, di.RuleParserInArgKey
            }
            if (strings.ContainsAny(runeStr,  validNumFirstSymbol)) {
                /* Number */
                return true, di.RuleParserInArgNum
            }
            if (strings.ContainsAny(runeStr, validFuncFirstSymbol)) {
                /* Function */
                return true, di.RuleParserInArgFunc
            }
            /* We don't know what's there, bailing out */
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, di.RuleParserInvalidState
        case di.RuleParserInArgFunc:
            if unicode.IsSpace(rRule[rPos]) {
                return false, di.RuleParserAfterArg
            }
            if strings.ContainsAny(runeStr, validFuncSymbols) {
                return true, di.RuleParserInArgFunc
            }
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, di.RuleParserInvalidState

        case di.RuleParserInArgKey:
            if unicode.IsSpace(rRule[rPos]) {
                return false, di.RuleParserAfterArg
            }
            if strings.ContainsAny(runeStr, validKeySymbols) {
                return true, di.RuleParserInArgKey
            }
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, di.RuleParserInvalidState

        case di.RuleParserInArgNum:
            if unicode.IsSpace(rRule[rPos]) {
                return false, di.RuleParserAfterArg
            }
            if strings.ContainsAny(runeStr, validNumSymbols) {
                return true, di.RuleParserInArgNum
            }
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, di.RuleParserInvalidState
        case di.RuleParserInArgStr: // Somewhat different
            if (rRule[rPos] == '"') {
                if ((currState & di.RULE_PARSER_MF_IN_BQ) == di.RULE_PARSER_MF_IN_BQ) {
                    return true, (currState &^ di.RULE_PARSER_MF_IN_BQ)
                }
                return false, di.RuleParserAfterArg
            }
            if (rRule[rPos] == '\\') { // Skipping backquote
                if ((currState & di.RULE_PARSER_MF_IN_BQ) == di.RULE_PARSER_MF_IN_BQ ) {
                    return true, (currState &^ di.RULE_PARSER_MF_IN_BQ)
                }
                return false, currState | di.RULE_PARSER_MF_IN_BQ
            }
            return true, currState
        case di.RuleParserInArgSet: // It's a very special case, we'll parse set argument later, so we need to keep the string intact
        	if (rRule[rPos] == '"') {
        		if ((currState & di.RULE_PARSER_MF_IN_BQ) == di.RULE_PARSER_MF_IN_BQ) {
        			return true, (currState &^ di.RULE_PARSER_MF_IN_BQ)
        		}
        		if ((currState & di.RULE_PARSER_MF_IN_QUOTE) == di.RULE_PARSER_MF_IN_QUOTE) {
        			return true, (currState &^ di.RULE_PARSER_MF_IN_QUOTE)
        		}
        		return true, currState | di.RULE_PARSER_MF_IN_QUOTE
        	}
            if (rRule[rPos] == ']') { // brackets are allowed inside strings
            	if ((currState & di.RULE_PARSER_MF_IN_QUOTE) == di.RULE_PARSER_MF_IN_QUOTE) {
            		return true, currState &^ di.RULE_PARSER_MF_IN_BQ // we are still inside a string, but definitely should clear backquote flag
            	}
                return false, di.RuleParserAfterArg
            }
			if (rRule[rPos] == '[') {
            	if ((currState & di.RULE_PARSER_MF_IN_QUOTE) == di.RULE_PARSER_MF_IN_QUOTE) {
            		return true, currState &^ di.RULE_PARSER_MF_IN_BQ
            	}
            	log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
                return false, di.RuleParserInvalidState
            }

            if (rRule[rPos] == '\\') { // Not skipping backquote, but accounting for it
                if ((currState & di.RULE_PARSER_MF_IN_BQ) == di.RULE_PARSER_MF_IN_BQ ) {
                    return true, (currState &^ di.RULE_PARSER_MF_IN_BQ)
                }
                return true, currState | di.RULE_PARSER_MF_IN_BQ
            }
            return true, currState
    }
    log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
    return false, di.RuleParserInvalidState
}
func channelParseRules(chDef *di.Channel ) error {
    var err error
    var ok bool
    /*
      So, Polish notation - function and variable number of arguments (currently one or two)
      arguments could be functions (looked up in PnFuncDispatcher maps) , key/field names (first character is a dot),
      numbers (first character is a decimal number 0-9 or a minus sign) or strings (first character is a single or double quote),
       to check if event field labels["projects"] contains element "backend", we'll write
        has .labels.projects "backend"

       compound rules will look like

       and has .labels.projects "backend" or eq .labels.severity "warning" eq .labels.severity "crit"  -
       labels["project"] eq "backend" and (labels.severity eq "warning" or labels.severity eq "crit")

       and or eq .labels.severity "warning" eq .labels.severity "crit" gt since .labels.begin 86400 -
    (labels.severity eq "warning" or labels.severity eq "crit") and since labels["begin"] > 86400

    easier to parse , similar to text/template boolean conditions
  */
    var chRule di.RulePtr
    var chRuleIdx int

    for chRuleIdx, chRule = range chDef.Rules {
        var rRule, partBuf []rune
        var currRuleParserState, newRuleParserState di.RuleParserStateType
        var rPos, partBufIdx int
        rRule = []rune(chRule.RuleStr)
        rPos = 0
        currRuleParserState = di.RuleParserLookingForArg
        partBuf = make([]rune, len(rRule))

        var copyToBuf bool
        var parserStack []di.RuleParserCtx
        var parserSP int

        partBufIdx = 0
        parserSP = 0
        parserStack = make([]di.RuleParserCtx, 0, 10)
        parserStack = append(parserStack, di.RuleParserCtx{&chRule.Root, 0})

        for (	(rPos <= len(rRule)) &&
                  (parserSP >= 0) &&
                  (err == nil) &&
                  (currRuleParserState != di.RuleParserInvalidState)) {
            var pCtx di.RuleParserCtx
            pCtx = parserStack[parserSP]
            copyToBuf, newRuleParserState = channelRuleParserRuneCB(rRule, rPos, currRuleParserState)
            switch (newRuleParserState & di.RULE_PARSER_MF_MASK) { // Sometimes higher bits (quote , backquote ) will be set here, but only in the RuleParserInArgXXX case
                case di.RuleParserLookingForArg:
                    partBufIdx = 0						// Just skipping the whitespace
                case di.RuleParserInArgFunc, di.RuleParserInArgKey,di. RuleParserInArgNum, di.RuleParserInArgStr, di.RuleParserInArgSet:
                    if (copyToBuf) {	// we may want to skip copying some characters (opening/closing and escaped quotes)
                        partBuf[partBufIdx] = rRule[rPos]
                        partBufIdx++
                    }
                case di.RuleParserAfterArg:
                    var argStr, funcName string
                    var hasArg bool
                    var argFunc di.PnFunc
                    hasArg = false
                    if (partBufIdx == 0) {
                        break; // End of line
                    }
                    argStr = string(partBuf[:partBufIdx])
					currRuleParserState = currRuleParserState & di.RULE_PARSER_MF_MASK // Clearing all modifiers

                    if (currRuleParserState == di.RuleParserInArgFunc) {
                        funcName = argStr
                    } else {
                        funcName = "value"
                        hasArg = true
                    }
                    if argFunc,ok = PnFuncDispatcher[funcName]; ok {
                        if (pCtx.PartPtr.Function == nil) {
                            pCtx.PartPtr.Function = argFunc.Function
                            pCtx.PartPtr.Arguments = make([]di.RulePartArg, argFunc.NumArgs)
                            pCtx.CurrArg = 0
                            parserStack[parserSP] = pCtx
                        } else {
                            var newPart di.RulePart
                            newPart = di.RulePart{Function: argFunc.Function, Arguments: make([]di.RulePartArg, argFunc.NumArgs)}
                            pCtx.PartPtr.Arguments[pCtx.CurrArg] = di.RulePartArg{di.PartArgTypeFunc, newPart}
                            pCtx.CurrArg++
                            parserStack[parserSP] = pCtx
                            parserStack = append(parserStack, di.RuleParserCtx{&newPart, 0})
                            parserSP++
                            pCtx = parserStack[parserSP]
                        }
                        if (hasArg) {
                            /*
                              We can convert between RuleParserState and di.RulePartArgType due to
                               carefully chosen values of both enums
                            */
                            if (currRuleParserState == di.RuleParserInArgSet) {
                            	pCtx.PartPtr.Arguments[pCtx.CurrArg] = di.RulePartArg{di.RulePartArgType(currRuleParserState), channelParseSet(argStr)}
                            } else {
                            	pCtx.PartPtr.Arguments[pCtx.CurrArg] = di.RulePartArg{di.RulePartArgType(currRuleParserState), argStr}
                            }
                            pCtx.CurrArg++
                            parserStack[parserSP] = pCtx
                        }
                    } else {
                        err = errors.New(fmt.Sprintf("%s: unknown function", argStr))
                        log.Printf("Error '%s' parsing rule %s", err, chRule.RuleStr)
                    }
                    newRuleParserState = di.RuleParserLookingForArg
                    partBufIdx = 0
                default:
                    err = errors.New(fmt.Sprintf("Invalid new parser state %d", newRuleParserState))
            } // End of a state machine switch
            if (newRuleParserState == di.RuleParserLookingForArg) {
              for parserSP >= 0 {
                  pCtx = parserStack[parserSP]
                  if (pCtx.CurrArg >= len(pCtx.PartPtr.Arguments)) { // We're done with current argument, let's pop this context from the stack
                      parserSP--;
                      parserStack = parserStack[:len(parserStack) - 1]
                  } else {
                      /* Stop unwinding stack */
                      break
                  }
              }
            }
            currRuleParserState = newRuleParserState
            rPos++;
        }
        if (err != nil) {
          return err
        }
        chDef.Rules[chRuleIdx] = chRule
    } // for rule in chDef.Rules
    log.Printf("Parsed channel: %+v", chDef)
    return err
}

func channelParseSet(setStr string) []di.RulePartArg {
	/* extremely simplified parser from channelParseRules . Every fancy feature removed, just basic tokenization */
	var err error
	var ret []di.RulePartArg
	var rRule, partBuf []rune
    var currRuleParserState, newRuleParserState di.RuleParserStateType
    var rPos, partBufIdx int
    var copyToBuf bool
   	rRule = []rune(setStr)
	rPos = 0
	currRuleParserState = di.RuleParserLookingForArg
	partBuf = make([]rune, len(rRule))
	ret = make([]di.RulePartArg, 0, 2)
	for (rPos <= len(rRule) && (err == nil) && (currRuleParserState != di.RuleParserInvalidState)) {
    	copyToBuf, newRuleParserState = channelRuleParserRuneCB(rRule, rPos, currRuleParserState)
    	switch (newRuleParserState & di.RULE_PARSER_MF_MASK) {
			case di.RuleParserLookingForArg:
            	partBufIdx = 0
            case di.RuleParserInArgNum, di.RuleParserInArgStr:
                if (copyToBuf) { // we may want to skip copying some characters (opening/closing and escaped quotes)
                    partBuf[partBufIdx] = rRule[rPos]
                    partBufIdx++
                }
            case di.RuleParserAfterArg:
                if (partBufIdx == 0) {
					break
                }
                var argStr string
				argStr = string(partBuf[:partBufIdx])
				currRuleParserState = currRuleParserState & di.RULE_PARSER_MF_MASK // Clearing all modifiers
				if (currRuleParserState == di.RuleParserInArgNum) {
            		/* this is a number */
            		var num float64
            		num, err = strconv.ParseFloat(argStr, 64)
            		ret = append(ret, di.RulePartArg{di.RulePartArgType(currRuleParserState), num})
            	} else {
            		/* string goes unmodified */
            		ret = append(ret, di.RulePartArg{di.RulePartArgType(currRuleParserState), argStr})
            	}
            	newRuleParserState = di.RuleParserLookingForArg
                partBufIdx = 0
            default:
                err = errors.New(fmt.Sprintf("Invalid new parser state %d while parsing set %s", newRuleParserState, setStr))
        }
        currRuleParserState = newRuleParserState
       	rPos++
	}
	return ret
}

func channelAddSinkId(srcChDef di.ChannelPtr, sinkId uint32) (error) {
    var err error

    mLog.Fatal().Msgf("FIXME")
    os.Exit(1)
    /*
    if srcChDef.Sinks == nil {
        srcChDef.Sinks = make([]di.ChannelPtr, 0, 2)
    }
    if (srcChDef.Id == sinkId) {
        log.Printf("Channel %d can't be a sink for itself", srcChDef.Id)
    } else {
        srcChDef.Sinks = append(srcChDef.Sinks , sinkId)
    }*/
    return err
}

func channelParseSrcChIds(srcChIds interface{}) ([]uint32) {
    mLog.Fatal().Msgf("FIXME")
    os.Exit(1)
/*
    var ret []uint32;
    switch  srcChIds.(type) {
        case float64: // single number
            ret = make([]uint32, 0, 1)
            ret[0] = srcChIds.(uint32)
        case []interface{}:
            ret = make([]uint32, 0, len(srcChIds.([]interface{})))
            for _, chId := range srcChIds.([]interface{}) {
                if (chId.(uint32) != 0) {
                    ret = append(ret, chId.(uint32))
                } else {
                    log.Printf("Invalid source channel id %v", chId);
                }
            }
        default:
            log.Printf("Invalid source channel ids %v", srcChIds);
    }
    return ret
*/
    return nil
}

func channelPipeSrcsToSinks(channelDefs []di.ChannelPtr, lastChannelId uint32) {
    mLog.Fatal().Msgf("FIXME")
    os.Exit(1)
    /*
    var chDef di.ChannelPtr;
    var chRule di.RulePtr
    for _, chDef = range channelDefs {
        for _, chRule = range chDef.Rules {
            if (chRule.Root.Function != nil) { // Rule was successfully parsed
                var srcChIds []uint32
                srcChIds = channelParseSrcChIds(chRule.SrcChId);
                for _, srcChId := range(srcChIds) {
                    if ((srcChId <= lastChannelId) && (channelDefs[srcChId] != nil)) {
                        channelAddSinkId(channelDefs[srcChId], chDef.Id)// almost placeholder at this point
                                                                        // will check for acyclicity later
                    } else {
                        log.Printf("Unknown channel id %d", srcChId)
                    }
                }
            }
        }
    }
    */
}


func channelMatchEvent(channel di.ChannelPtr, srcChId uint32, event di.Event) bool {
    mLog.Fatal().Msgf("FIXME")
    os.Exit(1)
    /*
    if (channel.Id == 0) {
        return true // special case
    }
    var chRule di.RulePtr
    var ret bool
    ret = false
    for _, chRule = range channel.Rules {
        if (chRule.SrcChId == srcChId) {
            var rootRule di.RulePartArg
            rootRule = di.RulePartArg{ArgType: di.PartArgTypeFunc, ArgValue: chRule.Root}
            ret = ChannelGetBool(rootRule, event)
            if ret {
                break
            }
        }
    }
    return ret
    */
    return false
}

func channelRunTheGauntlet(	channelDefs []di.ChannelPtr, initialChId uint32, srcChId uint32,
                            event di.Event, groupsToDeliver map[string]string, totalMatches uint32) (uint32) { // Everything is a pointer, God bless Go magic
    mLog.Fatal().Msgf("FIXME")
    os.Exit(1)
/*
    var channel di.ChannelPtr
    var match bool
    var err error
    channel = channelDefs[initialChId]
    match = channelMatchEvent(channel, srcChId, event)
    if (match == true) {
        totalMatches++
        if (len(channel.Group) > 0) {
            var tpOutput bytes.Buffer;
            var tp *template.Template;
            if (channel.MsgTemplate != nil) {
                tp = channel.MsgTemplate;
            } else {
                tp = channelDefs[0].MsgTemplate;
            };
            err = tp.Execute(&tpOutput, event)
            if (err != nil) {
                log.Printf("Error %s rendering template", err)
            }
            groupsToDeliver[channel.Group] = tpOutput.String()
        }
        var sinkId uint32
        for _, sinkId = range channel.Sinks {
            totalMatches += channelRunTheGauntlet(channelDefs, sinkId, initialChId, event, groupsToDeliver, totalMatches)
        }
    }
    return totalMatches
    */
    return 0
}
