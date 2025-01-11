package toy_test

import (
	"context"
	"testing"

	"github.com/infastin/toy"
)

func TestEval(t *testing.T) {
	expectEval(t, `nil`, nil, toy.Nil)
	expectEval(t, `1`, nil, toy.Int(1))
	expectEval(t, `19 + 23`, nil, toy.Int(42))
	expectEval(t, `"foo bar"`, nil, toy.String("foo bar"))
	expectEval(t, `[1, 2, 3][1]`, nil, toy.Int(2))

	expectEval(t,
		`5 + p`,
		map[string]toy.Object{
			"p": toy.Int(7),
		},
		toy.Int(12),
	)

	expectEval(t,
		`"seven is " + p`,
		map[string]toy.Object{
			"p": toy.String("7"),
		},
		toy.String("seven is 7"),
	)

	expectEval(t,
		`"" + a + b`,
		map[string]toy.Object{
			"a": toy.String("7"),
			"b": toy.String(" is seven"),
		},
		toy.String("7 is seven"),
	)

	expectEval(t,
		`a ? "success" : "fail"`,
		map[string]toy.Object{
			"a": toy.Int(1),
		},
		toy.String("success"),
	)
}

func expectEval(t *testing.T, expr string, params map[string]toy.Object, expected toy.Object) {
	ctx := context.Background()
	actual, err := toy.Eval(ctx, expr, params)
	expectNoError(t, err)
	expectEqual(t, expected, actual)
}
