package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPClient defines the interface for an HTTP client.
type HttpClient interface {
	Get(url string) (*http.Response, error)
}

type StandardHTTPClient struct {
	Client *http.Client
}

func (c *StandardHTTPClient) Get(url string) (*http.Response, error) {
	return c.Client.Get(url)
}

func SendHTTPRequest(ctx context.Context, method, url, token, apikey string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if apikey != "" {
		req.Header.Set("apikey", apikey)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	return resp, nil
}

func ParseJSONResponse(body io.Reader, target interface{}) error {
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	return nil
}
