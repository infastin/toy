package stdlib

import (
	"os"

	"github.com/infastin/toy"
)

var OSUserModule = &toy.BuiltinModule{
	Name: "user",
	Members: map[string]toy.Object{
		"current":       &toy.BuiltinFunction{},
		"lookup":        &toy.BuiltinFunction{},
		"lookupId":      &toy.BuiltinFunction{},
		"groups":        &toy.BuiltinFunction{},
		"lookupGroup":   &toy.BuiltinFunction{},
		"lookupGroupId": &toy.BuiltinFunction{},
		"cacheDir":      &toy.BuiltinFunction{Name: "cacheDir", Func: makeARSE(os.UserCacheDir)},
		"configDir":     &toy.BuiltinFunction{Name: "configDir", Func: makeARSE(os.UserConfigDir)},
		"homeDir":       &toy.BuiltinFunction{Name: "homeDir", Func: makeARSE(os.UserHomeDir)},
	},
}
