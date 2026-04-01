package epay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
)

// parsePrivateKey 解析 RSA 私钥（支持裸 base64 和 PEM 格式）
func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	key = strings.TrimSpace(key)
	if !strings.HasPrefix(key, "-----") {
		key = "-----BEGIN PRIVATE KEY-----\n" + wrapBase64(key) + "\n-----END PRIVATE KEY-----"
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("epay: failed to decode private key PEM")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// fallback to PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	rsaKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("epay: private key is not RSA")
	}
	return rsaKey, nil
}

// parsePublicKey 解析 RSA 公钥（支持裸 base64 和 PEM 格式）
func parsePublicKey(key string) (*rsa.PublicKey, error) {
	key = strings.TrimSpace(key)
	if !strings.HasPrefix(key, "-----") {
		key = "-----BEGIN PUBLIC KEY-----\n" + wrapBase64(key) + "\n-----END PUBLIC KEY-----"
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("epay: failed to decode public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("epay: parse public key: %w", err)
	}
	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("epay: public key is not RSA")
	}
	return rsaKey, nil
}

// rsaSign 使用商户私钥签名（SHA256WithRSA）
func rsaSign(privateKey *rsa.PrivateKey, data string) (string, error) {
	h := sha256.Sum256([]byte(data))
	sig, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", fmt.Errorf("epay: rsa sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// rsaVerify 使用平台公钥验签（SHA256WithRSA）
func rsaVerify(publicKey *rsa.PublicKey, data, sign string) bool {
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	h := sha256.Sum256([]byte(data))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, h[:], sig) == nil
}

// buildSignContent 构造待签名字符串：按 key 排序，跳过空值和 sign/sign_type
func buildSignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, "&")
}

func wrapBase64(s string) string {
	var buf strings.Builder
	for i, c := range s {
		if i > 0 && i%64 == 0 {
			buf.WriteByte('\n')
		}
		buf.WriteRune(c)
	}
	return buf.String()
}
