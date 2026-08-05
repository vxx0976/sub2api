package repository

import (
	"context"
	"errors"
	"fmt"

	captcha "github.com/alibabacloud-go/captcha-20230305/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const aliyunCaptchaTimeoutMillis = 10_000

type aliyunCaptchaVerifier struct {
	protocol      string // "HTTPS"；测试注入 "HTTP" 指向 httptest.Server
	timeoutMillis int
}

func NewAliyunCaptchaVerifier() service.AliyunCaptchaVerifier {
	return &aliyunCaptchaVerifier{
		protocol:      "HTTPS",
		timeoutMillis: aliyunCaptchaTimeoutMillis,
	}
}

// VerifyCaptcha 调用阿里云验证码 2.0 VerifyIntelligentCaptcha。
// AK/SK 是可热更的后台设置，每次调用按当前凭证新建 client。
func (v *aliyunCaptchaVerifier) VerifyCaptcha(ctx context.Context, cred service.AliyunCaptchaCredentials, captchaVerifyParam string) (*service.AliyunCaptchaVerifyResult, error) {
	client, err := captcha.NewClient(&openapiutil.Config{
		AccessKeyId:     dara.String(cred.AccessKeyID),
		AccessKeySecret: dara.String(cred.AccessKeySecret),
		Endpoint:        dara.String(cred.Endpoint),
		Protocol:        dara.String(v.protocol),
		ConnectTimeout:  dara.Int(v.timeoutMillis),
		ReadTimeout:     dara.Int(v.timeoutMillis),
	})
	if err != nil {
		return nil, fmt.Errorf("create aliyun captcha client: %w", err)
	}

	request := &captcha.VerifyIntelligentCaptchaRequest{
		CaptchaVerifyParam: dara.String(captchaVerifyParam),
		SceneId:            dara.String(cred.SceneID),
	}

	response, err := client.VerifyIntelligentCaptchaWithContext(ctx, request, &dara.RuntimeOptions{})
	if err != nil {
		return nil, normalizeAliyunCaptchaError(err)
	}

	result := &service.AliyunCaptchaVerifyResult{}
	if body := response.Body; body != nil && body.Result != nil {
		result.VerifyResult = dara.BoolValue(body.Result.VerifyResult)
		result.VerifyCode = dara.StringValue(body.Result.VerifyCode)
	}
	return result, nil
}

// normalizeAliyunCaptchaError 把 SDK 的两种错误类型归一化为 service.AliyunCaptchaAPIError，
// 其余错误（网络/超时等）原样返回。
// 传输层失败（连接被拒、超时等）也会被 SDK 包成 SDKError，但 Code 为空、
// Message 里只有合成的 "code: 503 ..."。这类错误不能归一化为 API 错误：
// 「验证码服务不可达」与「验证码校验被拒」的处置策略不同，混为一谈会让
// 调用方把网络故障当成用户验证失败。仅在带真实 Code 时才视为 API 错误。
func normalizeAliyunCaptchaError(err error) error {
	var teaErr *tea.SDKError
	if errors.As(err, &teaErr) {
		if code := tea.StringValue(teaErr.Code); isAliyunCaptchaAPICode(code) {
			return &service.AliyunCaptchaAPIError{
				Code:    code,
				Message: tea.StringValue(teaErr.Message),
			}
		}
		return err
	}
	var daraErr *dara.SDKError
	if errors.As(err, &daraErr) {
		if code := dara.StringValue(daraErr.Code); isAliyunCaptchaAPICode(code) {
			return &service.AliyunCaptchaAPIError{
				Code:    code,
				Message: dara.StringValue(daraErr.Message),
			}
		}
		return err
	}
	return err
}

// isAliyunCaptchaAPICode 判断 SDKError 是否真的携带服务端返回的错误码。
// 传输层失败（连接被拒、超时等）同样会被 SDK 包成 SDKError，但此时 Code 来自
// 空指针，被 SDK 格式化成字面量 "<nil>"（Message 里也只有合成的 "code: 503 ..."）。
func isAliyunCaptchaAPICode(code string) bool {
	return code != "" && code != "<nil>"
}
