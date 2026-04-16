package epay

import (
	"testing"
)

func TestVerifyNotifyAcceptsValidMD5Signature(t *testing.T) {
	client := &Client{key: "test_secret_key_123"}

	params := map[string]string{
		"pid":          "1008",
		"trade_no":     "2026040309051198850",
		"out_trade_no": "R20260403090514212704",
		"type":         "wxpay",
		"money":        "10.00",
		"trade_status": "TRADE_SUCCESS",
	}

	// Generate valid MD5 signature
	params["sign"] = md5Sign(params, client.key)

	if !client.VerifyNotify(params) {
		t.Fatal("expected VerifyNotify to succeed with valid MD5 signature")
	}
}

func TestVerifyNotifyRejectsInvalidSignature(t *testing.T) {
	client := &Client{key: "test_secret_key_123"}

	params := map[string]string{
		"pid":          "1008",
		"trade_no":     "2026040309051198850",
		"out_trade_no": "R20260403090514212704",
		"type":         "wxpay",
		"money":        "10.00",
		"trade_status": "TRADE_SUCCESS",
		"sign":         "00000000000000000000000000000000",
	}

	if client.VerifyNotify(params) {
		t.Fatal("expected VerifyNotify to reject invalid signature")
	}
}

func TestVerifyNotifyRejectsEmptySign(t *testing.T) {
	client := &Client{key: "test_secret_key_123"}

	params := map[string]string{
		"pid":   "1008",
		"money": "10.00",
	}

	if client.VerifyNotify(params) {
		t.Fatal("expected VerifyNotify to reject empty sign")
	}
}
