package resource_test

import (
	"context"
	"embed"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/viant/afs"
	afsio "github.com/viant/afs/adapter/io"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/resource"
	"github.com/viant/afs/storage"
)

//go:embed testdata
var embedded embed.FS

func service(t *testing.T, cfg resource.Config) afs.Service {
	t.Helper()
	p, err := resource.NewProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return afs.NewWithProviders(map[string]afs.Provider{cfg.Scheme: p})
}

func content(t *testing.T, svc afs.Service, uri, want string) {
	t.Helper()
	data, err := svc.DownloadWithURL(context.Background(), uri)
	if err != nil || string(data) != want {
		t.Fatalf("read %s: %q, %v; want %q", uri, data, err, want)
	}
}

func TestEmbeddedExactRootAndNativeIO(t *testing.T) {
	ctx := context.Background()
	svc := service(t, resource.Config{Scheme: "skill", Roots: map[string]resource.Root{
		"datly-reader": {FS: embedded, Path: "testdata/skills/datly-reader"},
		"parent":       {FS: embedded, Path: "testdata/skills"},
	}})
	uri := "skill://datly-reader/SKILL.md"
	want := "---\nname: datly-reader\n---\nRead Datly resources.\n"
	content(t, svc, uri, want)
	content(t, svc, "skill://datly-reader/references/guide.md", "Supporting guide.\n")
	content(t, svc, "skill://parent/SKILL.md", "Wrong parent root.\n")
	info, err := svc.Object(ctx, uri)
	if err != nil || info.URL() != uri || info.Name() != "SKILL.md" || info.Size() != int64(len(want)) || info.IsDir() {
		t.Fatalf("metadata: %#v %v", info, err)
	}
	dir, err := svc.Object(ctx, "skill://datly-reader/")
	if err != nil || !dir.IsDir() || dir.URL() != "skill://datly-reader/" {
		t.Fatalf("directory: %v %v", dir, err)
	}
	list, err := svc.List(ctx, "skill://datly-reader/")
	if err != nil || len(list) != 3 || list[1].URL() != uri || list[2].URL() != "skill://datly-reader/references" {
		t.Fatalf("list: %v %v", list, err)
	}
	filtered, err := svc.List(ctx, "skill://datly-reader/", option.Match(func(_ string, info os.FileInfo) bool { return !info.IsDir() }))
	if err != nil || len(filtered) != 2 {
		t.Fatalf("filtered list: %v %v", filtered, err)
	}
	single, err := svc.List(ctx, uri)
	if err != nil || len(single) != 1 || single[0].URL() != uri {
		t.Fatalf("file listing: %v %v", single, err)
	}
	adapter := afsio.NewFS(ctx, svc, "skill://datly-reader")
	if err := fstest.TestFS(adapter, "SKILL.md", "references/guide.md"); err != nil {
		t.Fatal(err)
	}
	if exists, err := svc.Exists(ctx, "skill://datly-reader/missing"); exists || err != nil {
		t.Fatalf("missing exists: %v %v", exists, err)
	}
	if _, err := svc.Object(ctx, "skill://datly-reader/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	if _, err := svc.OpenURL(ctx, "skill://datly-reader/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	for name, op := range map[string]func() error{
		"upload": func() error { return svc.Upload(ctx, uri, 0600, strings.NewReader("x")) },
		"create": func() error { return svc.Create(ctx, uri, 0600, false) },
		"delete": func() error { return svc.Delete(ctx, uri) },
		"writer": func() error { _, err := svc.NewWriter(ctx, uri, 0600); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := op(); !errors.Is(err, resource.ErrReadOnly) {
				t.Fatal(err)
			}
		})
	}
}

