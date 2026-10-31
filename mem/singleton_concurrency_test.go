package mem

import (
	"sync"
	"testing"

	"github.com/viant/afs/storage"
)

func TestSingletonConcurrentInitialization(t *testing.T) {
	singletonMu.Lock()
	original := singleton
	singleton = nil
	singletonMu.Unlock()
	defer func() { singletonMu.Lock(); singleton = original; singletonMu.Unlock() }()
	const count = 64
	values := make([]storage.Manager, count)
	ready := make(chan struct{})
	var group sync.WaitGroup
	for i := range values {
		group.Add(1)
		go func(index int) { defer group.Done(); <-ready; values[index] = Singleton() }(i)
	}
	close(ready)
	group.Wait()
	if values[0] == nil {
		t.Fatal("singleton was not initialized")
	}
	for _, value := range values {
		if value != values[0] {
			t.Fatal("concurrent callers received different memory filesystems")
		}
	}
}

func TestSingletonConcurrentReset(t *testing.T) {
	const count = 64
	var group sync.WaitGroup
	for i := 0; i < count; i++ {
		group.Add(1)
		go func(reset bool) {
			defer group.Done()
			if reset {
				ResetSingleton()
			}
			if Singleton() == nil {
				t.Error("reset exposed an uninitialized filesystem")
			}
		}(i%4 == 0)
	}
	group.Wait()
}
