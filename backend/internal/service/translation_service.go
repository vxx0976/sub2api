package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	defaultTranslationBaseURL   = "https://api.openai.com"
	defaultTranslationModel     = "gpt-4o-mini"
	defaultTranslationTimeoutMS = 30000
	maxTranslationTimeoutMS     = 120000
	maxTranslationBatch         = 32
	maxTranslationCharsPerItem  = 4000
)

// TranslationConfig 一键翻译配置（OpenAI Chat Completions 兼容）
type TranslationConfig struct {
	BaseURL   string `json:"base_url"`
	Model     string `json:"model"`
	APIKey    string `json:"api_key,omitempty"`
	TimeoutMS int    `json:"timeout_ms"`
}

// TranslationConfigView 对外返回的配置视图（脱敏 APIKey）
type TranslationConfigView struct {
	BaseURL          string `json:"base_url"`
	Model            string `json:"model"`
	TimeoutMS        int    `json:"timeout_ms"`
	APIKeyConfigured bool   `json:"api_key_configured"`
	APIKeyMasked     string `json:"api_key_masked"`
}

// UpdateTranslationConfigInput 更新配置入参（指针表示部分更新）
type UpdateTranslationConfigInput struct {
	BaseURL     *string `json:"base_url"`
	Model       *string `json:"model"`
	APIKey      *string `json:"api_key"`
	ClearAPIKey bool    `json:"clear_api_key"`
	TimeoutMS   *int    `json:"timeout_ms"`
}

// TranslateInput 翻译请求
type TranslateInput struct {
	Texts       []string `json:"texts"`        // 待翻译文本（按相同顺序返回）
	TargetLangs []string `json:"target_langs"` // 目标语言代码列表，例如 ["en","ru"]
	SourceLang  string   `json:"source_lang"`  // 源语言代码（zh/en/auto），可选
}

// TranslateResult 翻译结果：texts × target_langs 的二维结构
// translations[i][lang] 表示第 i 条文本的 lang 译文
type TranslateResult struct {
	Translations []map[string]string `json:"translations"`
}

func defaultTranslationConfig() *TranslationConfig {
	return &TranslationConfig{
		BaseURL:   defaultTranslationBaseURL,
		Model:     defaultTranslationModel,
		TimeoutMS: defaultTranslationTimeoutMS,
	}
}

func (c *TranslationConfig) normalize() {
	if c == nil {
		return
	}
	c.BaseURL = strings.TrimSpace(c.BaseURL)
	if c.BaseURL == "" {
		c.BaseURL = defaultTranslationBaseURL
	}
	c.Model = strings.TrimSpace(c.Model)
	if c.Model == "" {
		c.Model = defaultTranslationModel
	}
	if c.TimeoutMS <= 0 {
		c.TimeoutMS = defaultTranslationTimeoutMS
	}
	if c.TimeoutMS > maxTranslationTimeoutMS {
		c.TimeoutMS = maxTranslationTimeoutMS
	}
}

// TranslationService 一键翻译服务
type TranslationService struct {
	settingRepo SettingRepository
	httpClient  *http.Client
}

// NewTranslationService 构造翻译服务
func NewTranslationService(settingRepo SettingRepository) *TranslationService {
	return &TranslationService{
		settingRepo: settingRepo,
		httpClient:  &http.Client{},
	}
}

// GetConfig 读取并返回脱敏后的配置
func (s *TranslationService) GetConfig(ctx context.Context) (*TranslationConfigView, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.toView(), nil
}

// UpdateConfig 更新部分字段后持久化
func (s *TranslationService) UpdateConfig(ctx context.Context, input UpdateTranslationConfigInput) (*TranslationConfigView, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if input.BaseURL != nil {
		cfg.BaseURL = strings.TrimSpace(*input.BaseURL)
	}
	if input.Model != nil {
		cfg.Model = strings.TrimSpace(*input.Model)
	}
	if input.TimeoutMS != nil {
		cfg.TimeoutMS = *input.TimeoutMS
	}
	if input.ClearAPIKey {
		cfg.APIKey = ""
	} else if input.APIKey != nil {
		key := strings.TrimSpace(*input.APIKey)
		if key != "" {
			cfg.APIKey = key
		}
	}
	cfg.normalize()
	if err := s.validateConfig(cfg); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal translation config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyTranslationConfig, string(raw)); err != nil {
		return nil, fmt.Errorf("save translation config: %w", err)
	}
	return cfg.toView(), nil
}

