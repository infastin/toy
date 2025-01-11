package toy

import (
	"fmt"
	"iter"
	"math/big"
)

// hashtable is used to represent Toy map entries.
// It is a hash table whose key/value entries form a doubly-linked list
// in the order the entries were inserted.
//
// Initialized instances of hashtable must not be copied.
//
// It's a slightly modified version of the hashtable used
// in the Starlark interpreter written in Go.
// See https://github.com/google/starlark-go.
type hashtable struct {
	table     []bucket  // len is zero or a power of two
	bucket0   [1]bucket // inline allocation for small maps.
	len       uint64
	itercount uint64  // number of active iterators (ignored if frozen)
	head      *entry  // insertion order doubly-linked list; may be nil
	tailLink  **entry // address of nil link at end of list (perhaps &head)
	immutable bool

	_ noCopy // triggers vet copylock check on this type.
}

// noCopy is zero-sized type that triggers vet's copylock check.
// See https://github.com/golang/go/issues/8005#issuecomment-190753527.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

const bucketSize = 8

type bucket struct {
	entries [bucketSize]entry
	next    *bucket // linked list of buckets
}

type entry struct {
	hash       uint64 // nonzero => in use
	key, value Object
	next       *entry  // insertion order doubly-linked list; may be nil
	prevLink   **entry // address of link to this entry (perhaps &head)
}

func (ht *hashtable) init(size int) {
	if size < 0 {
		panic("size < 0")
	}
	nb := 1
	for overloaded(size, nb) {
		nb = nb << 1
	}
	if nb < 2 {
		ht.table = ht.bucket0[:1]
	} else {
		ht.table = make([]bucket, nb)
	}
	ht.tailLink = &ht.head
}

func (ht *hashtable) insert(k, v Object) error {
	if err := ht.checkMutable("insert into"); err != nil {
		return err
	}
	if ht.table == nil {
		ht.init(1)
	}
	h, err := Hash(k)
	if err != nil {
		return err // unhashable
	}
	if h == 0 {
		h = 1 // zero is reserved
	}

retry:
	var insert *entry

	// Inspect each bucket in the bucket list.
	p := &ht.table[h&(uint64(len(ht.table)-1))]
	for {
		for i := range p.entries {
			e := &p.entries[i]
			if e.hash != h {
				if e.hash == 0 {
					// Found empty entry; make a note.
					insert = e
				}
				continue
			}
			if eq, err := Equals(k, e.key); err != nil {
				return err
			} else if eq {
				e.value = v
				return nil
			}
		}
		if p.next == nil {
			break
		}
		p = p.next
	}

	// Key not found.  p points to the last bucket.

	// Does the number of elements exceed the buckets' load factor?
	if overloaded(int(ht.len), len(ht.table)) {
		ht.grow()
		goto retry
	}

	if insert == nil {
		// No space in existing buckets.  Add a new one to the bucket list.
		b := new(bucket)
		p.next = b
		insert = &b.entries[0]
	}

	// Insert key/value pair.
	insert.hash = h
	insert.key = k
	insert.value = v

	// Append entry to doubly-linked list.
	insert.prevLink = ht.tailLink
	*ht.tailLink = insert
	ht.tailLink = &insert.next

	ht.len++

	return nil
}

func overloaded(elems, buckets int) bool {
	const loadFactor = 6.5 // just a guess
	return elems >= bucketSize && float64(elems) >= loadFactor*float64(buckets)
}

func (ht *hashtable) grow() {
	// Double the number of buckets and rehash.
	//
	// Even though this makes reentrant calls to ht.insert,
	// calls Equals unnecessarily (since there can't be duplicate keys),
	// and recomputes the hash unnecessarily, the gains from
	// avoiding these steps were found to be too small to justify
	// the extra logic: -2% on hashtable benchmark.
	ht.table = make([]bucket, len(ht.table)<<1)
	oldhead := ht.head
	ht.head = nil
	ht.tailLink = &ht.head
	ht.len = 0
	for e := oldhead; e != nil; e = e.next {
		ht.insert(e.key, e.value)
	}
	ht.bucket0[0] = bucket{} // clear out unused initial bucket
}

