package provider

import (
	"time"

	"search-engine/infra/httpclient"

	"go.uber.org/zap"
)

func init() {
	Register("http_json", createHTTPJSONProvider)
	Register("http_xml", createHTTPXMLProvider)
}

func createHTTPJSONProvider(name, url string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) ContentProvider {
	provider, _ := NewHTTPProvider(name, url, "json", timeout, client, logger)
	return provider
}

func createHTTPXMLProvider(name, url string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) ContentProvider {
	provider, _ := NewHTTPProvider(name, url, "xml", timeout, client, logger)
	return provider
}
