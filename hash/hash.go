package hash

import (
	"encoding/binary"
	"hash/maphash"
	"math"
	"unsafe"
)

var seed = maphash.MakeSeed()

func Bytes(b []byte) uint64 {
	return maphash.Bytes(seed, b)
}

func String(s string) uint64 {
	b := unsafe.Slice(unsafe.StringData(string(s)), len(s))
	return Bytes(b)
}

func Int32(i int32) uint64 {
	var buf [4]byte
	binary.NativeEndian.PutUint32(buf[:], uint32(i))
	return Bytes(buf[:])
}

func Uint32(i uint32) uint64 {
	var buf [4]byte
	binary.NativeEndian.PutUint32(buf[:], uint32(i))
	return Bytes(buf[:])
}

func Int64(i int64) uint64 {
	var buf [8]byte
	binary.NativeEndian.PutUint64(buf[:], uint64(i))
	return Bytes(buf[:])
}

func Uint64(i uint64) uint64 {
	var buf [8]byte
	binary.NativeEndian.PutUint64(buf[:], uint64(i))
	return Bytes(buf[:])
}

func Float32(f float32) uint64 {
	var buf [4]byte
	binary.NativeEndian.PutUint32(buf[:], math.Float32bits(f))
	return Bytes(buf[:])
}

func Float64(f float64) uint64 {
	var buf [8]byte
	binary.NativeEndian.PutUint64(buf[:], math.Float64bits(f))
	return Bytes(buf[:])
}
