package scratchpad

import (
	"context"
	"strings"
	"time"

	"github.com/viant/afs"
)

type userIDContextKey struct{}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ctx
	}
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	userID, _ := ctx.Value(userIDContextKey{}).(string)
	return strings.TrimSpace(userID)
}

func WithAFS(fs afs.Service) Option {
	return func(c *Config) {
		if fs != nil {
			c.FS = fs
		}
	}
}

func WithRootURI(template string) Option {
	return func(c *Config) {
		c.RootURITemplate = strings.TrimSpace(template)
	}
}

func WithUserID(userID string) Option {
	return func(c *Config) {
		c.UserID = strings.TrimSpace(userID)
	}
}

func WithUserIDProvider(provider UserIDProvider) Option {
	return func(c *Config) {
		c.UserIDProvider = provider
	}
}

func WithBasePath(basePath string) Option {
	return func(c *Config) {
		c.BasePath = strings.TrimSpace(basePath)
	}
}

func WithMacros(macros map[string]string) Option {
	return func(c *Config) {
		if len(macros) == 0 {
			return
		}
		if c.Macros == nil {
			c.Macros = make(map[string]string)
		}
		for k, v := range macros {
			c.Macros[k] = v
		}
	}
}

func WithAllowedTargetSchemes(schemes ...string) Option {
	return func(c *Config) {
		c.AllowedTargetSchemes = normalizeSchemes(schemes)
	}
}

func WithNow(now func() time.Time) Option {
	return func(c *Config) {
		if now != nil {
			c.Now = now
		}
	}
}

func WithConfig(config Config) Option {
	return func(c *Config) {
		if config.FS != nil {
			c.FS = config.FS
		}
		if strings.TrimSpace(config.RootURITemplate) != "" {
			c.RootURITemplate = strings.TrimSpace(config.RootURITemplate)
		}
		if strings.TrimSpace(config.UserID) != "" {
			c.UserID = strings.TrimSpace(config.UserID)
		}
		if config.UserIDProvider != nil {
			c.UserIDProvider = config.UserIDProvider
		}
		if strings.TrimSpace(config.BasePath) != "" {
			c.BasePath = strings.TrimSpace(config.BasePath)
		}
		if len(config.Macros) > 0 {
			WithMacros(config.Macros)(c)
		}
		if len(config.AllowedTargetSchemes) > 0 {
			c.AllowedTargetSchemes = normalizeSchemes(config.AllowedTargetSchemes)
		}
		if config.Now != nil {
			c.Now = config.Now
		}
	}
}

func normalizeSchemes(schemes []string) []string {
	if len(schemes) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var result []string
	for _, scheme := range schemes {
		scheme = strings.ToLower(strings.TrimSpace(scheme))
		if scheme == "" || seen[scheme] {
			continue
		}
		seen[scheme] = true
		result = append(result, scheme)
	}
	return result
}
