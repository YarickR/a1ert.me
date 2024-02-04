package dm_core
import (
	"dagproc/internal/di"
)

func LoadTemplatesConfig(cfg di.CFConfig) (map[string]di.Template, error) {
	var ret map[string]di.Template
	var err error
	ret = make(map[string]di.Template)
	err = nil
	return ret, err
}