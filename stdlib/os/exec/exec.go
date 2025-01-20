package exec

import (
	"os/exec"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
)

var Module = &toy.BuiltinModule{
	Name: "exec",
	Members: map[string]toy.Object{
		"command":  &toy.BuiltinFunction{},
		"lookPath": toy.NewBuiltinFunction("lookPath", fndef.ASRSE("file", exec.LookPath)),
	},
}

type Command struct {
}