func (ht *hashtable) lookup(k Object) (v Object, err error) {
	h, err := Hash(k)
	if err != nil {
		return nil, err // unhashable
	}
	if h == 0 {
		h = 1 // zero is reserved
	}
	if ht.table == nil {
		return Nil, nil // empty
	}
	// Inspect each bucket in the bucket list.
	for p := &ht.table[h&(uint64(len(ht.table)-1))]; p != nil; p = p.next {
		for i := range p.entries {
			e := &p.entries[i]
			if e.hash != h {
				continue
			}
			if eq, err := Equals(k, e.key); err != nil {
				return nil, err
			} else if eq {
				return e.value, nil // found
			}
		}
	}
	return Nil, nil // not found
}

// count returns the number of distinct elements of iter that are elements of ht.
func (ht *hashtable) count(iter Iterator) (int, error) {
	if ht.table == nil {
		return 0, nil // empty
	}
	var (
		k     Object
		count = 0
	)
	// Use a bitset per table entry to record seen elements of ht.
	// Elements are identified by their bucket number and index within the bucket.
	// Each bitset gets one word initially, but may grow.
	storage := make([]big.Word, len(ht.table))
	bitsets := make([]big.Int, len(ht.table))
	for i := range bitsets {
		bitsets[i].SetBits(storage[i : i+1 : i+1])
	}
	for iter.Next(&k, nil) && count != int(ht.len) {
		h, err := Hash(k)
		if err != nil {
			return 0, err // unhashable
		}
		if h == 0 {
			h = 1 // zero is reserved
		}
		// Inspect each bucket in the bucket list.
		bucketId := h & (uint64(len(ht.table) - 1))
		i := 0
		for p := &ht.table[bucketId]; p != nil; p = p.next {
			for j := range p.entries {
				e := &p.entries[j]
				if e.hash != h {
					continue
				}
				if eq, err := Equals(k, e.key); err != nil {
					return 0, err
				} else if eq {
					bitIndex := i<<3 + j
					if bitsets[bucketId].Bit(bitIndex) == 0 {
						bitsets[bucketId].SetBit(&bitsets[bucketId], bitIndex, 1)
						count++
					}
				}
			}
			i++
		}
	}
	return count, nil
}

func (ht *hashtable) keys() []Object {
	keys := make([]Object, 0, ht.len)
	for e := ht.head; e != nil; e = e.next {
		keys = append(keys, e.key)
	}
	return keys
}

func (ht *hashtable) values() []Object {
	values := make([]Object, 0, ht.len)
	for e := ht.head; e != nil; e = e.next {
		values = append(values, e.value)
	}
	return values
}

func (ht *hashtable) items() []Tuple {
	items := make([]Tuple, 0, ht.len)
	for e := ht.head; e != nil; e = e.next {
		items = append(items, Tuple{e.key, e.value})
	}
	return items
}

func (ht *hashtable) delete(k Object) (v Object, err error) {
	if err := ht.checkMutable("delete from"); err != nil {
		return nil, err
	}
	if ht.table == nil {
		return Nil, nil // empty
	}
	h, err := Hash(k)
	if err != nil {
		return nil, err // unhashable
	}
	if h == 0 {
		h = 1 // zero is reserved
	}

	// Inspect each bucket in the bucket list.
	for p := &ht.table[h&(uint64(len(ht.table)-1))]; p != nil; p = p.next {
		for i := range p.entries {
			e := &p.entries[i]
			if e.hash != h {
				continue
			}
			if eq, err := Equals(k, e.key); err != nil {
				return nil, err
			} else if eq {
				// Remove e from doubly-linked list.
				*e.prevLink = e.next
				if e.next == nil {
					ht.tailLink = e.prevLink // deletion of last entry
				} else {
					e.next.prevLink = e.prevLink
				}

				v := e.value
				*e = entry{}
				ht.len--
				return v, nil // found
			}
		}
	}

	return Nil, nil // not found
}

// checkMutable reports an error if the hash table should not be mutated.
// verb+" immutable hash table" should describe the operation.
func (ht *hashtable) checkMutable(verb string) error {
	if ht.immutable {
		return fmt.Errorf("cannot %s immutable hash table", verb)
	}
	if ht.itercount > 0 {
		return fmt.Errorf("cannot %s hash table during iteration", verb)
	}
	return nil
}

