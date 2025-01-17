package stdlib

import (
	"time"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/token"
)

var TimeModule = &toy.BuiltinModule{
	Name: "time",
	Members: map[string]toy.Object{
		"Time":     TimeType,
		"Duration": DurationType,

		"parse":         toy.NewBuiltinFunction("time.parse", timeParse),
		"now":           toy.NewBuiltinFunction("time.now", timeNow),
		"date":          toy.NewBuiltinFunction("time.date", timeDate),
		"parseDuration": toy.NewBuiltinFunction("time.parseDuration", timeParseDuration),
		"since":         toy.NewBuiltinFunction("time.since", timeSince),
		"until":         toy.NewBuiltinFunction("time.until", timeUntil),
		"unix":          toy.NewBuiltinFunction("time.unix", timeUnix),
		"unixMicro":     toy.NewBuiltinFunction("time.unixMicro", timeUnixMicro),
		"unixMilli":     toy.NewBuiltinFunction("time.unixMilli", timeUnixMilli),

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

var TimeType = toy.NewType[Time]("time.Time", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) < 1 && len(args) > 3 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 3,
			Got:     len(args),
		}
	}
	if len(args) == 1 {
		_, isStr := args[0].(toy.String)
		if !isStr {
			var t Time
			if err := toy.Convert(&t, args[0]); err == nil {
				return t, nil
			}
		}
	}
	var (
		x        string
		layout   = time.RFC3339
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
})

func (t *Time) Unpack(o toy.Object) error {
	switch x := o.(type) {
	case Time:
		*t = x
	case toy.String:
		tm, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", string(x))
		if err != nil {
			return err
		}
		*t = Time(tm)
	default:
		return &toy.InvalidValueTypeError{
			Want: "time.Time or string",
			Got:  toy.TypeName(o),
		}
	}
	return nil
}

func (t Time) Type() toy.ObjectType { return TimeType }

func (t Time) String() string {
	return (time.Time)(t).Format("2006-01-02 15:04:05.999999999 -0700 MST")
}

func (t Time) IsFalsy() bool     { return (time.Time)(t).IsZero() }
func (t Time) Clone() toy.Object { return t }
func (t Time) Hash() uint64      { return hash.Int64(time.Time(t).UnixNano()) }

func (t Time) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(Time)
	if !ok {
		return false, toy.ErrInvalidOperator
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
	return false, toy.ErrInvalidOperator
}

func (t Time) BinaryOp(op token.Token, other toy.Object, right bool) (toy.Object, error) {
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
	return nil, toy.ErrInvalidOperator
}

func (t Time) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "year":
		return toy.Int(time.Time(t).Year()), nil
	case "month":
		return toy.Int(time.Time(t).Month()), nil
	case "day":
		return toy.Int(time.Time(t).Day()), nil
	case "weekday":
		return toy.Int(time.Time(t).Weekday()), nil
	case "isoWeek":
		year, week := time.Time(t).ISOWeek()
		return toy.Tuple{toy.Int(year), toy.Int(week)}, nil
	case "clock":
		hour, min, sec := time.Time(t).Clock()
		return toy.Tuple{toy.Int(hour), toy.Int(min), toy.Int(sec)}, nil
	case "hour":
		return toy.Int(time.Time(t).Hour()), nil
	case "minute":
		return toy.Int(time.Time(t).Minute()), nil
	case "second":
		return toy.Int(time.Time(t).Second()), nil
	case "nanosecond":
		return toy.Int(time.Time(t).Nanosecond()), nil
	case "yearDay":
		return toy.Int(time.Time(t).YearDay()), nil
	case "unix":
		return toy.Int(time.Time(t).Unix()), nil
	case "unixMilli":
		return toy.Int(time.Time(t).UnixMilli()), nil
	case "unixMicro":
		return toy.Int(time.Time(t).UnixMicro()), nil
	case "unixNano":
		return toy.Int(time.Time(t).UnixNano()), nil
	}
	method, ok := timeTimeMethods[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return method.WithReceiver(t), nil
}

var timeTimeMethods = map[string]*toy.BuiltinFunction{
	"format":     toy.NewBuiltinFunction("format", timeTimeFormat),
	"inLocation": toy.NewBuiltinFunction("inLocation", timeTimeInLocation),
	"round":      toy.NewBuiltinFunction("round", timeTimeRound),
	"truncate":   toy.NewBuiltinFunction("truncate", timeTimeTruncate),
}

func timeTimeFormat(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv   = args[0].(Time)
		layout string
	)
	if err := toy.UnpackArgs(args[1:], "layout", &layout); err != nil {
		return nil, err
	}
	return toy.String(time.Time(recv).Format(layout)), nil
}

func timeTimeInLocation(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv     = args[0].(Time)
		location string
	)
	if err := toy.UnpackArgs(args[1:], "location", &location); err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{Time(time.Time(recv).In(loc)), toy.Nil}, nil
}