func TestLocalAndMemCapabilities(t *testing.T) {
	ctx := context.Background()
	backend := afs.New()
	defer backend.CloseAll()
	local := t.TempDir()
	memRoot := "mem://localhost/" + t.Name()
	for _, root := range []string{"file://localhost" + local, memRoot} {
		t.Run(strings.Split(root, ":")[0], func(t *testing.T) {
			svc := service(t, resource.Config{Scheme: "assets", Service: backend, Roots: map[string]resource.Root{"docs": {URL: root}}})
			uri := "assets://docs/sub/file.txt"
			if err := svc.Upload(ctx, uri, 0600, strings.NewReader("native")); err != nil {
				t.Fatal(err)
			}
			content(t, svc, uri, "native")
			content(t, backend, root+"/sub/file.txt", "native")
			encoded := "assets://docs/sub/hello%20%E4%B8%96%E7%95%8C.txt"
			if err := svc.Upload(ctx, encoded, 0600, strings.NewReader("unicode")); err != nil {
				t.Fatal(err)
			}
			content(t, svc, encoded, "unicode")
			native, err := backend.Object(ctx, root+"/sub/file.txt")
			if err != nil {
				t.Fatal(err)
			}
			obj, err := svc.Object(ctx, uri)
			if err != nil || obj.URL() != uri || obj.Size() != native.Size() || obj.Mode() != native.Mode() || !obj.ModTime().Equal(native.ModTime()) {
				t.Fatalf("object: %v %v", obj, err)
			}
			entries, err := svc.List(ctx, "assets://docs/sub")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range entries {
				if !strings.HasPrefix(e.URL(), "assets://docs/") {
					t.Fatal(e.URL())
				}
				if e.URL() == uri {
					found = true
				}
			}
			if !found {
				t.Fatal("file missing from listing")
			}
			if err := svc.Create(ctx, "assets://docs/empty", 0700, true); err != nil {
				t.Fatal(err)
			}
			empty, err := svc.Object(ctx, "assets://docs/empty")
			if err != nil || !empty.IsDir() {
				t.Fatalf("empty directory: %v %v", empty, err)
			}
			w, err := svc.NewWriter(ctx, "assets://docs/written.txt", 0600)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = w.Write([]byte("writer")); err != nil {
				t.Fatal(err)
			}
			if err = w.Close(); err != nil {
				t.Fatal(err)
			}
			content(t, svc, "assets://docs/written.txt", "writer")
			if err = svc.Copy(ctx, uri, "assets://docs/copy.txt"); err != nil {
				t.Fatal(err)
			}
			content(t, svc, "assets://docs/copy.txt", "native")
			if err = svc.Move(ctx, "assets://docs/copy.txt", "assets://docs/moved.txt"); err != nil {
				t.Fatal(err)
			}
			content(t, svc, "assets://docs/moved.txt", "native")
			if err = svc.Delete(ctx, "assets://docs/moved.txt"); err != nil {
				t.Fatal(err)
			}
			if exists, err := svc.Exists(ctx, "assets://docs/moved.txt"); exists || err != nil {
				t.Fatalf("deleted: %v %v", exists, err)
			}
			if _, err := svc.OpenURL(ctx, "assets://docs/missing.txt"); err == nil {
				t.Fatal("missing opened")
			}
			if err = svc.Upload(ctx, "assets://docs/link", 0600, strings.NewReader(""), &object.Link{Linkname: filepath.Join(t.TempDir(), "escape")}); !errors.Is(err, resource.ErrBoundary) {
				t.Fatal(err)
			}
		})
	}
}

func TestEncodingAndBoundaries(t *testing.T) {
	ctx := context.Background()
	svc := service(t, resource.Config{Scheme: "skill", Roots: map[string]resource.Root{"datly-reader": {FS: fstest.MapFS{"hello world.txt": {Data: []byte("space")}, "SKILL.md": {Data: []byte("ok")}}}}})
	content(t, svc, "skill://datly-reader/hello%20world.txt", "space")
	entries, err := svc.List(ctx, "skill://datly-reader/")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() == "hello world.txt" {
			if e.URL() != "skill://datly-reader/hello%20world.txt" {
				t.Fatal(e.URL())
			}
			content(t, svc, e.URL(), "space")
		}
	}
	for _, uri := range []string{
		"skill://datly-reader/../SKILL.md", "skill://datly-reader/%2e%2e/SKILL.md", "skill://datly-reader/%252e%252e/SKILL.md",
		"skill://datly-reader/a%2f..%2fSKILL.md", "skill://datly-reader/a%5c..%5cSKILL.md", "skill://datly-reader/a\\..\\SKILL.md",
		"skill://datly-reader/%00", "skill://datly-reader/%zz", "skill://datly-reader/SKILL.md?x=1", "skill://datly-reader/SKILL.md#x", "skill://datly-reader/./SKILL.md",
	} {
		t.Run(uri, func(t *testing.T) {
			if _, err := svc.OpenURL(ctx, uri); !errors.Is(err, resource.ErrBoundary) {
				t.Fatalf("%s: %v", uri, err)
			}
		})
	}
	for _, uri := range []string{"skill://unknown/SKILL.md", "skill://datly-reader.evil/SKILL.md"} {
		if _, err := svc.OpenURL(ctx, uri); !errors.Is(err, resource.ErrUnknownNamespace) {
			t.Fatal(err)
		}
	}
}

