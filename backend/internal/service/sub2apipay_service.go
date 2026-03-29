package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Sub2apipayService calls sub2apipay HTTP API for recharge data.
type Sub2apipayService struct {
	apiURL     string
	adminToken string
	httpClient *http.Client
}

func NewSub2apipayService(apiURL, adminToken string) *Sub2apipayService {
	return &Sub2apipayService{
		apiURL:     apiURL,
		adminToken: adminToken,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SumRechargeByHosts returns total recharge amount per srcHost from sub2apipay.
func (s *Sub2apipayService) SumRechargeByHosts(ctx context.Context, hosts []string) (map[string]float64, error) {
	if len(hosts) == 0 {
		return nil, nil
	}

	body, _ := json.Marshal(map[string]interface{}{"hosts": hosts})
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL+"/api/admin/orders/sum-by-host", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("sub2apipay: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sub2apipay: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sub2apipay: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Totals map[string]float64 `json:"totals"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("sub2apipay: decode response: %w", err)
	}

	return result.Totals, nil
}
