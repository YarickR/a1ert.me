package dm_core

import (
	"dagproc/internal/di"
	"fmt"
)

func LoadTemplatesConfig(config di.MSI, path string) (map[string]di.TemplatePtr, error) {
	var ret map[string]di.TemplatePtr
	var err error
	mLog.Debug().Msg("loading templates")
	err = di.ValidateConfig(` { "*": "string" }`, config, path)
	ret = make(map[string]di.TemplatePtr)
	for k, v := range config {
		switch v := v.(type) {
		case string:
			var tt di.TemplatePtr
			var ok bool
			if tt, ok = ret[k]; ok {
				err = fmt.Errorf("template '%s' already defined, contents: '%s'", k, tt.Contents)
				return ret, err
			}
			var t di.TemplatePtr
			t = &di.Template{Contents: v}
			ret[k] = t
			mLog.Debug().Str("template", k).Msg("loaded template")
		default:
			err = fmt.Errorf("template '%s' should be a string", k)
			return ret, err
		}
	}
	err = nil
	mLog.Debug().Msgf("loaded templates: %+v", ret)
	return ret, err
}
