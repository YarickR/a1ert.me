package dm_core

import (
	"strings"
	"strconv"
)

/* Key is a dotted path */
func EventGetKeyValue(event map[string]interface{}, key string) interface{} {
    var err error
    var path []string
    path = strings.Split(key, ".")

    if ((len(path) <= 2) || (len(path[0]) > 0)) {
        // Second condition is extraneous, it checks if  key starts with a dot. But let's stay safe
        mLog.Error().Msgf("Invalid key %s", key)
        return nil
    }
    var pathIdx int
    var cursor interface{}
    cursor = event
    for pathIdx = 1; pathIdx < len(path); pathIdx++ {
        var currPart string
        currPart = path[pathIdx]
        if (len(currPart) == 0) {
            mLog.Error().Msgf("Zero length part in key %s", key)
            return nil
        }
        switch cursor.(type) { // since this data is unmarshaled by json, we know we have a limited set of possible types
            case []interface{}: // array, path[pathIdx] should contain array index

                var idx int
                idx,err = strconv.Atoi(currPart)
                if ((err != nil) || (idx < 0 ) || (idx >= len(cursor.([]interface{})))) {
                    mLog.Error().Msgf("Invalid index %d, full key %s, alert: %+v, cursor: %+v", idx, key, event, cursor)
                    return nil
                }
                cursor = cursor.([]interface{})[idx]
            case map[string]interface{}: // map, path[pathIdx] should contain map key
                var ok bool
                if cursor, ok = cursor.(map[string]interface{})[currPart]; ok == false {
                    mLog.Error().Msgf("No %s found , full key %s, alert: %+v, cursor: %+v", currPart, key, event, cursor)
                    return nil
                }
            default:
                mLog.Error().Msgf("Non-indexable member %s found , full key %s, alert: %+v, cursor: %+v", currPart, key, event, cursor)
                return nil
        }
    }
    return cursor
}
