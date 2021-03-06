package fimap

import (
	"fmt"
	"math"
)

type (
	keyType   = uint64
	valueType = interface{}
)

const (
	// PHI is for scrambling the keys
	PHI = 0x9E3779B9

	freeKey keyType = 0
)

var (
	nilValue = valueType(nil)
)

// Map is a fast key (uint64) - value (valueType) map.
type Map struct {
	keys   []keyType
	values []valueType

	fillFactor float64
	threshold  int // we will resize a map once it reaches this size
	size       int

	mask uint64

	hasFreeKey bool      // have 'free' key in the map?
	freeVal    valueType // value of 'free' key
}

func phiMix(x uint64) (h uint64) {
	h = x * PHI
	h ^= h >> 16
	return
}

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
		keys:   make([]keyType, capacity),
		values: make([]valueType, capacity),

		fillFactor: fillFactor,
		threshold:  int(math.Floor(float64(capacity>>1) * fillFactor)),

		mask: uint64(capacity) - 1,
	}
	return
}

// Get value by key.
func (m *Map) Get(k keyType) (v valueType, found bool) {
	if k != freeKey {
		if ind := index(m.keys, m.mask, k); m.keys[ind] != freeKey {
			v, found = m.values[ind], true
		}
	} else {
		v, found = m.freeVal, m.hasFreeKey
	}
	return
}

// Set key - value, overwrite if needed.
func (m *Map) Set(k keyType, v valueType) {
	if k != freeKey {
		if store(m.keys, m.values, m.mask, k, v) {
			if m.size++; m.size > m.threshold {
				m.grow()
			}
		}
	} else if m.freeVal = v; !m.hasFreeKey {
		m.hasFreeKey = true
		m.size++
	}
}

func distance(center, ind, nMask uint64) (dis uint64) {
	if ind >= center {
		dis = ind - center
	} else {
		dis = nMask + 1 - center + ind
	}
	return
}

// Remove an element
func (m *Map) Remove(k keyType) {
	var (
		keys, values, mask = m.keys, m.values, m.mask
		phi                uint64
		key                keyType
	)

	if k == freeKey {
		if m.hasFreeKey {
			m.freeVal, m.hasFreeKey = nilValue, false
			m.size--
		}
		return
	}

	// could remove?
	if ind := index(keys, mask, k); keys[ind] != freeKey {
		m.size--

		// find start position of current block
		startPos := uint64(ind)

	findStartPosLoop:
		if key = keys[startPos]; key != freeKey {
			phi = phiMix(uint64(key)) & mask

			// check if we could jump to phi right now
			if phi != startPos && keys[phi] != freeKey {
				startPos = phi
				goto findStartPosLoop
			}

			// if not, just move back to 1 step
			if startPos == 0 {
				startPos = mask
			} else {
				startPos = (startPos - 1) & mask
			}
			goto findStartPosLoop
		}

		// set free at ind
		keys[ind], values[ind] = freeKey, nilValue

		// now startPos is the position where current block start
		freePtr := uint64(ind) // index in uint64 type
		dis := distance(startPos, freePtr, mask)
		ptr := freePtr

	loop:
		// iterate each point in block and check if swapable
		ptr = (ptr + 1) & mask

		if key = keys[ptr]; key != freeKey {
			if phi = phiMix(uint64(key)) & mask; distance(startPos, phi, mask) <= dis { // swapable
				keys[freePtr], values[freePtr] = key, values[ptr]
				keys[ptr], values[ptr] = freeKey, nilValue

				freePtr = ptr
				dis = distance(startPos, freePtr, mask)
			}

			goto loop
		}

		m.shrink()
	}
}

// Iterate over map. Iteration will stop when handler return error.
func (m *Map) Iterate(handler func(keyType, valueType) error) (err error) {
	if handler != nil {
		values := m.values
		for i, k := range m.keys {
			if k != freeKey {
				if err = handler(k, values[i]); err != nil {
					return
				}
			}
		}

		if m.hasFreeKey {
			err = handler(freeKey, m.freeVal)
		}
	}
	return
}

// IterateAll iterates over map.
func (m *Map) IterateAll(handler func(keyType, valueType)) {
	if handler != nil {
		values := m.values
		for i, k := range m.keys {
			if k != freeKey {
				handler(k, values[i])
			}
		}

		if m.hasFreeKey {
			handler(freeKey, m.freeVal)
		}
	}
}

// Clone creates new map, copied from original one.
func (m *Map) Clone() *Map {
	c := *m

	c.keys, c.values = make([]keyType, len(m.keys)), make([]valueType, len(m.keys))
	copy(c.keys, m.keys)
	copy(c.values, m.values)

	return &c
}

// Size returns size of the map.
func (m *Map) Size() int {
	return m.size
}

// Reset map, keep underlying allocated space.
func (m *Map) Reset() {
	for i, k := range m.keys {
		if k != freeKey {
			m.keys[i], m.values[i] = freeKey, nilValue
		}
	}

	m.hasFreeKey, m.freeVal = false, nilValue

	m.size = 0
}

// (k != freeKey)
func index(keys []keyType, mask uint64, k keyType) (ind int) {
	ptr := phiMix(uint64(k)) & mask

	key := keys[ptr]

	if key == k || key == freeKey {
		ind = int(ptr)
		return
	}

loop:
	ptr = (ptr + 1) & mask
	key = keys[ptr]

	if key == k || key == freeKey {
		ind = int(ptr)
		return
	}
	goto loop
}

// store on external keys/values collection (k != freeKey)
func store(keys []keyType, values []valueType, mask uint64, k keyType, v valueType) (isNew bool) {
	ind := index(keys, mask, k)

	isNew = keys[ind] == freeKey

	if isNew {
		keys[ind], values[ind] = k, v
	} else {
		values[ind] = v
	}

	return
}

func (m *Map) grow() {
	originalLen := len(m.keys)

	// new capacity
	newCapacity := keyType(originalLen) << 1

	// update threshold and mask
	m.threshold = int(math.Floor(float64(newCapacity) * m.fillFactor))

	m.mask = newCapacity - 1
	mask := m.mask

	// original keys, values
	oriKeys, oriValues := m.keys, m.values

	// write to new data
	keys, values := make([]keyType, newCapacity), make([]valueType, newCapacity)
	for i, oriKey := range oriKeys {
		if oriKey != freeKey {
			store(keys, values, mask, oriKey, oriValues[i])
		}
	}

	m.keys, m.values = keys, values
}

func (m *Map) shrink() {
	if m.size < m.threshold>>1 {
		originalLen := len(m.keys)

		// new capacity
		newCapacity := keyType(originalLen) >> 1

		// update threshold and mask
		m.threshold = int(math.Floor(float64(newCapacity) * m.fillFactor))

		m.mask = newCapacity - 1
		mask := m.mask

		// original keys, values
		oriKeys, oriValues := m.keys, m.values

		// write to new data
		keys, values := make([]keyType, newCapacity), make([]valueType, newCapacity)
		for i, oriKey := range oriKeys {
			if oriKey != freeKey {
				store(keys, values, mask, oriKey, oriValues[i])
			}
		}

		m.keys, m.values = keys, values
	}
}
