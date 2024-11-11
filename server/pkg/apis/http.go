package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sysarmor/guard/server/pkg/apis/dto"
	"github.com/sysarmor/guard/server/pkg/signature"
)

type HTTPGuard struct {
	tgt        *url.URL
	nodeID     string
	nodeSecret string

	client *http.Client
}

const (
	HeaderTimestamp = "X-Timestamp"
	HeaderSignature = "X-Signature"
)

func NewHTTPGuard(address string, nodeID, nodeSecret string) (*HTTPGuard, error) {
	url, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}

	return &HTTPGuard{
		tgt:        url,
		nodeID:     nodeID,
		nodeSecret: nodeSecret,

		client: http.DefaultClient,
	}, nil
}

func (g *HTTPGuard) GetCA(ctx context.Context) (string, error) {
	url := *g.tgt
	url.Path = "/api/v1/guard/ca"
	url.RawQuery = "nodeID=" + g.nodeID

	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url,
	}

	resp, err := g.do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}

	return decodeResp[string](g.nodeSecret, resp)
}

func (g *HTTPGuard) GetPrincipals(ctx context.Context) ([]*dto.Principals, error) {
	url := *g.tgt
	url.Path = "/api/v1/guard/principals"
	url.RawQuery = "nodeID=" + g.nodeID

	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url,
	}

	resp, err := g.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	return decodeResp[[]*dto.Principals](g.nodeSecret, resp)
}

func (g *HTTPGuard) GetKRL(ctx context.Context) (string, error) {
	url := *g.tgt
	url.Path = "/api/v1/guard/krl"
	url.RawQuery = "nodeID=" + g.nodeID

	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url,
	}

	resp, err := g.do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}

	return decodeResp[string](g.nodeSecret, resp)
}

func (g *HTTPGuard) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	g.signTimestamp(req)
	req = req.WithContext(ctx)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	return resp, nil
}

func (g *HTTPGuard) setTimestamp(req *http.Request, timestamp string) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set(HeaderTimestamp, timestamp)
}

func (g *HTTPGuard) signTimestamp(req *http.Request) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	g.setTimestamp(req, timestamp)
	sign := signature.SimpleSignature(signature.SimpleString(timestamp), []byte(g.nodeSecret))
	req.Header.Set(HeaderSignature, sign)
}

func validateSignature(body []byte, secret, expertedSign string) error {
	sign := signature.SimpleSignature(signature.SimpleString(body), []byte(secret))
	if sign != expertedSign {
		return fmt.Errorf("signature is invalid")
	}

	return nil
}

// FIXME: decode response should not care about secret
func decodeResp[T any](secret string, resp *http.Response) (T, error) {
	var t T
	if resp.StatusCode != http.StatusOK {
		return t, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return t, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := validateSignature(body, secret, resp.Header.Get(HeaderSignature)); err != nil {
		return t, fmt.Errorf("failed to validate response: %w", err)
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		return t, fmt.Errorf("failed to decode response: %w", err)
	}

	return t, nil
}