// Translate 调用 LLM 把每条文本翻译到所有目标语言。
// 调用方传入 m 条文本和 n 个目标语言，得到 m × n 个译文。
func (s *TranslationService) Translate(ctx context.Context, input TranslateInput) (*TranslateResult, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, infraerrors.BadRequest("TRANSLATION_API_KEY_MISSING", "翻译服务未配置 API Key")
	}
	texts, err := normalizeTranslationTexts(input.Texts)
	if err != nil {
		return nil, err
	}
	langs, err := normalizeTranslationLangs(input.TargetLangs)
	if err != nil {
		return nil, err
	}
	source := strings.TrimSpace(input.SourceLang)
	if source == "" {
		source = "auto"
	}

	// 构造 prompt：让模型以 JSON 数组返回翻译结果，每条对应输入顺序，含全部目标语言
	prompt := buildTranslationPrompt(texts, langs, source)
	raw, err := s.callChatCompletion(ctx, cfg, prompt)
	if err != nil {
		return nil, err
	}
	parsed, err := parseTranslationResponse(raw, len(texts), langs)
	if err != nil {
		return nil, infraerrors.InternalServer("TRANSLATION_PARSE_FAILED",
			fmt.Sprintf("无法解析翻译结果: %v", err))
	}
	return &TranslateResult{Translations: parsed}, nil
}

func (s *TranslationService) loadConfig(ctx context.Context) (*TranslationConfig, error) {
	cfg := defaultTranslationConfig()
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyTranslationConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			cfg.normalize()
			return cfg, nil
		}
		return nil, fmt.Errorf("get translation config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		cfg.normalize()
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, infraerrors.BadRequest("INVALID_TRANSLATION_CONFIG", "翻译配置不是有效 JSON")
	}
	cfg.normalize()
	return cfg, nil
}

func (s *TranslationService) validateConfig(cfg *TranslationConfig) error {
	if cfg == nil {
		return infraerrors.BadRequest("INVALID_TRANSLATION_CONFIG", "翻译配置不能为空")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return infraerrors.BadRequest("INVALID_TRANSLATION_BASE_URL", "Base URL 无效")
	}
	if cfg.Model == "" {
		return infraerrors.BadRequest("INVALID_TRANSLATION_MODEL", "Model 不能为空")
	}
	return nil
}

func (c *TranslationConfig) toView() *TranslationConfigView {
	v := &TranslationConfigView{
		BaseURL:   c.BaseURL,
		Model:     c.Model,
		TimeoutMS: c.TimeoutMS,
	}
	if k := strings.TrimSpace(c.APIKey); k != "" {
		v.APIKeyConfigured = true
		v.APIKeyMasked = maskAPIKey(k)
	}
	return v
}

func maskAPIKey(k string) string {
	if len(k) <= 8 {
		return strings.Repeat("*", len(k))
	}
	return k[:4] + strings.Repeat("*", 6) + k[len(k)-4:]
}

func normalizeTranslationTexts(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, infraerrors.BadRequest("TRANSLATION_TEXTS_EMPTY", "待翻译文本为空")
	}
	if len(in) > maxTranslationBatch {
		return nil, infraerrors.BadRequest("TRANSLATION_BATCH_TOO_LARGE",
			fmt.Sprintf("一次最多翻译 %d 条文本", maxTranslationBatch))
	}
	out := make([]string, 0, len(in))
	for _, t := range in {
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, infraerrors.BadRequest("TRANSLATION_TEXT_EMPTY", "待翻译文本含空白条目")
		}
		if len(s) > maxTranslationCharsPerItem {
			return nil, infraerrors.BadRequest("TRANSLATION_TEXT_TOO_LONG",
				fmt.Sprintf("单条文本不能超过 %d 字符", maxTranslationCharsPerItem))
		}
		out = append(out, s)
	}
	return out, nil
}

func normalizeTranslationLangs(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, infraerrors.BadRequest("TRANSLATION_LANGS_EMPTY", "目标语言为空")
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, lang := range in {
		l := strings.TrimSpace(lang)
		if l == "" {
			continue
		}
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}
	if len(out) == 0 {
		return nil, infraerrors.BadRequest("TRANSLATION_LANGS_EMPTY", "目标语言为空")
	}
	return out, nil
}