func TestLocalSymlinkBoundaries(t *testing.T) {
	ctx := context.Background()
	local, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(local, "escape")); err != nil {
		t.Fatal(err)
	}
	svc := service(t, resource.Config{Scheme: "skill", Service: afs.New(), Roots: map[string]resource.Root{"datly-reader": {URL: "file://localhost" + local}}})
	uri := "skill://datly-reader/escape/secret"
	for name, op := range map[string]func() error{
		"open":   func() error { _, err := svc.OpenURL(ctx, uri); return err },
		"object": func() error { _, err := svc.Object(ctx, uri); return err },
		"list":   func() error { _, err := svc.List(ctx, "skill://datly-reader/"); return err },
		"upload": func() error { return svc.Upload(ctx, uri, 0600, strings.NewReader("bad")) },
		"create": func() error { return svc.Create(ctx, "skill://datly-reader/escape/new/dir", 0700, true) },
		"delete": func() error { return svc.Delete(ctx, uri) },
		"writer": func() error { _, err := svc.NewWriter(ctx, uri, 0600); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := op(); !errors.Is(err, resource.ErrBoundary) {
				t.Fatal(err)
			}
		})
	}
	if data, err := os.ReadFile(filepath.Join(outside, "secret")); err != nil || string(data) != "outside" {
		t.Fatalf("outside modified: %q %v", data, err)
	}
	if err := os.Symlink(filepath.Join(local, "escape", "secret"), filepath.Join(local, "direct")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.OpenURL(ctx, "skill://datly-reader/direct"); !errors.Is(err, resource.ErrBoundary) {
		t.Fatal(err)
	}
}

func TestImmutableConcurrentMappings(t *testing.T) {
	ctx := context.Background()
	makeProvider := func(scheme, value string) afs.Provider {
		cfg := resource.Config{Scheme: scheme, Roots: map[string]resource.Root{"same": {FS: fstest.MapFS{"SKILL.md": {Data: []byte(value)}}}}}
		p, err := resource.NewProvider(cfg)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Roots["same"] = resource.Root{FS: fstest.MapFS{}}
		return p
	}
	a, b := makeProvider("skill", "a"), makeProvider("skill", "b")
	providers := map[string]afs.Provider{"skill": a, "manual": makeProvider("manual", "manual")}
	first := afs.NewWithProviders(providers)
	providers["skill"] = b
	second := afs.NewWithProviders(map[string]afs.Provider{"skill": b})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tc := range []struct {
				svc       afs.Service
				uri, want string
			}{{first, "skill://same/SKILL.md", "a"}, {second, "skill://same/SKILL.md", "b"}, {first, "manual://same/SKILL.md", "manual"}} {
				data, err := tc.svc.DownloadWithURL(ctx, tc.uri)
				if err != nil || string(data) != tc.want {
					t.Errorf("isolation: %q %v", data, err)
				}
			}
		}()
	}
	wg.Wait()
	if _, err := afs.GetRegistry().Get("skill"); err == nil {
		t.Fatal("private mapping leaked globally")
	}
	m, err := a(resource.Config{Roots: map[string]resource.Root{"same": {FS: fstest.MapFS{}}}})
	if err != nil {
		t.Fatal(err)
	}
	reader, err := m.OpenURL(ctx, "skill://same/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil || string(data) != "a" {
		t.Fatalf("factory override: %q %v", data, err)
	}
}

