package fimap

import (
	"fmt"
	"math"
)

const (
	// PHI is for scrambling the keys
	PHI = 0x9E3779B9

	freeKey = 0
)

type keyType = uint64

// Map is a fast key (uint64) - value (interface{}) map.
type Map struct {
	keys   []keyType
	values []interface{}

	fillFactor float64
	threshold  int // we will resize a map once it reaches this size
	size       int

	mask keyType

	hasFreeKey bool        // have 'free' key in the map?
	freeVal    interface{} // value of 'free' key
}

//go:nosplit
func phiMix(x keyType) (h keyType) {
	h = x * PHI
	h ^= h >> 16
	return
}

//go:nosplit
func nextPowerOf2(x uint32) uint32 {
	if x == math.MaxUint32 {
		return x
	}

	if x == 0 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16

	return x + 1
}

//go:nosplit
func arraySize(exp uint, fill float64) (s uint32) {
	s = nextPowerOf2(uint32(math.Ceil(float64(exp) / fill)))
	if s < 2 {
		s = 2
	}
	return
}

// New returns a map initialized with preallocated `size` spaces and uses the stated fillFactor.
//
// When cardinality > capacity * fillFactor, the map will grow with 2x cap.
func New(size uint, fillFactor float64) (m *Map, err error) {
	if fillFactor <= 0 || fillFactor >= 1 {
		err = fmt.Errorf("FillFactor must be in (0, 1)")
		return
	}

	if size == 0 {
		size = 1
	}

	capacity := arraySize(size, fillFactor)
	m = &Map{
		keys:   make([]keyType, capacity), // x2 capacity for level-2
		values: make([]interface{}, capacity),

		fillFactor: fillFactor,
		threshold:  int(math.Floor(float64(capacity>>1) * fillFactor)),

		mask: keyType(capacity) - 1,
	}
	return
}

// Get value by key.
//go:nosplit
func (m *Map) Get(_key keyType) (_value interface{}, ok bool) {
	if _key == freeKey {
		_value, ok = m.freeVal, m.hasFreeKey

		return
	}

	ptr := (phiMix(_key) & m.mask)
	keys := m.keys
	key := keys[ptr]

	if key == _key {
		_value, ok = m.values[ptr], true
		return
	}
	if key == freeKey { // end of chain
		return
	}

	for {
		ptr = (ptr + 1) & m.mask
		key = keys[ptr]

		if key == _key {
			_value, ok = m.values[ptr], true
			return
		}
		if key == freeKey { // end of chain
			return
		}
	}
}

// Set key - value, overwrite if needed.
//go:nosplit
func (m *Map) Set(_key keyType, _value interface{}) {
	if _key != freeKey {
		if m.store(m.keys, m.values, _key, _value) {
			if m.size++; m.size > m.threshold {
				m.rehash()
			}
		}
	} else {
		m.freeVal = _value

		if !m.hasFreeKey {
			m.hasFreeKey = true
			m.size++
		}
	}
}

// store on external keys/values collection
//go:nosplit
func (m *Map) store(keys []keyType, values []interface{}, _key keyType, _value interface{}) (isNew bool) {
	ptr := (phiMix(_key) & m.mask)
	key := keys[ptr]

	if key == freeKey {
		keys[ptr], values[ptr] = _key, _value
		isNew = true
		return
	}
	if key == _key {
		values[ptr] = _value
		isNew = false
		return
	}

	for {
		ptr = (ptr + 1) & m.mask
		key = keys[ptr]

		if key == freeKey {
			keys[ptr], values[ptr] = _key, _value
			isNew = true
			return
		}
		if key == _key {
			values[ptr] = _value
			isNew = false
			return
		}
	}
}

//go:nosplit
func (m *Map) rehash() {
	originalLen := len(m.keys)

	// new capacity
	newCapacity := keyType(originalLen) << 1

	// update threshold and mask
	m.threshold = int(math.Floor(float64(newCapacity) * m.fillFactor))
	m.mask = newCapacity - 1

	// original keys, values
	oriKeys, oriValues := m.keys, m.values

	// write to new data
	keys, values := make([]keyType, newCapacity), make([]interface{}, newCapacity)
	for i, oriKey := range oriKeys {
		if oriKey != freeKey {
			m.store(keys, values, oriKey, oriValues[i])
		}
	}

	m.keys, m.values = keys, values
}

// Clone creates new map, copied from original one.
//go:nosplit
func (m *Map) Clone() *Map {
	c := *m

	c.keys, c.values = make([]keyType, len(m.keys)), make([]interface{}, len(m.keys))
	copy(c.keys, m.keys)
	copy(c.values, m.values)

	return &c
}

// Size returns size of the map.
//go:nosplit
func (m *Map) Size() int {
	return m.size
}