func timeTimeRound(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(Time)
		dur  Duration
	)
	if err := toy.UnpackArgs(args, "dur", &dur); err != nil {
		return nil, err
	}
	return Time(time.Time(recv).Round(time.Duration(dur))), nil
}

func timeTimeTruncate(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(Time)
		dur  Duration
	)
	if err := toy.UnpackArgs(args, "dur", &dur); err != nil {
		return nil, err
	}
	return Time(time.Time(recv).Truncate(time.Duration(dur))), nil
}

func timeParse(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		x        string
		layout   = time.RFC3339
		location = "UTC"
	)
	if err := toy.UnpackArgs(args, "x", &x, "layout?", &layout, "location?", &location); err != nil {
		return nil, err
	}
	if location == "UTC" {
		t, err := time.Parse(layout, x)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		return Time(t), nil
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	t, err := time.ParseInLocation(layout, x, loc)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{Time(t), toy.Nil}, nil
}

func timeNow(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return Time(time.Now()), nil
}

func timeDate(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{Time(time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc)), toy.Nil}, nil
}

type Duration time.Duration

var DurationType = toy.NewType[Duration]("time.Duration", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func (d *Duration) Unpack(o toy.Object) error {
	switch x := o.(type) {
	case Duration:
		*d = x
	case toy.String:
		dur, err := time.ParseDuration(string(x))
		if err != nil {
			return err
		}
		*d = Duration(dur)
	default:
		return &toy.InvalidValueTypeError{
			Want: "time.Duration or string",
			Got:  toy.TypeName(o),
		}
	}
	return nil
}

func (d Duration) Type() toy.ObjectType { return DurationType }
func (d Duration) String() string       { return (time.Duration)(d).String() }
func (d Duration) IsFalsy() bool        { return d == 0 }
func (d Duration) Clone() toy.Object    { return d }

func (d Duration) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(Duration)
	if !ok {
		return false, toy.ErrInvalidOperator
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
	return false, toy.ErrInvalidOperator
}

func (d Duration) BinaryOp(op token.Token, other toy.Object, right bool) (toy.Object, error) {
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
			return d / y, nil
		case toy.Int:
			if !right {
				return d / Duration(y), nil
			}
		}
	case token.Mul:
		switch y := other.(type) {
		case toy.Int:
			return d * Duration(y), nil
		}
	}
	return nil, toy.ErrInvalidOperator
}

func (d Duration) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "hours":
		return toy.Float(time.Duration(d).Hours()), nil
	case "minutes":
		return toy.Float(time.Duration(d).Minutes()), nil
	case "seconds":
		return toy.Float(time.Duration(d).Seconds()), nil
	case "milliseconds":
		return toy.Int(time.Duration(d).Milliseconds()), nil
	case "microseconds":
		return toy.Int(time.Duration(d).Microseconds()), nil
	case "nanoseconds":
		return toy.Int(time.Duration(d).Nanoseconds()), nil
	}
	method, ok := timeDurationMethods[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return method.WithReceiver(d), nil
}

var timeDurationMethods = map[string]*toy.BuiltinFunction{
	"round":    toy.NewBuiltinFunction("round", timeDurationRound),
	"truncate": toy.NewBuiltinFunction("truncate", timeDurationTruncate),
}

func timeDurationRound(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(Duration)
		m    Duration
	)
	if err := toy.UnpackArgs(args, "m", &m); err != nil {
		return nil, err
	}
	return Duration(time.Duration(recv).Round(time.Duration(m))), nil
}

func timeDurationTruncate(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(Duration)
		m    Duration
	)
	if err := toy.UnpackArgs(args, "m", &m); err != nil {
		return nil, err
	}
	return Duration(time.Duration(recv).Truncate(time.Duration(m))), nil
}

func timeParseDuration(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var x string
	if err := toy.UnpackArgs(args, "x", &x); err != nil {
		return nil, err
	}
	d, err := time.ParseDuration(x)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{Duration(d), toy.Nil}, nil
}

func timeSince(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var t Time
	if err := toy.UnpackArgs(args, "t", &t); err != nil {
		return nil, err
	}
	return Duration(time.Since(time.Time(t))), nil
}

func timeUntil(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var t Time
	if err := toy.UnpackArgs(args, "t", &t); err != nil {
		return nil, err
	}
	return Duration(time.Until(time.Time(t))), nil
}

func timeUnix(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var sec, nsec int64
	if err := toy.UnpackArgs(args, "sec", &sec, "nsec", &nsec); err != nil {
		return nil, err
	}
	return Time(time.Unix(sec, nsec)), nil
}

func timeUnixMicro(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var usec int64
	if err := toy.UnpackArgs(args, "usec", &usec); err != nil {
		return nil, err
	}
	return Time(time.UnixMicro(usec)), nil
}

func timeUnixMilli(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var msec int64
	if err := toy.UnpackArgs(args, "msec", &msec); err != nil {
		return nil, err
	}
	return Time(time.UnixMilli(msec)), nil
}
