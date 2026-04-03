package epay

import (
	"crypto/rand"
	"crypto/rsa"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestVerifyNotifyAcceptsURLDecodedPlusInSign(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	client := &Client{
		publicKey:  &privateKey.PublicKey,
		privateKey: privateKey,
	}

	params := map[string]string{
		"pid":          "1008",
		"trade_no":     "2026040309051198850",
		"out_trade_no": "R20260403090514212704",
		"type":         "wxpay",
		"money":        "10.00",
		"trade_status": "TRADE_SUCCESS",
		"timestamp":    strconv.FormatInt(time.Now().Unix(), 10),
		"sign_type":    "RSA",
	}

	foundPlus := false
	for i := 0; i < 512; i++ {
		params["buyer"] = "buyer" + strconv.Itoa(i)
		sign, err := rsaSign(privateKey, buildSignContent(params))
		if err != nil {
			t.Fatalf("sign params: %v", err)
		}
		if !strings.Contains(sign, "+") {
			continue
		}

		foundPlus = true
		params["sign"] = strings.ReplaceAll(sign, "+", " ")
		if !client.VerifyNotify(params) {
			t.Fatalf("expected notify verification to succeed when sign '+' becomes space")
		}
		break
	}

	if !foundPlus {
		t.Fatal("failed to generate a signature containing '+' for URL-decoding regression test")
	}
}
