package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Sub2apipayService calls sub2apipay HTTP API for recharge data.
type Sub2apipayService struct {
	apiURL    string
	jwtSecret string
	httpClient *http.Client
}

func NewSub2apipayService(apiURL, jwtSecret string) *Sub2apipayService {
	return &Sub2apipayService{
		apiURL:     apiURL,
		jwtSecret:  jwtSecret,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// generateAdminToken creates a short-lived admin JWT for sub2apipay auth.
func (s *Sub2apipayService) generateAdminToken() (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": 1,
		"role":    "admin",
		"exp":     now.Add(5 * time.Minute).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// doRequest creates and executes an authenticated request to sub2apipay.
func (s *Sub2apipayService) doRequest(ctx context.Context, path string, payload interface{}) (*http.Response, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("sub2apipay: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	token, err := s.generateAdminToken()
	if err != nil {
		return nil, fmt.Errorf("sub2apipay: generate token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sub2apipay: request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("sub2apipay: unexpected status %d", resp.StatusCode)
	}
	return resp, nil
}

// SumRechargeByHosts returns total recharge amount per srcHost from sub2apipay.
func (s *Sub2apipayService) SumRechargeByHosts(ctx context.Context, hosts []string) (map[string]float64, error) {
	if len(hosts) == 0 {
		return nil, nil
	}

	resp, err := s.doRequest(ctx, "/api/admin/orders/sum-by-host", map[string]interface{}{"hosts": hosts})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	resp, err := s.doRequest(ctx, "/api/admin/orders/list-by-host", map[string]interface{}{
		"hosts":     hosts,
		"page":      page,
		"page_size": pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Orders []struct {
			ID     string  `json:"id"`
			UserID int64   `json:"userId"`
			Amount float64 `json:"amount"`
			PaidAt *string `json:"paidAt"`
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
