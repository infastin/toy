package stdlib

import "github.com/infastin/toy"

var StdLib = toy.ModuleMap{
	"base64": Base64Module,
	"fmt":    FmtModule,
	"json":   JSONModule,
	"os":     OSModule,
	"path":   PathModule,
	"rand":   RandModule,
	"regexp": RegexpModule,
	"text":   TextModule,
	"time":   TimeModule,
	"uuid":   UUIDModule,
	"yaml":   YAMLModule,
}