func TestCancellation(t *testing.T) {
	active, cancel := context.WithCancel(context.Background())
	svc := service(t, resource.Config{Scheme: "skill", Roots: map[string]resource.Root{"same": {FS: fstest.MapFS{"SKILL.md": {Data: []byte("data")}}}}})
	reader, err := svc.OpenURL(active, "skill://same/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	cancel()
	if _, err = reader.Read(make([]byte, 1)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, op := range []func() error{
		func() error { _, err := svc.OpenURL(active, "skill://same/SKILL.md"); return err },
		func() error { _, err := svc.Object(active, "skill://same/SKILL.md"); return err },
		func() error { _, err := svc.List(active, "skill://same/"); return err },
		func() error { _, err := svc.Exists(active, "skill://same/SKILL.md"); return err },
		func() error { return svc.Upload(active, "skill://same/new", 0600, strings.NewReader("")) },
	} {
		if err := op(); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
}

func TestInvalidConfig(t *testing.T) {
	for name, cfg := range map[string]resource.Config{
		"scheme":           {Scheme: "bad/scheme"},
		"authority":        {Scheme: "skill", Roots: map[string]resource.Root{"user@host": {FS: embedded}}},
		"no root":          {Scheme: "skill", Roots: map[string]resource.Root{"a": {}}},
		"both roots":       {Scheme: "skill", Roots: map[string]resource.Root{"a": {FS: embedded, URL: "mem://localhost/a"}}},
		"subtree escape":   {Scheme: "skill", Roots: map[string]resource.Root{"a": {FS: embedded, Path: "../a"}}},
		"no backend":       {Scheme: "skill", Roots: map[string]resource.Root{"a": {URL: "mem://localhost/a"}}},
		"relative URL":     {Scheme: "skill", Service: afs.New(), Roots: map[string]resource.Root{"a": {URL: "../a"}}},
		"recursive URL":    {Scheme: "skill", Service: afs.New(), Roots: map[string]resource.Root{"a": {URL: "skill://other/a"}}},
		"URL traversal":    {Scheme: "skill", Service: afs.New(), Roots: map[string]resource.Root{"a": {URL: "mem://localhost/a/../b"}}},
		"remote file host": {Scheme: "skill", Service: afs.New(), Roots: map[string]resource.Root{"a": {URL: "file://remote/tmp"}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resource.NewProvider(cfg); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}

func TestCrossRootCopyAndModifier(t *testing.T) {
	ctx := context.Background()
	backend := afs.New()
	svc := service(t, resource.Config{Scheme: "resource", Service: backend, Roots: map[string]resource.Root{
		"embedded": {FS: embedded, Path: "testdata/skills/datly-reader"},
		"writable": {URL: "mem://localhost/" + t.Name()},
	}})
	if err := svc.Copy(ctx, "resource://embedded/references", "resource://writable/references"); err != nil {
		t.Fatal(err)
	}
	content(t, svc, "resource://writable/references/guide.md", "Supporting guide.\n")
	calls := 0
	modifier := option.Modifier(func(_ string, info os.FileInfo, reader io.ReadCloser) (os.FileInfo, io.ReadCloser, error) {
		calls++
		return info, reader, nil
	})
	data, err := svc.DownloadWithURL(ctx, "resource://writable/references/guide.md", modifier)
	if err != nil || string(data) != "Supporting guide.\n" || calls != 1 {
		t.Fatalf("modifier called %d times; %q %v", calls, data, err)
	}
}

func TestURLCancellation(t *testing.T) {
	backend := afs.New()
	uri := "files://docs/data.txt"
	svc := service(t, resource.Config{Scheme: "files", Service: backend, Roots: map[string]resource.Root{"docs": {URL: "file://localhost" + t.TempDir()}}})
	ctx, cancel := context.WithCancel(context.Background())
	if err := svc.Upload(ctx, uri, 0600, strings.NewReader("data")); err != nil {
		t.Fatal(err)
	}
	reader, err := svc.OpenURL(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	writer, err := svc.NewWriter(ctx, "files://docs/new.txt", 0600)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := reader.Read(make([]byte, 1)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("x")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := writer.Close(); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, op := range []func() error{
		func() error { return svc.Create(ctx, "files://docs/no-create", 0700, true) },
		func() error { return svc.Delete(ctx, uri) },
		func() error { _, err := svc.NewWriter(ctx, "files://docs/no-write", 0600); return err },
		func() error { _, err := svc.List(ctx, "files://docs/"); return err },
	} {
		if err := op(); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	content(t, svc, uri, "data")
}

// Unknown URI authorities must fail before any native provider is called.
type observedBackend struct {
	afs.Service
	calls int
}

func (s *observedBackend) OpenURL(context.Context, string, ...storage.Option) (io.ReadCloser, error) {
	s.calls++
	return nil, errors.New("backend called")
}
func TestUnknownNamespaceDoesNotResolveAuthority(t *testing.T) {
	backend := &observedBackend{}
	svc := service(t, resource.Config{Scheme: "skill", Service: backend, Roots: map[string]resource.Root{"datly-reader": {URL: "mem://localhost/root"}}})
	for _, uri := range []string{"skill://unconfigured.example/SKILL.md", "skill://datly-reader.evil/SKILL.md"} {
		if _, err := svc.OpenURL(context.Background(), uri); !errors.Is(err, resource.ErrUnknownNamespace) {
			t.Fatal(err)
		}
	}
	if backend.calls != 0 {
		t.Fatalf("unexpected backend resolution: %d", backend.calls)
	}
}

func TestConcurrentNativeRootIsolation(t *testing.T) {
	ctx := context.Background()
	backend := afs.New()
	roots := []string{"mem://localhost/" + t.Name() + "/one", "mem://localhost/" + t.Name() + "/two"}
	services := make([]afs.Service, 2)
	for i, root := range roots {
		if err := backend.Upload(ctx, root+"/SKILL.md", 0600, strings.NewReader(root)); err != nil {
			t.Fatal(err)
		}
		services[i] = service(t, resource.Config{Scheme: "skill", Service: backend, Roots: map[string]resource.Root{"same": {URL: root}}})
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			j := i % 2
			data, err := services[j].DownloadWithURL(ctx, "skill://same/SKILL.md")
			if err != nil || string(data) != roots[j] {
				t.Errorf("native isolation: %q %v", data, err)
			}
		}(i)
	}
	wg.Wait()
}
