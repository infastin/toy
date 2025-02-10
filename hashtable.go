package toy

import (
	"fmt"
	"iter"
)

// hashtable is used to represent Table values.
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
	key, value Value
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

func (ht *hashtable) insert(k, v Value) error {
	if ht.table == nil {
		if err := ht.checkMutable("insert into", false); err != nil {
			return err
		}
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
			if eq, err := Equal(k, e.key); err != nil {
				return err
			} else if eq {
				if err := ht.checkMutable("assign to element of", true); err != nil {
					return err
				}
				e.value = v
				return nil
			}
		}
		if p.next == nil {
			break
		}
		p = p.next
	}

	// Key not found. p points to the last bucket.
	if err := ht.checkMutable("insert into", false); err != nil {
		return err
	}

	// Does the number of elements exceed the buckets' load factor?
	if overloaded(int(ht.len), len(ht.table)) {
		ht.grow()
		goto retry
	}

	if insert == nil {
		// No space in existing buckets. Add a new one to the bucket list.
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
	// calls Equal unnecessarily (since there can't be duplicate keys),
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

func (ht *hashtable) lookup(k Value) (v Value, found bool, err error) {
	h, err := Hash(k)
	if err != nil {
		return nil, false, err // unhashable
	}
	if h == 0 {
		h = 1 // zero is reserved
	}
	if ht.table == nil {
		return Nil, false, nil // empty
	}
	// Inspect each bucket in the bucket list.
	for p := &ht.table[h&(uint64(len(ht.table)-1))]; p != nil; p = p.next {
		for i := range p.entries {
			e := &p.entries[i]
			if e.hash != h {
				continue
			}
			if eq, err := Equal(k, e.key); err != nil {
				return nil, false, err
			} else if eq {
				return e.value, true, nil // found
			}
		}
	}
	return Nil, false, nil // not found
}

func (ht *hashtable) keys() []Value {
	keys := make([]Value, 0, ht.len)
	for e := ht.head; e != nil; e = e.next {
		keys = append(keys, e.key)
	}
	return keys
}

func (ht *hashtable) values() []Value {
	values := make([]Value, 0, ht.len)
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

func (ht *hashtable) delete(k Value) (v Value, err error) {
	if err := ht.checkMutable("delete from", false); err != nil {
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
			if eq, err := Equal(k, e.key); err != nil {
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
// verb+" immutable table" should describe the operation.
func (ht *hashtable) checkMutable(verb string, iterOk bool) error {
	if ht.immutable {
		return fmt.Errorf("cannot %s immutable table", verb)
	}
	if !iterOk && ht.itercount > 0 {
		return fmt.Errorf("cannot %s table during iteration", verb)
	}
	return nil
}

func (ht *hashtable) clear() error {
	if err := ht.checkMutable("clear", false); err != nil {
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

func (ht *hashtable) cloneAll(other *hashtable) error {
	for e := other.head; e != nil; e = e.next {
		if err := ht.insert(e.key.Clone(), e.value.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (ht *hashtable) freeze() {
	if !ht.immutable {
		ht.immutable = true
		for e := ht.head; e != nil; e = e.next {
			e.key = Freeze(e.key)
			e.value = Freeze(e.value)
		}
	}
}

func (ht *hashtable) addAll(other *hashtable) error {
	for e := other.head; e != nil; e = e.next {
		if err := ht.insert(e.key, e.value); err != nil {
			return err
		}
	}
	return nil
}

func (ht *hashtable) equal(other *hashtable) (bool, error) {
	if ht.len != other.len {
		return false, nil
	}
	for e := ht.head; e != nil; e = e.next {
		key, xval := e.key, e.value
		if yval, found, _ := other.lookup(key); !found {
			return false, nil
		} else if eq, err := Equal(xval, yval); err != nil {
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

func (ht *hashtable) contains(key Value) (bool, error) {
	h, err := Hash(key)
	if err != nil {
		return false, err // unhashable
	}
	if h == 0 {
		h = 1 // zero is reserved
	}
	if ht.table == nil {
		return false, nil // empty
	}
	// Inspect each bucket in the bucket list.
	for p := &ht.table[h&(uint64(len(ht.table)-1))]; p != nil; p = p.next {
		for i := range p.entries {
			e := &p.entries[i]
			if e.hash != h {
				continue
			}
			if eq, err := Equal(key, e.key); err != nil {
				return false, err
			} else if eq {
				return true, nil // found
			}
		}
	}
	return false, nil // not found
}

func (ht *hashtable) elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		if !ht.immutable {
			ht.itercount++
			defer func() { ht.itercount-- }()
		}
		for e := ht.head; e != nil && yield(e.value); e = e.next {
		}
	}
}

func (ht *hashtable) entries() iter.Seq2[Value, Value] {
	return func(yield func(Value, Value) bool) {
		if !ht.immutable {
			ht.itercount++
			defer func() { ht.itercount-- }()
		}
		for e := ht.head; e != nil && yield(e.key, e.value); e = e.next {
		}
	}
}
