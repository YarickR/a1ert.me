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
	if ret != nil {
		return fmt.Errorf("Error parsing JSON config template %s: %s", template, ret)
	}
	switch jst.(type) {
	case map[string]interface{}:
		ret = validateConfigNode(jst, cfg, path)
	default:
		ret = fmt.Errorf("Config template is expected to be a json dictionary")
	}
	return ret
}

func checkRequiredKeys(tKeys []string, cKeys []string, path string) (bool, error) {
	var (
		i           int
		ckm         map[string]bool
		key         string
		ok, hasStar bool
	)
	ckm = make(map[string]bool)
	hasStar = false
	for i = range cKeys { // list to dict, for easier checking
		ckm[cKeys[i]] = true
	}
	for i = range tKeys {
		key = tKeys[i]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			_, ok = ckm[key]
			if !ok {
				return false, fmt.Errorf("Missing required key %s.%s", path, key)
			}
		}
		if key == "*" {
			hasStar = true
		}
	}
	return hasStar, nil
}

func compareTypes(tt reflect.Type, ct reflect.Type, path string) error {
	var (
		tk, ck reflect.Kind
	)
	tk = tt.Kind()
	ck = ct.Kind()
	if tk != ck {
		return fmt.Errorf("Completely different types: template: %s, config: %s, path: %s", tt, ct, path)
	}
	switch tk {
	case reflect.Map:
		if tt.Key().Kind() != ct.Key().Kind() {
			return fmt.Errorf("Different key types: %s vs %s, path: %s", tt.Key().Kind(), ct.Key().Kind(), path)
		}
		if tt.Elem().Kind() != ct.Elem().Kind() {
			return fmt.Errorf("Different element types: %s vs %s, path: %s", tt.Elem().Kind(), ct.Elem().Kind(), path)
		}
	case reflect.Array, reflect.Slice:
		if tt.Elem().Kind() != ct.Elem().Kind() {
			return fmt.Errorf("Different element types: %s vs %s, path: %s", tt.Elem().Kind(), ct.Elem().Kind(), path)
		}
	default:
		if tk != ck {
			return fmt.Errorf("Different types: %s vs %s, path: %s", tt, ct, path)
		}

	}
	return nil
}
func normalizedMapKey(mk string) string {
	if strings.HasSuffix(mk, "!") {
		return strings.TrimSuffix(mk, "!")
	}
	return mk
}

func validateConfigNode(t interface{}, c interface{}, path string) error {
	var (
		ret error
	)
	if (t == nil) || (c == nil) {
		return nil // This is not a mandatory key
	}
	if (reflect.ValueOf(t).Kind() == reflect.String) && (t.(string) == "%") { // "%" means any type , no need to check further
		return nil
	}
	tt := reflect.TypeOf(t)
	ct := reflect.TypeOf(c)
	ret = compareTypes(tt, ct, path)
	if ret != nil {
		return ret
	}
	ret = nil
	switch t.(type) {
	case map[string]interface{}:
		var (
			tmsi, cmsi  map[string]interface{}
			tks, cks    []string // template keys slice, config keys slice
			mk, nmk     string   // map key, normalized map key
			ok, hasStar bool
		)
		tmsi = t.(map[string]interface{})
		cmsi = c.(map[string]interface{})
		tks = make([]string, 0, len(tmsi))
		cks = make([]string, 0, len(cmsi))
		for mk = range tmsi {
			tks = append(tks, mk)
		}
		for mk = range cmsi {
			cks = append(cks, mk)
		}
		hasStar, ret = checkRequiredKeys(tks, cks, path)
		if ret == nil {
			if hasStar {
				for mk = range cmsi {
					_, ok = tmsi[mk]
					if ok {
						ret = validateConfigNode(tmsi[mk], cmsi[mk], fmt.Sprintf("%s.%s", path, mk))
					} else {
						ret = validateConfigNode(tmsi["*"], cmsi[mk], fmt.Sprintf("%s.%s", path, mk))
					}
					if ret != nil {
						break
					}
				}

			} else {
				for mk = range tmsi {
					nmk = normalizedMapKey(mk)
					ret = validateConfigNode(tmsi[mk], cmsi[nmk], fmt.Sprintf("%s.%s", path, nmk))
					if ret != nil {
						break
					}
				}
			}
		}
	case []interface{}:
		var (
			li       int // list index
			tsi, csi []interface{}
		)
		tsi = t.([]interface{})
		csi = c.([]interface{})
		if len(tsi) == 1 { // single item in template array means every item in config should be of the same type
			for li = range csi {
				ret = validateConfigNode(tsi[0], csi[li], fmt.Sprintf("%s.%d", path, li))
				if ret != nil {
					return ret
				}
			}
		} else {
			for li = range tsi {
				ret = validateConfigNode(tsi[li], csi[li], fmt.Sprintf("%s.%d", path, li))
				if ret != nil {
					return ret
				}
			}
		}
	default:
		return nil
	}
	return ret
}

func MergeStructs(fields []string, src interface{}, dst interface{}) interface{} {
	for _, f := range fields {
		s := reflect.ValueOf(src)
		sf := s.FieldByName(f)
		d := reflect.ValueOf(dst)
		if sf.Len() > 0 {
			d.FieldByName(f).Set(sf)
		}
	}
	return dst
}
