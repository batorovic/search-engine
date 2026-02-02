package provider

import (
	"fmt"
	"sync"
	"time"

	"search-engine/infra/httpclient"

	"go.uber.org/zap"
)

type ProviderFactory func(name, url string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) ContentProvider

type Registry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

var globalRegistry = &Registry{
	factories: make(map[string]ProviderFactory),
}

func Register(formatType string, factory ProviderFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[formatType] = factory
}

func GetFactory(formatType string) (ProviderFactory, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	factory, exists := globalRegistry.factories[formatType]
	if !exists {
		return nil, fmt.Errorf("unknown provider format: %s", formatType)
	}
	return factory, nil
}

func CreateProvider(formatType, name, url string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) (ContentProvider, error) {
	factory, err := GetFactory(formatType)
	if err != nil {
		return nil, err
	}
	return factory(name, url, timeout, client, logger), nil
}

func ListRegisteredFormats() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	formats := make([]string, 0, len(globalRegistry.factories))
	for format := range globalRegistry.factories {
		formats = append(formats, format)
	}
	return formats
}
