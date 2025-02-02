package rand

import (
	"crypto/rand"
	"math"
	"math/big"
	"math/bits"
	"strings"

	"github.com/infastin/toy"
)

var Module = &toy.BuiltinModule{
	Name: "rand",
	Members: map[string]toy.Value{
		"int":          toy.NewBuiltinFunction("rand.int", intFn),
		"float":        toy.NewBuiltinFunction("rand.float", floatFn),
		"text":         toy.NewBuiltinFunction("rand.text", textFn),
		"alpha":        toy.NewBuiltinFunction("rand.alpha", alphaFn),
		"alphanumeric": toy.NewBuiltinFunction("rand.alphanumeric", alphanumericFn),
		"ascii":        toy.NewBuiltinFunction("rand.ascii", asciiFn),
	},
}

func intFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	maxNum := int64(math.MaxInt64)
	if err := toy.UnpackArgs(args, "n?", &maxNum); err != nil {
		return nil, err
	}
	x, err := rand.Int(rand.Reader, big.NewInt(maxNum))
	if err != nil {
		return nil, err
	}
	return toy.Int(x.Int64()), nil
}

func floatFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	x, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		return nil, err
	}
	return toy.Float(x.Int64()) / (1 << 53), nil
}

var (
	alphaAlphabet    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	alphaNumAlphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	asciiAlphabet    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!\"#$%&'()*+,-./:;<=>?@[]^_`{|}~")
)

func randStr(count int, alphabet []rune) (string, error) {
	var (
		b       strings.Builder
		idxBits = bits.Len(uint(len(alphabet)))
		idxMask = 1<<idxBits - 1
		idxMax  = (bits.UintSize - 1) / idxBits
	)
	b.Grow(count)
	tmp, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}
	for i, cache, remain := count-1, int(tmp.Int64()), idxMax; i >= 0; {
		if remain == 0 {
			tmp, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
			if err != nil {
				return "", err
			}
			cache, remain = int(tmp.Int64()), idxMax
		}
		if idx := cache & idxMask; idx < len(alphabet) {
			b.WriteRune(alphabet[idx])
			i--
		}
		cache >>= idxBits
		remain--
	}
	return b.String(), nil
}

func textFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		alphabet string
		n        int
	)
	if err := toy.UnpackArgs(args, "alphabet", &alphabet, "n", &n); err != nil {
		return nil, err
	}
	str, err := randStr(n, []rune(alphabet))
	if err != nil {
		return nil, err
	}
	return toy.String(str), nil
}

func alphaFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var n int
	if err := toy.UnpackArgs(args, "n", &n); err != nil {
		return nil, err
	}
	str, err := randStr(n, alphaAlphabet)
	if err != nil {
		return nil, err
	}
	return toy.String(str), nil
}

func alphanumericFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var n int
	if err := toy.UnpackArgs(args, "n", &n); err != nil {
		return nil, err
	}
	str, err := randStr(n, alphaNumAlphabet)
	if err != nil {
		return nil, err
	}
	return toy.String(str), nil
}

func asciiFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var n int
	if err := toy.UnpackArgs(args, "n", &n); err != nil {
		return nil, err
	}
	str, err := randStr(n, asciiAlphabet)
	if err != nil {
		return nil, err
	}
	return toy.String(str), nil
}
