package tar

import (
	"context"
	"fmt"
	"github.com/viant/afs/base"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
)

type manager struct {
	*base.Manager
}

func (m *manager) provider(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	var manager storage.Manager
	option.Assign(options, &manager)
	options = m.Options(options)
	URL := url.SchemeExtensionURL(baseURL)
	if URL == "" {
		return nil, fmt.Errorf("extneded URL was empty: %v", baseURL)
	}
	if manager == nil {
		return nil, fmt.Errorf("manager for URL was empty: %v", URL)
	}
	return newStorager(ctx, baseURL, manager)

}

func newManager(options ...storage.Option) *manager {
	result := &manager{}
	baseMgr := base.New(result, Scheme, result.provider, options)
	result.Manager = baseMgr
	return result
}

//New creates zip manager
func New(options ...storage.Option) storage.Manager {
	return newManager(options...)
}
