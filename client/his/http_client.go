package his

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/khaimook/hospital-middleware/di/config"
)

type hisHTTPClient struct {
	cfg    config.AppConfig
	client *http.Client
}

func newHISHTTPClient(cfg config.AppConfig) *hisHTTPClient {
	return &hisHTTPClient{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.HISTimeout},
	}
}

func (h *hisHTTPClient) SearchPatient(ctx context.Context, hospitalCode, id string) (*HISPatient, error) {
	baseURL, ok := h.cfg.HISBaseURLs[hospitalCode]
	if !ok {
		return nil, ErrHISNotConfigured
	}

	reqURL := baseURL + "/patient/search/" + urlpkg.PathEscape(id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("his: create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("his: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPatientNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("his: unexpected status %d", resp.StatusCode)
	}

	var patient HISPatient
	if err := json.NewDecoder(resp.Body).Decode(&patient); err != nil {
		return nil, fmt.Errorf("his: decode response: %w", err)
	}

	if err := validatePatient(&patient); err != nil {
		return nil, err
	}

	return &patient, nil
}

func validatePatient(p *HISPatient) error {
	if p.Gender != "" && p.Gender != "M" && p.Gender != "F" {
		return fmt.Errorf("his: invalid gender %q", p.Gender)
	}
	if p.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", p.DateOfBirth); err != nil {
			return fmt.Errorf("his: invalid date_of_birth: %w", err)
		}
	}
	return nil
}
