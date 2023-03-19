package whatsapp

import "net/http"

type clientConfig struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

func defaultClientConfig() *clientConfig {
	return &clientConfig{
		httpClient:  http.DefaultClient,
		accessToken: "",
		baseURL:     "https://graph.facebook.com",
	}
}