func (ht *hashtable) clear() error {
	if err := ht.checkMutable("clear"); err != nil {
		return err
	}
	if ht.table != nil {
		for i := range ht.table {
			ht.table[i] = bucket{}
		}
	}
	ht.head = nil
	ht.tailLink = &ht.head
	ht.len = 0
	return nil
}

func (ht *hashtable) copyAll(other *hashtable) error {
	for e := other.head; e != nil; e = e.next {
		if err := ht.insert(e.key.Copy(), e.value.Copy()); err != nil {
			return err
		}
	}
	return nil
}

func (ht *hashtable) copyAllImmutable(other *hashtable) error {
	for e := other.head; e != nil; e = e.next {
		if err := ht.insert(AsImmutable(e.key), AsImmutable(e.value)); err != nil {
			return err
		}
	}
	return nil
}

func (ht *hashtable) addAll(other *hashtable) error {
	for e := other.head; e != nil; e = e.next {
		if err := ht.insert(e.key, e.value); err != nil {
			return err
		}
	}
	return nil
}

func (ht *hashtable) equals(other *hashtable) (bool, error) {
	if ht.len != other.len {
		return false, nil
	}
	for e := ht.head; e != nil; e = e.next {
		key, xval := e.key, e.value
		if yval, _ := other.lookup(key); yval == Nil {
			return false, nil
		} else if eq, err := Equals(xval, yval); err != nil {
			return false, err
		} else if !eq {
			return false, nil
		}
	}
	return true, nil
}

// dump is provided as an aid to debugging.
func (ht *hashtable) dump() {
	fmt.Printf("hashtable %p len=%d head=%p tailLink=%p",
		ht, ht.len, ht.head, ht.tailLink)
	if ht.tailLink != nil {
		fmt.Printf(" *tailLink=%p", *ht.tailLink)
	}
	fmt.Println()
	for j := range ht.table {
		fmt.Printf("bucket chain %d\n", j)
		for p := &ht.table[j]; p != nil; p = p.next {
			fmt.Printf("bucket %p\n", p)
			for i := range p.entries {
				e := &p.entries[i]
				fmt.Printf("\tentry %d @ %p hash=%d key=%v value=%v\n",
					i, e, e.hash, e.key, e.value)
				fmt.Printf("\t\tnext=%p &next=%p prev=%p",
					e.next, &e.next, e.prevLink)
				if e.prevLink != nil {
					fmt.Printf(" *prev=%p", *e.prevLink)
				}
				fmt.Println()
			}
		}
	}
}

func (ht *hashtable) iterate() *htIterator {
	if !ht.immutable {
		ht.itercount++
	}
	return &htIterator{ht: ht, e: ht.head}
}

type htIterator struct {
	ht *hashtable
	e  *entry
}

func (it *htIterator) TypeName() string { return "hashtable-iterator" }
func (it *htIterator) String() string   { return "<hashtable-iterator>" }
func (it *htIterator) IsFalsy() bool    { return true }
func (it *htIterator) Copy() Object     { return &htIterator{ht: it.ht, e: it.e} }

func (it *htIterator) Next(key, value *Object) bool {
	if it.e != nil {
		if key != nil {
			*key = it.e.key
		}
		if value != nil {
			*value = it.e.value
		}
		it.e = it.e.next
		return true
	}
	return false
}

func (it *htIterator) Close() {
	if !it.ht.immutable {
		it.ht.itercount--
	}
}

// elements is a go1.23 iterator over the values of the hash table.
func (ht *hashtable) elements() iter.Seq[Object] {
	return func(yield func(Object) bool) {
		if !ht.immutable {
			ht.itercount++
			defer func() { ht.itercount-- }()
		}
		for e := ht.head; e != nil && yield(e.value); e = e.next {
		}
	}
}

// entries is a go1.23 iterator over the entries of the hash table.
func (ht *hashtable) entries() iter.Seq2[Object, Object] {
	return func(yield func(Object, Object) bool) {
		if !ht.immutable {
			ht.itercount++
			defer func() { ht.itercount-- }()
		}
		for e := ht.head; e != nil && yield(e.key, e.value); e = e.next {
		}
	}
}
