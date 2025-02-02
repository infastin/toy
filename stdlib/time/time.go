package time

import (
	"fmt"
	"time"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/token"
)

var Module = &toy.BuiltinModule{
	Name: "time",
	Members: map[string]toy.Value{
		"Time":     TimeType,
		"Duration": DurationType,

		"now":       toy.NewBuiltinFunction("time.now", nowFn),
		"date":      toy.NewBuiltinFunction("time.date", dateFn),
		"since":     toy.NewBuiltinFunction("time.since", sinceFn),
		"until":     toy.NewBuiltinFunction("time.until", untilFn),
		"unix":      toy.NewBuiltinFunction("time.unix", unixFn),
		"unixMicro": toy.NewBuiltinFunction("time.unixMicro", unixMicroFn),
		"unixMilli": toy.NewBuiltinFunction("time.unixMilli", unixMilliFn),

		"nsec": Duration(time.Nanosecond),
		"usec": Duration(time.Microsecond),
		"msec": Duration(time.Millisecond),
		"sec":  Duration(time.Second),
		"min":  Duration(time.Minute),
		"hour": Duration(time.Hour),

		"ansic":       toy.String(time.ANSIC),
		"unixDate":    toy.String(time.UnixDate),
		"rubyDate":    toy.String(time.RubyDate),
		"rfc822":      toy.String(time.RFC822),
		"rfc822Z":     toy.String(time.RFC822Z),
		"rfc1123":     toy.String(time.RFC1123),
		"rfc1123Z":    toy.String(time.RFC1123Z),
		"rfc3339":     toy.String(time.RFC3339),
		"rfc3339Nano": toy.String(time.RFC3339Nano),
		"kitchen":     toy.String(time.Kitchen),
		"stamp":       toy.String(time.Stamp),
		"stampMilli":  toy.String(time.StampMilli),
		"stampMicro":  toy.String(time.StampMicro),
		"stampNano":   toy.String(time.StampNano),
		"dateTime":    toy.String(time.DateTime),
		"dateOnly":    toy.String(time.DateOnly),
		"timeOnly":    toy.String(time.TimeOnly),
	},
}

type Time time.Time

var TimeType = toy.NewType[Time]("time.Time", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	switch len(args) {
	case 1:
		if _, isStr := args[0].(toy.String); !isStr {
			var t Time
			if err := toy.Convert(&t, args[0]); err == nil {
				return t, nil
			}
		}
		fallthrough
	case 3:
		var (
			x        string
			layout   = time.RFC3339Nano
			location = "UTC"
		)
		if err := toy.UnpackArgs(args, "x", &x, "layout?", &layout, "location?", &location); err != nil {
			return nil, err
		}
		if location == "UTC" {
			t, err := time.Parse(layout, x)
			if err != nil {
				return nil, err
			}
			return Time(t), nil
		}
		loc, err := time.LoadLocation(location)
		if err != nil {
			return nil, err
		}
		t, err := time.ParseInLocation(layout, x, loc)
		if err != nil {
			return nil, err
		}
		return Time(t), nil
	default:
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 3,
			Got:     len(args),
		}
	}
})

func (t Time) Type() toy.ValueType { return TimeType }

func (t Time) String() string {
	s := (time.Time)(t).Format(time.RFC3339Nano)
	return fmt.Sprintf("time.Time(%q)", s)
}

func (t Time) IsFalsy() bool    { return (time.Time)(t).IsZero() }
func (t Time) Clone() toy.Value { return t }
func (t Time) Hash() uint64     { return hash.Int64(time.Time(t).UnixNano()) }

func (t Time) Compare(op token.Token, rhs toy.Value) (bool, error) {
	y, ok := rhs.(Time)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return time.Time(t).Equal(time.Time(y)), nil
	case token.NotEqual:
		return !time.Time(t).Equal(time.Time(y)), nil
	case token.Less:
		return time.Time(t).Compare(time.Time(y)) < 0, nil
	case token.Greater:
		return time.Time(t).Compare(time.Time(y)) > 0, nil
	case token.LessEq:
		return time.Time(t).Compare(time.Time(y)) <= 0, nil
	case token.GreaterEq:
		return time.Time(t).Compare(time.Time(y)) >= 0, nil
	}
	return false, toy.ErrInvalidOperation
}

