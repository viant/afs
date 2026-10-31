package mem

import (
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/viant/afs/storage"
)

// A missing child's diagnostic calls its parent's URL after observing absence.
// Hold the first two such calls so both creators have observed the same missing
// child before either can publish it. This forces the lost-folder interleaving
// without sleeps or scheduler-dependent repetition.
type missingChildBarrier struct {
	storage.Object
	ctx             context.Context
	arrived         chan struct{}
	release         chan struct{}
	calls           int32
	once            sync.Once
	secondPublisher <-chan struct{}
}

func (b *missingChildBarrier) URL() string {
	call := atomic.AddInt32(&b.calls, 1)
	if call <= 2 {
		b.arrived <- struct{}{}
		select {
		case <-b.release:
		case <-b.ctx.Done():
		}
	}
	// The fourth root URL call is the second creator's post-absence
	// child URL. A public upload test may hold it until the first upload is
	// durable, making an old unconditional directory replacement lose data.
	if call == 4 && b.secondPublisher != nil {
		select {
		case <-b.secondPublisher:
		case <-b.ctx.Done():
		}
	}
	return b.Object.URL()
}
func (b *missingChildBarrier) unblock() { b.once.Do(func() { close(b.release) }) }
func holdMissingChild(t *testing.T, ctx context.Context, root *Folder) *missingChildBarrier {
	t.Helper()
	barrier := &missingChildBarrier{Object: root.Object, ctx: ctx, arrived: make(chan struct{}, 2), release: make(chan struct{})}
	root.Object = barrier
	t.Cleanup(barrier.unblock)
	return barrier
}
func (b *missingChildBarrier) waitForBoth(t *testing.T) {
	t.Helper()
	for i := 0; i < 2; i++ {
		select {
		case <-b.arrived:
		case <-b.ctx.Done():
			b.unblock()
			t.Fatal("both creators must observe the missing child before publication")
		}
	}
	b.unblock()
}
func TestFolderConcurrentCreationReturnsCanonicalChildAndRetainsFiles(t *testing.T) {
	// This deadline only bounds a broken test/implementation deadlock. Ordering is
	// established exclusively by the two arrival events and explicit release.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	root := NewFolder("mem://lookup-creation/", 0755)
	barrier := holdMissingChild(t, ctx, root)
	type result struct {
		folder *Folder
		err    error
	}
	results := make(chan result, 2)
	for _, name := range []string{"alpha", "bravo"} {
		go func(name string) {
			leaf, err := root.Folder("same/leaf", 0755)
			if err == nil {
				err = leaf.Put(NewFile("mem://lookup-creation/same/leaf/"+name+".txt", 0644, []byte(name), time.Time{}).Object)
			}
			results <- result{folder: leaf, err: err}
		}(name)
	}
	barrier.waitForBoth(t)
	first, second := <-results, <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("creation errors: %v / %v", first.err, second.err)
	}
	if first.folder != second.folder {
		t.Error("concurrent creators returned different child ownership; one tree can be replaced")
	}
	canonical, err := root.Folder("same/leaf", 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.folder != canonical || second.folder != canonical {
		t.Error("both creators must return the child reachable from the parent")
	}
	for _, name := range []string{"alpha", "bravo"} {
		file, err := root.File("same/leaf/" + name + ".txt")
		if err != nil {
			t.Errorf("published file %s was lost: %v", name, err)
			continue
		}
		reader := file.NewReader()
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil || string(data) != name {
			t.Errorf("file %s content=%q error=%v", name, data, err)
		}
	}
}
func TestConcurrentPublicUploadsRetainColdParentFiles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	service := New()
	root := service.(*manager).Root(ctx, "mem://upload-creation/")
	barrier := holdMissingChild(t, ctx, root)
	published := make(chan struct{})
	var publishOnce sync.Once
	releasePublisher := func() { publishOnce.Do(func() { close(published) }) }
	defer releasePublisher()
	barrier.secondPublisher = published
	results := make(chan error, 2)
	for _, name := range []string{"alpha", "bravo"} {
		go func(name string) {
			results <- service.Upload(ctx, "mem://upload-creation/same/leaf/"+name+".txt", 0644, strings.NewReader(name))
		}(name)
	}
	barrier.waitForBoth(t)
	select {
	case err := <-results:
		if err != nil {
			t.Error(err)
		}
	case <-ctx.Done():
		t.Fatal("first upload must complete before the second folder is published")
	}
	releasePublisher()
	select {
	case err := <-results:
		if err != nil {
			t.Error(err)
		}
	case <-ctx.Done():
		t.Fatal("second upload did not complete")
	}
	for _, name := range []string{"alpha", "bravo"} {
		reader, err := service.OpenURL(ctx, "mem://upload-creation/same/leaf/"+name+".txt")
		if err != nil {
			t.Errorf("successful upload %s is unreachable: %v", name, err)
			continue
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil || string(data) != name {
			t.Errorf("uploaded file %s content=%q error=%v", name, data, err)
		}
	}
}
