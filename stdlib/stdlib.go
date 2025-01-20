package stdlib

import (
	"github.com/infastin/toy"
	"github.com/infastin/toy/stdlib/base64"
	"github.com/infastin/toy/stdlib/fmt"
	"github.com/infastin/toy/stdlib/hex"
	"github.com/infastin/toy/stdlib/json"
	"github.com/infastin/toy/stdlib/math"
	"github.com/infastin/toy/stdlib/net"
	"github.com/infastin/toy/stdlib/os"
	"github.com/infastin/toy/stdlib/os/env"
	"github.com/infastin/toy/stdlib/os/exec"
	ospath "github.com/infastin/toy/stdlib/os/path"
	"github.com/infastin/toy/stdlib/os/user"
	"github.com/infastin/toy/stdlib/path"
	"github.com/infastin/toy/stdlib/rand"
	"github.com/infastin/toy/stdlib/regexp"
	"github.com/infastin/toy/stdlib/text"
	"github.com/infastin/toy/stdlib/time"
	"github.com/infastin/toy/stdlib/uuid"
	"github.com/infastin/toy/stdlib/yaml"
)

var StdLib = toy.ModuleMap{
	"base64":  base64.Module,
	"fmt":     fmt.Module,
	"hex":     hex.Module,
	"json":    json.Module,
	"math":    math.Module,
	"net":     net.Module,
	"os":      os.Module,
	"os/env":  env.Module,
	"os/exec": exec.Module,
	"os/path": ospath.Module,
	"os/user": user.Module,
	"path":    path.Module,
	"rand":    rand.Module,
	"regexp":  regexp.Module,
	"text":    text.Module,
	"time":    time.Module,
	"uuid":    uuid.Module,
	"yaml":    yaml.Module,
}
