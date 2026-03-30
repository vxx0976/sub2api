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

// ListRechargesByHosts returns paginated recharge orders filtered by srcHost from sub2apipay.
func (s *Sub2apipayService) ListRechargesByHosts(ctx context.Context, hosts []string, page, pageSize int) ([]*RechargeDetailRecord, int, error) {
	if len(hosts) == 0 {
		return nil, 0, nil
	}

	body, _ := json.Marshal(map[string]interface{}{
		"hosts":     hosts,
		"page":      page,
		"page_size": pageSize,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL+"/api/admin/orders/list-by-host", bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("sub2apipay: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("sub2apipay: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("sub2apipay: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Orders []struct {
			ID      string   `json:"id"`
			UserID  int64    `json:"userId"`
			Amount  float64  `json:"amount"`
			PaidAt  *string  `json:"paidAt"`
			SrcHost *string  `json:"srcHost"`
		} `json:"orders"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("sub2apipay: decode response: %w", err)
	}

	records := make([]*RechargeDetailRecord, 0, len(result.Orders))
	for _, o := range result.Orders {
		r := &RechargeDetailRecord{
			UserID:       o.UserID,
			OrderNo:      o.ID,
			CreditAmount: o.Amount,
		}
		if o.PaidAt != nil {
			if t, err := time.Parse(time.RFC3339Nano, *o.PaidAt); err == nil {
				r.PaidAt = t
			}
		}
		records = append(records, r)
	}

	return records, result.Total, nil
}
