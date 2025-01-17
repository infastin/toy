package stdlib

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/infastin/toy"
)

var OSUserModule = &toy.BuiltinModule{
	Name: "user",
	Members: map[string]toy.Object{
		"current":       &toy.BuiltinFunction{Name: "user.current", Func: osUserCurrent},
		"lookup":        &toy.BuiltinFunction{Name: "user.lookup", Func: osUserLookup},
		"lookupId":      &toy.BuiltinFunction{Name: "user.lookupId", Func: osUserLookupID},
		"groups":        &toy.BuiltinFunction{Name: "user.groups", Func: osUserGroups},
		"lookupGroup":   &toy.BuiltinFunction{Name: "user.lookupGroup", Func: osUserLookupGroup},
		"lookupGroupId": &toy.BuiltinFunction{Name: "user.lookupGroupId", Func: osUserLookupGroupID},
		"cacheDir":      &toy.BuiltinFunction{Name: "user.cacheDir", Func: makeARSE(os.UserCacheDir)},
		"configDir":     &toy.BuiltinFunction{Name: "user.configDir", Func: makeARSE(os.UserConfigDir)},
		"homeDir":       &toy.BuiltinFunction{Name: "user.homeDir", Func: makeARSE(os.UserHomeDir)},
	},
}

type User user.User

func (u *User) TypeName() string { return "user.User" }
func (u *User) String() string   { return fmt.Sprintf("user.User(%q)", u.Name) }
func (u *User) IsFalsy() bool    { return false }

func (u *User) Clone() toy.Object {
	c := new(User)
	*c = *u
	return c
}

func (u *User) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "name":
		return toy.String(u.Name), nil
	case "uid":
		return toy.String(u.Uid), nil
	case "gid":
		return toy.String(u.Gid), nil
	case "username":
		return toy.String(u.Username), nil
	case "homeDir":
		return toy.String(u.HomeDir), nil
	}
	return nil, toy.ErrNoSuchField
}

type Group user.Group

func (g *Group) TypeName() string { return "user.Group" }
func (g *Group) String() string   { return fmt.Sprintf("user.Group(%q)", g.Name) }
func (g *Group) IsFalsy() bool    { return false }

func (g *Group) Clone() toy.Object {
	c := new(Group)
	*c = *g
	return c
}

func (g *Group) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "name":
		return toy.String(g.Name), nil
	case "gid":
		return toy.String(g.Gid), nil
	}
	return nil, toy.ErrNoSuchField
}

func osUserCurrent(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	user, err := user.Current()
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*User)(user), toy.Nil}, nil
}

func osUserLookup(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	user, err := user.Lookup(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*User)(user), toy.Nil}, nil
}

func osUserLookupID(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var uid string
	if err := toy.UnpackArgs(args, "uid", &uid); err != nil {
		return nil, err
	}
	user, err := user.LookupId(uid)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*User)(user), toy.Nil}, nil
}

func osUserGroups(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	ids, err := os.Getgroups()
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	groups := make([]toy.Object, 0, len(ids))
	for _, id := range ids {
		group, err := user.LookupGroupId(strconv.Itoa(id))
		if err != nil {
			return toy.Tuple{
				toy.Nil,
				toy.NewErrorf("group %d: %s", id, err.Error()),
			}, nil
		}
		groups = append(groups, (*Group)(group))
	}
	return toy.Tuple{toy.NewArray(groups), toy.Nil}, nil
}

func osUserLookupGroup(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	group, err := user.LookupGroup(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*Group)(group), toy.Nil}, nil
}

func osUserLookupGroupID(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var gid string
	if err := toy.UnpackArgs(args, "gid", &gid); err != nil {
		return nil, err
	}
	group, err := user.LookupGroupId(gid)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*Group)(group), toy.Nil}, nil
}
