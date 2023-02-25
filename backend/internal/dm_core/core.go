package dm_core

import (
	//	"encoding/json"
	//	"flag"
	"fmt"
	//	"io/ioutil"
	//	"os"
	"regexp"
	//	"sync"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/gomodule/redigo/redis"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
    coreConfig CoreConfig
    mLog zerolog.Logger
)
func ModInit() (ModDispTable, error) {
    mLog = log.With().Str("module", "core").Logger()
    mLog.Debug().Msg("ModInit")
    return ModDispTable{ LoadConfig: loadCoreConfig }, nil
}

func loadCoreConfig(values []interface{}) error {
    var ret error
    if (len(values) == 0) {
        ret = errors.New("Empty config")
    } else {
        ret = redis.ScanStruct(values, &coreConfig)
        if (ret == nil) {
            mLog.Debug().Msgf("Channel definitions version: %d, last channel id: %d", coreConfig.ChannelDefVer, coreConfig.LastChannelId)
            if (coreConfig.ChannelDefVer == 0) {
                ret = errors.New("Invalid config")
            } else {
            }
            // load variable number of channels
        }
    }
    return ret
}

func channelGetKeyValue(alert map[string]interface{}, key string) interface{} {
    var err error
    var path []string
    path = strings.Split(key, ".")

    if ((len(path) <= 2) || (len(path[0]) > 0)) {
        // Second condition is extraneous, it checks if  key starts with a dot. But let's stay safe
        log.Printf("Invalid key %s", key)
        return nil
    }
    var pathIdx int
    var cursor interface{}
    cursor = alert
    for pathIdx = 1; pathIdx < len(path); pathIdx++ {
        var currPart string
        currPart = path[pathIdx]
        if (len(currPart) == 0) {
            log.Printf("Zero length part in key %s", key)
            return nil
        }
        switch cursor.(type) { // since this data is unmarshaled by json, we know we have a limited set of possible types
            case []interface{}: // array, path[pathIdx] should contain array index

                var idx int
                idx,err = strconv.Atoi(currPart)
                if ((err != nil) || (idx < 0 ) || (idx >= len(cursor.([]interface{})))) {
                    log.Printf("Invalid index %d, full key %s, alert: %+v, cursor: %+v", idx, key, alert, cursor)
                    return nil
                }
                cursor = cursor.([]interface{})[idx]
            case map[string]interface{}: // map, path[pathIdx] should contain map key
                var ok bool
                if cursor, ok = cursor.(map[string]interface{})[currPart]; ok == false {
                    log.Printf("No %s found , full key %s, alert: %+v, cursor: %+v", currPart, key, alert, cursor)
                    return nil
                }
            default:
                log.Printf("Non-indexable member %s found , full key %s, alert: %+v, cursor: %+v", currPart, key, alert, cursor)
                return nil
        }
    }
    return cursor
}

func ruleFuncTrue(args []RulePartArg, alert Alert) interface{} {
    return true
}

func ruleFuncFalse(args []RulePartArg, alert Alert) interface{} {
    return false
}

func ruleFuncValue(args []RulePartArg, alert Alert) interface{} {
    if len(args) != 1 {
        log.Printf("value() takes one argument, %d given", len(args))
        return nil
    }
    switch (args[0].argType) {
        case PartArgTypeStr, PartArgTypeNum, PartArgTypeSet:
            return args[0].argValue
        case PartArgTypeKey:
            return channelGetKeyValue(alert, args[0].argValue.(string))
        default:
            log.Printf("value() takes string, number, key or set argument, %d given", args[0].argType)
    }
    return nil
}

func ChannelGetBool(arg RulePartArg, alert Alert) bool {
    var argVal interface{}
    if (arg.argType != PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetBool", arg)
        return false
    }
    var part RulePart
    part = arg.argValue.(RulePart)
    argVal = part.function(part.arguments, alert)
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
    log.Printf("Cannot convert arg %+v to boolean, returning true", argVal)
    return true
}

