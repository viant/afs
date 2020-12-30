package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/viant/afs"
	"github.com/viant/afs/base"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/url"
	"net/http"
	"time"
)

const (
	maxRetries = 10
)

//Counter represents a sync counter
type Counter struct {
	URL   string
	Data interface{} `json:",omitempty"`
	Count int
	fs    afs.Service
}

//Increment increments counter
func (g *Counter) Increment(ctx context.Context) (int, error) {
	return g.updateWithRetries(ctx, 1)
}

//Decrement decrements counter
func (g *Counter) Decrement(ctx context.Context) (int, error) {
	return g.updateWithRetries(ctx, -1)
}

func (g *Counter) updateWithRetries(ctx context.Context, delta int) (res int, err error) {
	retry := base.NewRetry()
	for ; retry.Count < maxRetries; retry.Count++ {
		if res, err = g.update(ctx, delta); err == nil {
			break
		}
		scheme := url.Scheme(g.URL, file.Scheme)
		code := g.fs.ErrorCode(scheme, err)
		if (code/100) == 5 || code == http.StatusPreconditionFailed || code == http.StatusTooManyRequests {
			time.Sleep(retry.Pause())
			continue
		}
		break
	}
	return res, err
}

func (g *Counter) update(ctx context.Context, delta int) (int, error) {
	generation := option.Generation{WhenMatch: true}
	ok, err := g.fs.Exists(ctx, g.URL, &generation, option.NewObjectKind(true))
	if err != nil {
		return 0, err
	}
	if !ok {
		g.Count = delta
		data, jErr := json.Marshal(g)
		if jErr != nil {
			return 0, jErr
		}
		scheme := url.Scheme(g.URL, file.Scheme)
		err := g.fs.Upload(ctx, g.URL, file.DefaultFileOsMode, bytes.NewReader(data), &generation)
		if err == nil {
			return g.Count, nil
		}
		if code := g.fs.ErrorCode(scheme, err); code != http.StatusPreconditionFailed {
			return 0, err
		}
	}

	data, err := g.fs.DownloadWithURL(ctx, g.URL, &generation)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(data, g)
	if err != nil {
		return 0, err
	}
	g.Count += delta
	data, _ = json.Marshal(g)
	err = g.fs.Upload(ctx, g.URL, file.DefaultFileOsMode, bytes.NewReader(data), &generation)
	return g.Count, err
}

//Delete deletes counter
func (g *Counter) Delete(ctx context.Context)  error {
	generation := &option.Generation{Generation:0, WhenMatch: true}
	ok,  _ := g.fs.Exists(ctx, g.URL, generation)
	if ! ok {
		return nil
	}
	return g.fs.Delete(ctx, g.URL, generation)
}


//NewCounter creates a fs based counter
func NewCounter(URL string, fs afs.Service) *Counter {
	return &Counter{URL: URL, fs: fs}
}
