package nexmo

import "net/http"

type clientConfig struct {
	httpClient *http.Client
	apiKey     string
	apiSecret  string
	baseURL    string
}

func defaultClientConfig() *clientConfig {
	return &clientConfig{
		httpClient: http.DefaultClient,
		apiKey:     "",
		apiSecret:  "",
		baseURL:    "https://rest.nexmo.com",
	}
}