func ChannelGetNum(arg RulePartArg, alert Alert) float64 {
    var argVal interface{}
    if (arg.argType != PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetNum", arg)
        return 0
    }
    var part RulePart
    part = arg.argValue.(RulePart)
    argVal = part.function(part.arguments, alert)
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

func ChannelGetStr(arg RulePartArg, alert Alert) string {
    var argVal interface{}
    if (arg.argType != PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetStr", arg)
        return ""
    }
    var part RulePart
    part = arg.argValue.(RulePart)
    argVal = part.function(part.arguments, alert)
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

func ChannelGetRaw(arg RulePartArg, alert Alert) interface{} {
    if (arg.argType != PartArgTypeFunc) {
        log.Printf("Argument %+v should be a function for ChannelGetRaw", arg)
        return ""
    }
    var part RulePart
    part = arg.argValue.(RulePart)
    return part.function(part.arguments, alert)
}

func channelCheckArgs(args []RulePartArg, numArgs int, argTypes ...RulePartArgType) bool {
    var goodCnt int
    goodCnt = 0
    if (len(args) != numArgs) {
        log.Printf("Must be %d args, %d given", numArgs, len(args))
        return false
    }
    for argPos, argType := range argTypes {
        if (args[argPos].argType != argType) {
            log.Printf("Argument %d has type %d, expected %d", argPos, args[argPos].argType, argType);
            break;
        }
        goodCnt++;
    }
    return goodCnt == len(args);
}

func ruleFuncEq(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], alert) ==  ChannelGetNum(args[1], alert)
}

func ruleFuncNe(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], alert) !=  ChannelGetNum(args[1], alert)
}

func ruleFuncLt(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], alert) <  ChannelGetNum(args[1], alert)
}

func ruleFuncLe(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], alert) <= ChannelGetNum(args[1], alert)
}

func ruleFuncGt(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], alert) >  ChannelGetNum(args[1], alert)
}

func ruleFuncGe(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetNum(args[0], alert) >= ChannelGetNum(args[1], alert)
}

func ruleFuncAnd(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetBool(args[0], alert) && ChannelGetBool(args[1], alert)
}

func ruleFuncOr(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    return ChannelGetBool(args[0], alert) || ChannelGetBool(args[1], alert)
}

func ruleFuncNot(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 1, PartArgTypeFunc)) {
        return false;
    }
    return !ChannelGetBool(args[0], alert)
}

func ruleFuncRegex(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    var pattern, s string
    pattern = ChannelGetStr(args[0], alert) // Pattern
    s = ChannelGetStr(args[1], alert) // String to match against the pattern
    var err error
    var ret bool
    ret, err = regexp.MatchString(pattern, s)
    if (err != nil) {
        log.Printf("Error matching %s against %s", pattern, s)
        return false
    }
    return ret
}

