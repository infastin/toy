package toy

import (
	"context"
	"testing"
)

func TestEval(t *testing.T) {
	eval := func(
		expr string,
		params map[string]Object,
		expected Object,
	) {
		ctx := context.Background()
		actual, err := Eval(ctx, expr, params)
		expectNoError(t, err)
		expectEqual(t, expected, actual)
	}

	eval(`undefined`, nil, nil)
	eval(`1`, nil, Int(1))
	eval(`19 + 23`, nil, Int(42))
	eval(`"foo bar"`, nil, String("foo bar"))
	eval(`[1, 2, 3][1]`, nil, Int(2))

	eval(
		`5 + p`,
		map[string]Object{
			"p": Int(7),
		},
		Int(12),
	)
	eval(
		`"seven is " + p`,
		map[string]Object{
			"p": Int(7),
		},
		String("seven is 7"),
	)
	eval(
		`"" + a + b`,
		map[string]Object{
			"a": Int(7),
			"b": String(" is seven"),
		},
		String("7 is seven"),
	)
	eval(
		`a ? "success" : "fail"`,
		map[string]Object{
			"a": Int(1),
		},
		String("success"),
	)
}
