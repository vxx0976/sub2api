package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// TranslationHandler 暴露一键翻译相关的管理端接口
type TranslationHandler struct {
	service *service.TranslationService
}

// NewTranslationHandler 构造
func NewTranslationHandler(svc *service.TranslationService) *TranslationHandler {
	return &TranslationHandler{service: svc}
}

type translationConfigRequest struct {
	BaseURL     *string `json:"base_url"`
	Model       *string `json:"model"`
	APIKey      *string `json:"api_key"`
	ClearAPIKey bool    `json:"clear_api_key"`
	TimeoutMS   *int    `json:"timeout_ms"`
}

type translateRequest struct {
	Texts       []string `json:"texts"`
	TargetLangs []string `json:"target_langs"`
	SourceLang  string   `json:"source_lang"`
}

// GetConfig GET /admin/translation/config
func (h *TranslationHandler) GetConfig(c *gin.Context) {
	cfg, err := h.service.GetConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateConfig PUT /admin/translation/config
func (h *TranslationHandler) UpdateConfig(c *gin.Context) {
	var req translationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	cfg, err := h.service.UpdateConfig(c.Request.Context(), service.UpdateTranslationConfigInput{
		BaseURL:     req.BaseURL,
		Model:       req.Model,
		APIKey:      req.APIKey,
		ClearAPIKey: req.ClearAPIKey,
		TimeoutMS:   req.TimeoutMS,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// Translate POST /admin/translation/translate
func (h *TranslationHandler) Translate(c *gin.Context) {
	var req translateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Translate(c.Request.Context(), service.TranslateInput{
		Texts:       req.Texts,
		TargetLangs: req.TargetLangs,
		SourceLang:  req.SourceLang,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