func ruleFuncHas(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    var slice interface{}
    var item interface{}
    slice = ChannelGetRaw(args[0], alert)
    item = ChannelGetRaw(args[1], alert)
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

func ruleFuncSince(args []RulePartArg, alert Alert) interface{} {
    var err error
  /* "Application level" function */
    if (!channelCheckArgs(args, 1, PartArgTypeFunc)) {
        return false;
    }
    var rawDate interface{}
    var parsedDate time.Time
    rawDate = ChannelGetRaw(args[0], alert)
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

func ruleFuncJoin(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    var collection interface{}
    var joinSym string
    var temp []string

    collection = ChannelGetRaw(args[0], alert)
    joinSym = ChannelGetStr(args[1], alert)
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

func ruleFuncHasany(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    var slice interface{}
    var itemSet []RulePartArg
    itemSet = ChannelGetRaw(args[0], alert).([]RulePartArg)
    slice = ChannelGetRaw(args[1], alert)
    log.Printf("Itemset is %v", itemSet)
    for _, item := range itemSet {
	    switch slice.(type) {
	        case []interface{}:
	            for idx, _ := range slice.([]interface{}) {
	                if (slice.([]interface{})[idx] == item.argValue) {
	                    return true;
	                }
	            }
	        case map[string]interface{}:
	            for idx, _ := range slice.(map[string]interface{}) {
	                if (slice.(map[string]interface{})[idx] == item.argValue) {
	                    return true;
	                }
	            }
	    }
    }
    return false;
}

func ruleFuncHasall(args []RulePartArg, alert Alert) interface{} {
    if (!channelCheckArgs(args, 2, PartArgTypeFunc, PartArgTypeFunc)) {
        return false;
    }
    var slice interface{}
    var itemSet []RulePartArg
    itemSet = ChannelGetRaw(args[0], alert).([]RulePartArg)
    slice = ChannelGetRaw(args[1], alert)
    var itemCount, matchCount int = 0, 0
    for _, item := range itemSet {
    	itemCount++
	    switch slice.(type) {
	        case []interface{}:
	            for idx, _ := range slice.([]interface{}) {
	                if (slice.([]interface{})[idx] == item.argValue) {
	                    matchCount++
	                    break
	                }
	            }
	        case map[string]interface{}:
	            for idx, _ := range slice.(map[string]interface{}) {
	                if (slice.(map[string]interface{})[idx] == item.argValue) {
	                    matchCount++
	                    break
	                }
	            }
	    }
    }
    return itemCount == matchCount
}

var PnFuncDispatcher map[string]PnFunc = map[string]PnFunc {
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

func channelRuleParserRuneCB(rRule []rune, rPos int, currState RuleParserStateType) (bool, RuleParserStateType) {
    /* returns whether or not to copy current rune to part buffer and new parser state */
    const validKeySymbols string = "._0123456789abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    const validKeyFirstSymbol string = "."
    const validFuncSymbols string = "_0123456789abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    const validFuncFirstSymbol string = "_abcdedfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    const validNumSymbols string = "0123456789abcdefxABCDEFX-"
    const validNumFirstSymbol string = "0123456789-"
    if (rPos >= len(rRule)) {
        return false, RuleParserAfterArg;
    }

    var runeStr string = string(string(rRule[rPos]))

    switch (currState & RULE_PARSER_MF_MASK) {
        case RuleParserLookingForArg:
            if unicode.IsSpace(rRule[rPos]) {
                return false, RuleParserLookingForArg
            }
            if ((rPos > 0) && !unicode.IsSpace(rRule[rPos-1])) {
                /* We were not at the beginning of the line, and previous symbol was not a space
                      (meaning tokens were not delimited by at least one space)
                */
                log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
                return false, RuleParserInvalidState
            }
            if (rRule[rPos] == '"') {
                /* Opening quote of a string argument */
                return false, (RuleParserInArgStr | RULE_PARSER_MF_IN_QUOTE)
            }
            if (rRule[rPos] == '[') {
                /* Set */
                return false, RuleParserInArgSet
            }
            if (strings.ContainsAny(runeStr, validKeyFirstSymbol)) {
                /* Key */
                return true, RuleParserInArgKey
            }
            if (strings.ContainsAny(runeStr,  validNumFirstSymbol)) {
                /* Number */
                return true, RuleParserInArgNum
            }
            if (strings.ContainsAny(runeStr, validFuncFirstSymbol)) {
                /* Function */
                return true, RuleParserInArgFunc
            }
            /* We don't know what's there, bailing out */
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, RuleParserInvalidState
        case RuleParserInArgFunc:
            if unicode.IsSpace(rRule[rPos]) {
                return false, RuleParserAfterArg
            }
            if strings.ContainsAny(runeStr, validFuncSymbols) {
                return true, RuleParserInArgFunc
            }
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, RuleParserInvalidState

        case RuleParserInArgKey:
            if unicode.IsSpace(rRule[rPos]) {
                return false, RuleParserAfterArg
            }
            if strings.ContainsAny(runeStr, validKeySymbols) {
                return true, RuleParserInArgKey
            }
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, RuleParserInvalidState

        case RuleParserInArgNum:
            if unicode.IsSpace(rRule[rPos]) {
                return false, RuleParserAfterArg
            }
            if strings.ContainsAny(runeStr, validNumSymbols) {
                return true, RuleParserInArgNum
            }
            log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
            return false, RuleParserInvalidState
        case RuleParserInArgStr: // Somewhat different
            if (rRule[rPos] == '"') {
                if ((currState & RULE_PARSER_MF_IN_BQ) == RULE_PARSER_MF_IN_BQ) {
                    return true, (currState &^ RULE_PARSER_MF_IN_BQ)
                }
                return false, RuleParserAfterArg
            }
            if (rRule[rPos] == '\\') { // Skipping backquote
                if ((currState & RULE_PARSER_MF_IN_BQ) == RULE_PARSER_MF_IN_BQ ) {
                    return true, (currState &^ RULE_PARSER_MF_IN_BQ)
                }
                return false, currState | RULE_PARSER_MF_IN_BQ
            }
            return true, currState
        case RuleParserInArgSet: // It's a very special case, we'll parse set argument later, so we need to keep the string intact
        	if (rRule[rPos] == '"') {
        		if ((currState & RULE_PARSER_MF_IN_BQ) == RULE_PARSER_MF_IN_BQ) {
        			return true, (currState &^ RULE_PARSER_MF_IN_BQ)
        		}
        		if ((currState & RULE_PARSER_MF_IN_QUOTE) == RULE_PARSER_MF_IN_QUOTE) {
        			return true, (currState &^ RULE_PARSER_MF_IN_QUOTE)
        		}
        		return true, currState | RULE_PARSER_MF_IN_QUOTE
        	}
            if (rRule[rPos] == ']') { // brackets are allowed inside strings
            	if ((currState & RULE_PARSER_MF_IN_QUOTE) == RULE_PARSER_MF_IN_QUOTE) {
            		return true, currState &^ RULE_PARSER_MF_IN_BQ // we are still inside a string, but definitely should clear backquote flag
            	}
                return false, RuleParserAfterArg
            }
			if (rRule[rPos] == '[') {
            	if ((currState & RULE_PARSER_MF_IN_QUOTE) == RULE_PARSER_MF_IN_QUOTE) {
            		return true, currState &^ RULE_PARSER_MF_IN_BQ
            	}
            	log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
                return false, RuleParserInvalidState
            }

            if (rRule[rPos] == '\\') { // Not skipping backquote, but accounting for it
                if ((currState & RULE_PARSER_MF_IN_BQ) == RULE_PARSER_MF_IN_BQ ) {
                    return true, (currState &^ RULE_PARSER_MF_IN_BQ)
                }
                return true, currState | RULE_PARSER_MF_IN_BQ
            }
            return true, currState
    }
    log.Printf("Error parsing: String %s, position %d", string(rRule), rPos)
    return false, RuleParserInvalidState
}
func channelParseRules(chDef *ChannelDef ) error {
    var err error
    var ok bool
    /*
      So, Polish notation - function and variable number of arguments (currently one or two)
      arguments could be functions (looked up in PnFuncDispatcher maps) , key/field names (first character is a dot),
      numbers (first character is a decimal number 0-9 or a minus sign) or strings (first character is a single or double quote),
       to check if alert field labels["projects"] contains element "backend", we'll write
        has .labels.projects "backend"

       compound rules will look like

       and has .labels.projects "backend" or eq .labels.severity "warning" eq .labels.severity "crit"  -
       labels["project"] eq "backend" and (labels.severity eq "warning" or labels.severity eq "crit")

       and or eq .labels.severity "warning" eq .labels.severity "crit" gt since .labels.begin 86400 -
    (labels.severity eq "warning" or labels.severity eq "crit") and since labels["begin"] > 86400

    easier to parse , similar to text/template boolean conditions
  */
    var chRule ChannelRule
    var chRuleIdx int

    for chRuleIdx, chRule = range chDef.Rules {
        var rRule, partBuf []rune
        var currRuleParserState, newRuleParserState RuleParserStateType
        var rPos, partBufIdx int
        rRule = []rune(chRule.RuleStr)
        rPos = 0
        currRuleParserState = RuleParserLookingForArg
        partBuf = make([]rune, len(rRule))

        var copyToBuf bool
        var parserStack []RuleParserCtx
        var parserSP int

        partBufIdx = 0
        parserSP = 0
        parserStack = make([]RuleParserCtx, 0, 10)
        parserStack = append(parserStack, RuleParserCtx{&chRule.Root, 0})

        for (	(rPos <= len(rRule)) &&
                  (parserSP >= 0) &&
                  (err == nil) &&
                  (currRuleParserState != RuleParserInvalidState)) {
            var pCtx RuleParserCtx
            pCtx = parserStack[parserSP]
            copyToBuf, newRuleParserState = channelRuleParserRuneCB(rRule, rPos, currRuleParserState)
            switch (newRuleParserState & RULE_PARSER_MF_MASK) { // Sometimes higher bits (quote , backquote ) will be set here, but only in the RuleParserInArgXXX case
                case RuleParserLookingForArg:
                    partBufIdx = 0						// Just skipping the whitespace
                case RuleParserInArgFunc, RuleParserInArgKey, RuleParserInArgNum, RuleParserInArgStr, RuleParserInArgSet:
                    if (copyToBuf) {	// we may want to skip copying some characters (opening/closing and escaped quotes)
                        partBuf[partBufIdx] = rRule[rPos]
                        partBufIdx++
                    }
                case RuleParserAfterArg:
                    var argStr, funcName string
                    var hasArg bool
                    var argFunc PnFunc
                    hasArg = false
                    if (partBufIdx == 0) {
                        break; // End of line
                    }
                    argStr = string(partBuf[:partBufIdx])
					currRuleParserState = currRuleParserState & RULE_PARSER_MF_MASK // Clearing all modifiers

                    if (currRuleParserState == RuleParserInArgFunc) {
                        funcName = argStr
                    } else {
                        funcName = "value"
                        hasArg = true
                    }
                    if argFunc,ok = PnFuncDispatcher[funcName]; ok {
                        if (pCtx.partPtr.function == nil) {
                            pCtx.partPtr.function = argFunc.function
                            pCtx.partPtr.arguments = make([]RulePartArg, argFunc.numArgs)
                            pCtx.currArg = 0
                            parserStack[parserSP] = pCtx
                        } else {
                            var newPart RulePart
                            newPart = RulePart{function: argFunc.function, arguments: make([]RulePartArg, argFunc.numArgs)}
                            pCtx.partPtr.arguments[pCtx.currArg] = RulePartArg{PartArgTypeFunc, newPart}
                            pCtx.currArg++
                            parserStack[parserSP] = pCtx
                            parserStack = append(parserStack, RuleParserCtx{&newPart, 0})
                            parserSP++
                            pCtx = parserStack[parserSP]
                        }
                        if (hasArg) {
                            /*
                              We can convert between RuleParserState and RulePartArgType due to
                               carefully chosen values of both enums
                            */
                            if (currRuleParserState == RuleParserInArgSet) {
                            	pCtx.partPtr.arguments[pCtx.currArg] = RulePartArg{RulePartArgType(currRuleParserState), channelParseSet(argStr)}
                            } else {
                            	pCtx.partPtr.arguments[pCtx.currArg] = RulePartArg{RulePartArgType(currRuleParserState), argStr}
                            }
                            pCtx.currArg++
                            parserStack[parserSP] = pCtx
                        }
                    } else {
                        err = errors.New(fmt.Sprintf("%s: unknown function", argStr))
                        log.Printf("Error '%s' parsing rule %s", err, chRule.RuleStr)
                    }
                    newRuleParserState = RuleParserLookingForArg
                    partBufIdx = 0
                default:
                    err = errors.New(fmt.Sprintf("Invalid new parser state %d", newRuleParserState))
            } // End of a state machine switch
            if (newRuleParserState == RuleParserLookingForArg) {
              for parserSP >= 0 {
                  pCtx = parserStack[parserSP]
                  if (pCtx.currArg >= len(pCtx.partPtr.arguments)) { // We're done with current argument, let's pop this context from the stack
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

func channelParseSet(setStr string) []RulePartArg {
	/* extremely simplified parser from channelParseRules . Every fancy feature removed, just basic tokenization */
	var err error
	var ret []RulePartArg
	var rRule, partBuf []rune
    var currRuleParserState, newRuleParserState RuleParserStateType
    var rPos, partBufIdx int
    var copyToBuf bool
   	rRule = []rune(setStr)
	rPos = 0
	currRuleParserState = RuleParserLookingForArg
	partBuf = make([]rune, len(rRule))
	ret = make([]RulePartArg, 0, 2)
	for (rPos <= len(rRule) && (err == nil) && (currRuleParserState != RuleParserInvalidState)) {
    	copyToBuf, newRuleParserState = channelRuleParserRuneCB(rRule, rPos, currRuleParserState)
    	switch (newRuleParserState & RULE_PARSER_MF_MASK) {
			case RuleParserLookingForArg:
            	partBufIdx = 0
            case RuleParserInArgNum, RuleParserInArgStr:
                if (copyToBuf) { // we may want to skip copying some characters (opening/closing and escaped quotes)
                    partBuf[partBufIdx] = rRule[rPos]
                    partBufIdx++
                }
            case RuleParserAfterArg:
                if (partBufIdx == 0) {
					break
                }
                var argStr string
				argStr = string(partBuf[:partBufIdx])
				currRuleParserState = currRuleParserState & RULE_PARSER_MF_MASK // Clearing all modifiers
				if (currRuleParserState == RuleParserInArgNum) {
            		/* this is a number */
            		var num float64
            		num, err = strconv.ParseFloat(argStr, 64)
            		ret = append(ret, RulePartArg{RulePartArgType(currRuleParserState), num})
            	} else {
            		/* string goes unmodified */
            		ret = append(ret, RulePartArg{RulePartArgType(currRuleParserState), argStr})
            	}
            	newRuleParserState = RuleParserLookingForArg
                partBufIdx = 0
            default:
                err = errors.New(fmt.Sprintf("Invalid new parser state %d while parsing set %s", newRuleParserState, setStr))
        }
        currRuleParserState = newRuleParserState
       	rPos++
	}
	return ret
}

func channelAddSinkId(srcChDef *ChannelDef, sinkId uint32) (error) {
    var err error
    if srcChDef.Sinks == nil {
        srcChDef.Sinks = make([]uint32, 0, 2)
    }
    if (srcChDef.Id == sinkId) {
        log.Printf("Channel %d can't be a sink for itself", srcChDef.Id)
    } else {
        srcChDef.Sinks = append(srcChDef.Sinks , sinkId)
    }
    return err
}

func channelParseSrcChIds(srcChIds interface{}) ([]uint32) {
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
}

func channelPipeSrcsToSinks(channelDefs []*ChannelDef, lastChannelId uint32) {
    var chDef *ChannelDef;
    var chRule ChannelRule
    for _, chDef = range channelDefs {
        for _, chRule = range chDef.Rules {
            if (chRule.Root.function != nil) { // Rule was successfully parsed
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
}


func channelMatchAlert(channel *ChannelDef, srcChId uint32, alert Alert) bool {
    if (channel.Id == 0) {
        return true // special case
    }
    var chRule ChannelRule
    var ret bool
    ret = false
    for _, chRule = range channel.Rules {
        if (chRule.SrcChId == srcChId) {
            var rootRule RulePartArg
            rootRule = RulePartArg{argType: PartArgTypeFunc, argValue: chRule.Root}
            ret = ChannelGetBool(rootRule, alert)
            if ret {
                break
            }
        }
    }
    return ret
}

func channelRunTheGauntlet(	channelDefs []*ChannelDef, initialChId uint32, srcChId uint32,
                            alert Alert, groupsToDeliver map[string]string, totalMatches uint32) (uint32) { // Everything is a pointer, God bless Go magic
    var channel *ChannelDef
    var match bool
    var err error
    channel = channelDefs[initialChId]
    match = channelMatchAlert(channel, srcChId, alert)
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
            err = tp.Execute(&tpOutput, alert)
            if (err != nil) {
                log.Printf("Error %s rendering template", err)
            }
            groupsToDeliver[channel.Group] = tpOutput.String()
        }
        var sinkId uint32
        for _, sinkId = range channel.Sinks {
            totalMatches += channelRunTheGauntlet(channelDefs, sinkId, initialChId, alert, groupsToDeliver, totalMatches)
        }
    }
    return totalMatches
}