func buildTranslationPrompt(texts []string, langs []string, source string) string {
	// 系统级要求：仅返回 JSON 数组，不要任何额外说明
	type item struct {
		Index int    `json:"index"`
		Text  string `json:"text"`
	}
	items := make([]item, len(texts))
	for i, t := range texts {
		items[i] = item{Index: i, Text: t}
	}
	body, _ := json.Marshal(items)

	var b strings.Builder
	_, _ = b.WriteString("You are a professional translator. Translate each input text into the target languages.\n")
	if source != "" && source != "auto" {
		_, _ = b.WriteString("The source language is: ")
		_, _ = b.WriteString(source)
		_, _ = b.WriteString(".\n")
	} else {
		_, _ = b.WriteString("Detect the source language automatically.\n")
	}
	_, _ = b.WriteString("Target languages (BCP-47 / ISO 639-1 codes): ")
	_, _ = b.WriteString(strings.Join(langs, ", "))
	_, _ = b.WriteString(".\n\n")
	_, _ = b.WriteString("Rules:\n")
	_, _ = b.WriteString("1. Preserve the original meaning, tone and any inline punctuation.\n")
	_, _ = b.WriteString("2. Do NOT add explanations, prefixes, suffixes, or quotes around the translation.\n")
	_, _ = b.WriteString("3. If the source already matches a target language, just return it as-is for that language.\n")
	_, _ = b.WriteString("4. Output ONLY a JSON array, no markdown fences, no commentary.\n\n")
	_, _ = b.WriteString("Input array (each item has an `index` and `text`):\n")
	_, _ = b.Write(body)
	_, _ = b.WriteString("\n\nOutput format: a JSON array of the same length, each element is an object with key `index` (matching input) and a key per target language code holding the translation.\nExample for target_langs=[\"en\",\"ru\"]:\n[{\"index\":0,\"en\":\"Hello\",\"ru\":\"Привет\"}]\n")
	return b.String()
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model          string                  `json:"model"`
	Messages       []chatCompletionMessage `json:"messages"`
	Temperature    float64                 `json:"temperature"`
	ResponseFormat *responseFormat         `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatCompletionMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (s *TranslationService) callChatCompletion(ctx context.Context, cfg *TranslationConfig, userPrompt string) (string, error) {
	base := strings.TrimRight(cfg.BaseURL, "/")
	endpoint, err := url.JoinPath(base, "/v1/chat/completions")
	if err != nil {
		return "", err
	}
	payload := chatCompletionRequest{
		Model: cfg.Model,
		Messages: []chatCompletionMessage{
			{Role: "system", Content: "You are a translation engine. Always reply with a single JSON array only — no markdown, no commentary."},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0,
		// 多数 OpenAI 兼容网关支持 json_object；不支持时会被忽略或返回 400，
		// 因此即便在 system prompt 中已经强约束过了，这里也只是作为 best-effort 提示。
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", infraerrors.InternalServer("TRANSLATION_UPSTREAM_ERROR",
			fmt.Sprintf("调用翻译服务失败: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", infraerrors.InternalServer("TRANSLATION_UPSTREAM_ERROR",
			fmt.Sprintf("翻译服务返回 %d: %s", resp.StatusCode, truncateTranslationError(string(body), 256)))
	}

	var out chatCompletionResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", infraerrors.InternalServer("TRANSLATION_UPSTREAM_BAD_RESPONSE",
			fmt.Sprintf("无法解析翻译服务响应: %v", err))
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", infraerrors.InternalServer("TRANSLATION_UPSTREAM_ERROR", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", infraerrors.InternalServer("TRANSLATION_UPSTREAM_BAD_RESPONSE", "翻译服务未返回 choices")
	}
	return out.Choices[0].Message.Content, nil
}

// parseTranslationResponse 解析 LLM 返回，得到与输入文本顺序对齐的 []map[lang]text
func parseTranslationResponse(raw string, expectedLen int, langs []string) ([]map[string]string, error) {
	content := strings.TrimSpace(raw)
	// 去掉 ```json ... ``` 代码块包裹
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}
	// 截取首个 [ 与最后一个 ] 之间的片段（容忍模型在前后多说几句）
	if l, r := strings.Index(content, "["), strings.LastIndex(content, "]"); l >= 0 && r > l {
		content = content[l : r+1]
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(content), &rows); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if len(rows) != expectedLen {
		return nil, fmt.Errorf("expected %d items, got %d", expectedLen, len(rows))
	}

	results := make([]map[string]string, expectedLen)
	for i, row := range rows {
		out := make(map[string]string, len(langs))
		for _, lang := range langs {
			v, ok := row[lang]
			if !ok {
				return nil, fmt.Errorf("row %d missing language %q", i, lang)
			}
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("row %d language %q is not a string", i, lang)
			}
			out[lang] = strings.TrimSpace(s)
		}
		results[i] = out
	}
	return results, nil
}

func truncateTranslationError(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
