package tengo

import (
	"encoding/binary"
	"math"
	"math/bits"
	"unsafe"
)

func murmur32(b []byte) uint32 {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
	)

	var h1 uint32
	p := b

	for len(p) >= 4 {
		k1 := binary.LittleEndian.Uint32(p)
		p = p[4:]

		k1 *= c1
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2

		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	var k1 uint32
	switch len(p) & 3 {
	case 3:
		k1 ^= uint32(p[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(p[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(p[0])
		k1 *= c1
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2
		h1 ^= k1
	}

	h1 ^= uint32(len(b))

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}

func murmur32String(s string) uint32 {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	return murmur32(b)
}

func murmur32Int32(i int32) uint32 {
	var buf [4]byte
	binary.NativeEndian.PutUint32(buf[:], uint32(i))
	return murmur32(buf[:])
}

func murmur32Int64(i int64) uint32 {
	var buf [8]byte
	binary.NativeEndian.PutUint64(buf[:], uint64(i))
	return murmur32(buf[:])
}

func murmur32Float64(f float64) uint32 {
	var buf [8]byte
	binary.NativeEndian.PutUint64(buf[:], math.Float64bits(f))
	return murmur32(buf[:])
}
