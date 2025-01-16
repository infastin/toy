package stdlib

import "github.com/infastin/toy"

var StdLib = toy.ModuleMap{
	"base64":  Base64Module,
	"fmt":     FmtModule,
	"hex":     HexModule,
	"json":    JSONModule,
	"math":    MathModule,
	"net":     NetModule,
	"os":      OSModule,
	"os/env":  OSEnvModule,
	"os/exec": OSExecModule,
	"os/path": OSPathModule,
	"os/user": OSUserModule,
	"path":    PathModule,
	"rand":    RandModule,
	"regexp":  RegexpModule,
	"text":    TextModule,
	"time":    TimeModule,
	"uuid":    UUIDModule,
	"yaml":    YAMLModule,
}
