package epay

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

// md5Sign 使用 MD5 签名：sign = md5(排序参数拼接 + key)
func md5Sign(params map[string]string, key string) string {
	content := buildSignContent(params)
	data := content + key
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// md5Verify 使用 MD5 验签
func md5Verify(params map[string]string, key string, sign string) bool {
	expected := md5Sign(params, key)
	return strings.EqualFold(expected, sign)
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