func (t Time) BinaryOp(op token.Token, other toy.Value, right bool) (toy.Value, error) {
	switch y := other.(type) {
	case Time:
		switch op {
		case token.Sub:
			return Duration(time.Time(t).Sub(time.Time(y))), nil
		}
	case Duration:
		switch op {
		case token.Add:
			return Time(time.Time(t).Add(time.Duration(y))), nil
		case token.Sub:
			if !right {
				return Time(time.Time(t).Add(time.Duration(-y))), nil
			}
		}
	}
	return nil, toy.ErrInvalidOperation
}

func (t Time) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String((time.Time)(t).Format(time.RFC3339Nano))
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (t Time) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "year":
		return toy.Int(time.Time(t).Year()), true, nil
	case "month":
		return toy.Int(time.Time(t).Month()), true, nil
	case "day":
		return toy.Int(time.Time(t).Day()), true, nil
	case "weekday":
		return toy.Int(time.Time(t).Weekday()), true, nil
	case "isoWeek":
		year, week := time.Time(t).ISOWeek()
		return toy.Tuple{toy.Int(year), toy.Int(week)}, true, nil
	case "clock":
		hour, min, sec := time.Time(t).Clock()
		return toy.Tuple{toy.Int(hour), toy.Int(min), toy.Int(sec)}, true, nil
	case "hour":
		return toy.Int(time.Time(t).Hour()), true, nil
	case "minute":
		return toy.Int(time.Time(t).Minute()), true, nil
	case "second":
		return toy.Int(time.Time(t).Second()), true, nil
	case "nanosecond":
		return toy.Int(time.Time(t).Nanosecond()), true, nil
	case "yearDay":
		return toy.Int(time.Time(t).YearDay()), true, nil
	case "unix":
		return toy.Int(time.Time(t).Unix()), true, nil
	case "unixMilli":
		return toy.Int(time.Time(t).UnixMilli()), true, nil
	case "unixMicro":
		return toy.Int(time.Time(t).UnixMicro()), true, nil
	case "unixNano":
		return toy.Int(time.Time(t).UnixNano()), true, nil
	}
	method, ok := timeMethods[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return method.WithReceiver(t), true, nil
}

var timeMethods = map[string]*toy.BuiltinFunction{
	"format":     toy.NewBuiltinFunction("format", timeFormatMd),
	"inLocation": toy.NewBuiltinFunction("inLocation", timeInLocationMd),
	"round":      toy.NewBuiltinFunction("round", timeRoundMd),
	"truncate":   toy.NewBuiltinFunction("truncate", timeTruncateMd),
}

func timeFormatMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv   = args[0].(Time)
		layout string
	)
	if err := toy.UnpackArgs(args[1:], "layout", &layout); err != nil {
		return nil, err
	}
	return toy.String(time.Time(recv).Format(layout)), nil
}

func timeInLocationMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv     = args[0].(Time)
		location string
	)
	if err := toy.UnpackArgs(args[1:], "location", &location); err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, err
	}
	return Time(time.Time(recv).In(loc)), nil
}

func timeRoundMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv = args[0].(Time)
		dur  Duration
	)
	if err := toy.UnpackArgs(args, "dur", &dur); err != nil {
		return nil, err
	}
	return Time(time.Time(recv).Round(time.Duration(dur))), nil
}

func timeTruncateMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv = args[0].(Time)
		dur  Duration
	)
	if err := toy.UnpackArgs(args, "dur", &dur); err != nil {
		return nil, err
	}
	return Time(time.Time(recv).Truncate(time.Duration(dur))), nil
}

func nowFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return Time(time.Now()), nil
}

func dateFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		year, month, day, hour, min, sec, nsec int
		location                               string
	)
	if err := toy.UnpackArgs(args,
		"year?", &year,
		"month?", &month,
		"day?", &day,
		"hour?", &hour,
		"minute?", &min,
		"second?", &sec,
		"nanosecond?", &nsec,
		"location?", &location,
	); err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, err
	}
	return Time(time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc)), nil
}

type Duration time.Duration

var DurationType = toy.NewType[Duration]("time.Duration", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		d, err := time.ParseDuration(string(x))
		if err != nil {
			return nil, err
		}
		return Duration(d), nil
	default:
		var d Duration
		if err := toy.Convert(&d, x); err != nil {
			return nil, err
		}
		return d, nil
	}
})

