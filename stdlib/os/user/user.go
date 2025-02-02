package user

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
)

var Module = &toy.BuiltinModule{
	Name: "user",
	Members: map[string]toy.Value{
		"current":       toy.NewBuiltinFunction("user.current", currentFn),
		"lookup":        toy.NewBuiltinFunction("user.lookup", lookupFn),
		"lookupID":      toy.NewBuiltinFunction("user.lookupID", lookupIDFn),
		"groups":        toy.NewBuiltinFunction("user.groups", groupsFn),
		"lookupGroup":   toy.NewBuiltinFunction("user.lookupGroup", lookupGroupFn),
		"lookupGroupID": toy.NewBuiltinFunction("user.lookupGroupID", lookupGroupIDFn),
		"cacheDir":      toy.NewBuiltinFunction("user.cacheDir", fndef.ARSE(os.UserCacheDir)),
		"configDir":     toy.NewBuiltinFunction("user.configDir", fndef.ARSE(os.UserConfigDir)),
		"homeDir":       toy.NewBuiltinFunction("user.homeDir", fndef.ARSE(os.UserHomeDir)),
	},
}

type User user.User

var UserType = toy.NewType[*User]("user.User", nil)

func (u *User) Type() toy.ValueType { return UserType }
func (u *User) String() string      { return fmt.Sprintf("user.User(%q)", u.Name) }
func (u *User) IsFalsy() bool       { return false }

func (u *User) Clone() toy.Value {
	c := new(User)
	*c = *u
	return c
}

func (u *User) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "name":
		return toy.String(u.Name), true, nil
	case "uid":
		return toy.String(u.Uid), true, nil
	case "gid":
		return toy.String(u.Gid), true, nil
	case "username":
		return toy.String(u.Username), true, nil
	case "homeDir":
		return toy.String(u.HomeDir), true, nil
	}
	return toy.Nil, false, nil
}

type Group user.Group

var GroupType = toy.NewType[*User]("user.Group", nil)

func (g *Group) Type() toy.ValueType { return GroupType }
func (g *Group) String() string      { return fmt.Sprintf("user.Group(%q)", g.Name) }
func (g *Group) IsFalsy() bool       { return false }

func (g *Group) Clone() toy.Value {
	c := new(Group)
	*c = *g
	return c
}

func (g *Group) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "name":
		return toy.String(g.Name), true, nil
	case "gid":
		return toy.String(g.Gid), true, nil
	}
	return toy.Nil, false, nil
}

func currentFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	return (*User)(user), nil
}

func lookupFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	user, err := user.Lookup(name)
	if err != nil {
		return nil, err
	}
	return (*User)(user), nil
}

func lookupIDFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var uid string
	if err := toy.UnpackArgs(args, "uid", &uid); err != nil {
		return nil, err
	}
	user, err := user.LookupId(uid)
	if err != nil {
		return nil, err
	}
	return (*User)(user), nil
}

func groupsFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	ids, err := os.Getgroups()
	if err != nil {
		return nil, err
	}
	groups := make([]toy.Value, 0, len(ids))
	for _, id := range ids {
		group, err := user.LookupGroupId(strconv.Itoa(id))
		if err != nil {
			return nil, fmt.Errorf("group %d: %s", id, err.Error())
		}
		groups = append(groups, (*Group)(group))
	}
	return toy.NewArray(groups), nil
}

func lookupGroupFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	group, err := user.LookupGroup(name)
	if err != nil {
		return nil, err
	}
	return (*Group)(group), nil
}

func lookupGroupIDFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var gid string
	if err := toy.UnpackArgs(args, "gid", &gid); err != nil {
		return nil, err
	}
	group, err := user.LookupGroupId(gid)
	if err != nil {
		return nil, err
	}
	return (*Group)(group), nil
}
