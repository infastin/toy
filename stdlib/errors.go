package stdlib

import (
	"github.com/d5/tengo/v2"
)

func wrapError(err error) tengo.Object {
	if err == nil {
		return tengo.True
	}
	return &tengo.Error{Value: &tengo.String{Value: err.Error()}}
}
