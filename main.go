package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type atomicMapBuffer struct {
	buffer unsafe.Pointer
}

type keyValue struct {
	key   string
	value interface{}
}

func newAtomicMapBuffer() *atomicMapBuffer {
	return &atomicMapBuffer{
		buffer: unsafe.Pointer(&([]keyValue{})),
	}
}

func (m *atomicMapBuffer) load(key string) (interface{}, bool) {
	buffer := (*[]keyValue)(atomic.LoadPointer(&m.buffer))
	for _, kv := range *buffer {
		if kv.key == key {
			return kv.value, true
		}
	}
	return nil, false
}

func (m *atomicMapBuffer) store(key string, value interface{}) {
	for {
		oldBuffer := (*[]keyValue)(atomic.LoadPointer(&m.buffer))
		newBuffer := make([]keyValue, len(*oldBuffer)+1)
		copy(newBuffer, *oldBuffer)
		newBuffer[len(*oldBuffer)] = keyValue{key: key, value: value}

		if atomic.CompareAndSwapPointer(&m.buffer, unsafe.Pointer(oldBuffer), unsafe.Pointer(&newBuffer)) {
			return
		}
	}
}

func (m *atomicMapBuffer) store2(key string, value interface{}) {
	for {
		oldBuffer := (*[]keyValue)(atomic.LoadPointer(&m.buffer))
		newBuffer := make([]keyValue, len(*oldBuffer)+1)
		copy(newBuffer, *oldBuffer)
		newBuffer[len(*oldBuffer)] = keyValue{key: key, value: value}

		atomic.StorePointer(&m.buffer, unsafe.Pointer(&newBuffer))
		return
	}
}

func main() {
	// fmt.Println("Hash map ...")
	// hash := uint64(10)
	// //key := "key"
	// fmt.Println(hash)
	// fmt.Println(uintptr(0))
	// fmt.Println(math.Log(8))
	// fmt.Println(hashFunc("apple"))

	// fmt.Println(log_2(8))
	// fmt.Println(*(*[]byte)(unsafe.Pointer(&key)))
	// for _, c := range *(*[]byte)(unsafe.Pointer(&key)) {
	// 	hash ^= uint64(c)
	// 	fmt.Println(hash, c)
	// }
	m := newAtomicMapBuffer()

	// Store values in the map concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.store(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}(i)
	}
	wg.Wait()

	// Retrieve values from the map concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, ok := m.load(fmt.Sprintf("key%d", i))
			if ok {
				fmt.Printf("Value for key%d: %v\n", i, val)
			} else {
				fmt.Printf("Key%d not found\n", i)
			}
		}(i)
	}
	wg.Wait()
}
