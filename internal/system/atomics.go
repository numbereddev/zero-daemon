package system

import "sync/atomic"

type AtomicString struct {
	ptr atomic.Pointer[string]
}

func NewAtomicString(val string) *AtomicString {
	as := &AtomicString{}
	as.Store(val)
	return as
}

func (as *AtomicString) Load() string {
	p := as.ptr.Load()
	if p == nil {
		return ""
	}
	return *p
}

func (as *AtomicString) Store(val string) {
	as.ptr.Store(&val)
}

func (as *AtomicString) Swap(newVal string) string {
	oldPtr := as.ptr.Swap(&newVal)
	if oldPtr == nil {
		return ""
	}
	return *oldPtr
}

func (as *AtomicString) CompareAndSwap(oldVal, newVal string) bool {
	for {
		currentPtr := as.ptr.Load()
		currentVal := ""
		if currentPtr != nil {
			currentVal = *currentPtr
		}

		if currentVal != oldVal {
			return false
		}

		if as.ptr.CompareAndSwap(currentPtr, &newVal) {
			return true
		}
	}
}

func (as *AtomicString) String() string {
	return as.Load()
}
