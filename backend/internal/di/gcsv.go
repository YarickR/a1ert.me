/* Generic config structure validator */
package di
import (
//	"errors"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)


func ValidateConfig(template string, cfg interface{}, path string) error {
	var (
		ret error
		jst interface{}
	)
	ret = json.Unmarshal([]byte(template), &jst)
	if (ret != nil) {
		return fmt.Errorf("Error parsing JSON config template %s: %s", template, ret)
	}
	switch jst.(type) {
		case MSI:
		case map[string]interface{}:
			ret = validateConfigNode(jst, cfg, path)
		default:
			ret = fmt.Errorf("Config template is expected to be a json dictionary")
	}
	return ret
}

func checkRequiredKeys(tKeys []string, cKeys []string, path string) (bool, error) {
	var (
		i int
		ckm map[string]bool
		key string
		ok, hasStar bool
	)
	ckm = make(map[string]bool)
	hasStar = false
	for i  = range (cKeys) { // list to dict, for easier checking 
		ckm[cKeys[i]] = true
	};
	for i = range (tKeys) {
		key = tKeys[i]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			_, ok = ckm[key];
			if !ok {
				return false, fmt.Errorf("Missing required key %s.%s", path, key);
			}
		}
		if (key == "*") {
			hasStar = true
		}
	}
	return hasStar, nil
}

func validateConfigNode(t interface{}, c interface{}, path string) error {
	var (
		ret error
	)
	if ((t == nil) || (c == nil)) {
		return fmt.Errorf("template or config is nil at %s", path)
	};
	if ((reflect.ValueOf(t).Kind() == reflect.String) && (t.(string) == "%"))  { // "%" means any type , no need to check further
		return nil
	};
	tt := reflect.TypeOf(t)
	ct := reflect.TypeOf(c)
	if (tt != ct) {
		if (tt.Kind() == ct.Kind()) {
			return fmt.Errorf("Same kind, different types: template: %s, config: %s, path: %s", tt, ct, path)
		};
		return fmt.Errorf("Completely different types: template: %s, config: %s, path: %s", tt, ct, path)
	}
	ret = nil
	switch t.(type) {
		case map[string]interface{}:
			var (
				tk []string
				ck []string
				mi string // map index, star index, new path
				ok, hasStar bool
			)
			tk = make([]string, 0, len(t.(MSI)))
			ck = make([]string, 0, len(c.(MSI)))
			for mi = range(t.(MSI)) {
				tk = append(tk, mi)
			};
			for mi = range(c.(MSI)) {
				ck = append(ck, mi)
			};
			hasStar, ret  = checkRequiredKeys(tk, ck, path)
			if (ret == nil) {
				if hasStar {
					for mi = range(c.(MSI)) {
						_, ok = t.(MSI)[mi]
						if ok {
							ret = validateConfigNode(t.(MSI)[mi], c.(MSI)[mi], fmt.Sprintf("%s.%s", path, mi));
						} else {
							ret = validateConfigNode(t.(MSI)["*"], c.(MSI)[mi], fmt.Sprintf("%s.%s", path, mi));
						}
						if (ret != nil) {
							break ;
						}
					}

				} else {
					for mi = range(t.(MSI)) {
						ret = validateConfigNode(t.(MSI)[mi], c.(MSI)[mi], fmt.Sprintf("%s.%s", path, mi));
						if (ret != nil) {
							break ;
						}
					}
				}
			}
		case []interface{}:
			var li int // list index
			if (len(t.([]interface{})) == 1) { // single item in template array means every item in config should be of the same type
				for li = range(c.([]interface{})) {
					ret = validateConfigNode(t.([]interface{})[0], c.([]interface{})[li], fmt.Sprintf("%s.%d", path, li));
					if (ret != nil) {
						return ret;
					}
				}
			} else {
				for li = range(t.([]interface{})) {
					ret = validateConfigNode(t.([]interface{})[li], c.([]interface{})[li], fmt.Sprintf("%s.%d", path, li));
					if (ret != nil) {
						return ret;
					}
				}
			}
		default:
			return nil;  
	}
	return ret;
}
