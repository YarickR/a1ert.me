package dm_core
import ( 
    "strings"
    "strconv"
    "dagproc/internal/di"
    "github.com/rs/zerolog/log"
)

func LoadChannelsConfig(jsc map[string]interface{}, plugins map[string]di.Plugin) (map[string]di.Channel, error) {
    var ret map[string]di.Channel
    var err error
    mLog.Debug().Msg("LoadChannelsConfig")
    ret = make(map[string]di.Channel)
    err = nil
    return ret, err
}

func ChannelGetKeyValue(event di.Event, key string) interface{} {
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
    cursor = event
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
                    log.Printf("Invalid index %d, full key %s, event: %+v, cursor: %+v", idx, key, event, cursor)
                    return nil
                }
                cursor = cursor.([]interface{})[idx]
            case map[string]interface{}: // map, path[pathIdx] should contain map key
                var ok bool 
                if cursor, ok = cursor.(map[string]interface{})[currPart]; ok == false {
                    log.Printf("No %s found , full key %s, event: %+v, cursor: %+v", currPart, key, event, cursor)
                    return nil
                }
            default:
                log.Printf("Non-indexable member %s found , full key %s, event: %+v, cursor: %+v", currPart, key, event, cursor)
                return nil
        }
    }
    return cursor
}