func (d Duration) Type() toy.ValueType { return DurationType }

func (d Duration) String() string {
	s := time.Duration(d).String()
	return fmt.Sprintf("time.Duration(%q)", s)
}

func (d Duration) IsFalsy() bool    { return d == 0 }
func (d Duration) Clone() toy.Value { return d }

func (d Duration) Compare(op token.Token, rhs toy.Value) (bool, error) {
	y, ok := rhs.(Duration)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return d == y, nil
	case token.NotEqual:
		return d != y, nil
	case token.Less:
		return d < y, nil
	case token.Greater:
		return d > y, nil
	case token.LessEq:
		return d <= y, nil
	case token.GreaterEq:
		return d >= y, nil
	}
	return false, toy.ErrInvalidOperation
}

func (d Duration) BinaryOp(op token.Token, other toy.Value, right bool) (toy.Value, error) {
	switch op {
	case token.Add:
		switch y := other.(type) {
		case Duration:
			return d + y, nil
		case Time:
			return Time(time.Time(y).Add(time.Duration(d))), nil
		}
	case token.Sub:
		switch y := other.(type) {
		case Duration:
			return d - y, nil
		}
	case token.Quo:
		switch y := other.(type) {
		case Duration:
			if y == 0 {
				return nil, toy.ErrDivisionByZero
			}
			return d / y, nil
		case toy.Int:
			if !right {
				if y == 0 {
					return nil, toy.ErrDivisionByZero
				}
				return d / Duration(y), nil
			}
		}
	case token.Mul:
		switch y := other.(type) {
		case toy.Int:
			return d * Duration(y), nil
		}
	}
	return nil, toy.ErrInvalidOperation
}

func (d Duration) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String((time.Duration)(d).String())
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (d Duration) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "hours":
		return toy.Float(time.Duration(d).Hours()), true, nil
	case "minutes":
		return toy.Float(time.Duration(d).Minutes()), true, nil
	case "seconds":
		return toy.Float(time.Duration(d).Seconds()), true, nil
	case "milliseconds":
		return toy.Int(time.Duration(d).Milliseconds()), true, nil
	case "microseconds":
		return toy.Int(time.Duration(d).Microseconds()), true, nil
	case "nanoseconds":
		return toy.Int(time.Duration(d).Nanoseconds()), true, nil
	}
	method, ok := timeDurationMethods[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return method.WithReceiver(d), true, nil
}

var timeDurationMethods = map[string]*toy.BuiltinFunction{
	"round":    toy.NewBuiltinFunction("round", durationRoundMd),
	"truncate": toy.NewBuiltinFunction("truncate", durationTruncateMd),
}

func durationRoundMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv = args[0].(Duration)
		m    Duration
	)
	if err := toy.UnpackArgs(args, "m", &m); err != nil {
		return nil, err
	}
	return Duration(time.Duration(recv).Round(time.Duration(m))), nil
}

func durationTruncateMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv = args[0].(Duration)
		m    Duration
	)
	if err := toy.UnpackArgs(args, "m", &m); err != nil {
		return nil, err
	}
	return Duration(time.Duration(recv).Truncate(time.Duration(m))), nil
}

func sinceFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var t Time
	if err := toy.UnpackArgs(args, "t", &t); err != nil {
		return nil, err
	}
	return Duration(time.Since(time.Time(t))), nil
}

func untilFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var t Time
	if err := toy.UnpackArgs(args, "t", &t); err != nil {
		return nil, err
	}
	return Duration(time.Until(time.Time(t))), nil
}

func unixFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var sec, nsec int64
	if err := toy.UnpackArgs(args, "sec", &sec, "nsec", &nsec); err != nil {
		return nil, err
	}
	return Time(time.Unix(sec, nsec)), nil
}

func unixMicroFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var usec int64
	if err := toy.UnpackArgs(args, "usec", &usec); err != nil {
		return nil, err
	}
	return Time(time.UnixMicro(usec)), nil
}

func unixMilliFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var msec int64
	if err := toy.UnpackArgs(args, "msec", &msec); err != nil {
		return nil, err
	}
	return Time(time.UnixMilli(msec)), nil
}
