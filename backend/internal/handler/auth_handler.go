package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	cfg                  *config.Config
	authService          *service.AuthService
	userService          *service.UserService
	settingSvc           *service.SettingService
	promoService         *service.PromoService
	referralService      *service.ReferralService
	redeemService        *service.RedeemService
	totpService          *service.TotpService
	resellerSettingRepo  service.ResellerSettingRepository
}

// SetResellerSettingRepo injects reseller setting repository for price_multiplier lookup in auth/me.
func (h *AuthHandler) SetResellerSettingRepo(repo service.ResellerSettingRepository) {
	h.resellerSettingRepo = repo
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(cfg *config.Config, authService *service.AuthService, userService *service.UserService, settingService *service.SettingService, promoService *service.PromoService, referralService *service.ReferralService, redeemService *service.RedeemService, totpService *service.TotpService) *AuthHandler {
	return &AuthHandler{
		cfg:             cfg,
		authService:     authService,
		userService:     userService,
		settingSvc:      settingService,
		promoService:    promoService,
		referralService: referralService,
		redeemService:   redeemService,
		totpService:     totpService,
	}
}

// buildEmailOptions constructs EmailOptions from the request context.
// Extracts reseller domain branding and user locale.
func (h *AuthHandler) buildEmailOptions(c *gin.Context) service.EmailOptions {
	var opts service.EmailOptions

	// User locale from Accept-Language header
	opts.Locale = c.GetHeader("Accept-Language")

	// Reseller domain overrides
	if resellerCtx := middleware2.GetResellerDomainFromContext(c); resellerCtx != nil && resellerCtx.Domain != "" {
		opts.FromName = resellerCtx.Domain
		// Site name: prefer reseller's configured name, fallback to domain
		if resellerCtx.SiteName != "" {
			opts.SiteName = resellerCtx.SiteName
		} else {
			opts.SiteName = resellerCtx.Domain
		}
	}

	return opts
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	VerifyCode     string `json:"verify_code"`
	TurnstileToken string `json:"turnstile_token"`
	PromoCode      string `json:"promo_code"`      // 注册优惠码
	InvitationCode string `json:"invitation_code"` // 邀请码
	ParentID       *int64 `json:"parent_id"`       // 上级分销商用户 ID
}

// SendVerifyCodeRequest 发送验证码请求
type SendVerifyCodeRequest struct {
	Email          string `json:"email" binding:"required,email"`
	TurnstileToken string `json:"turnstile_token"`
}

// SendVerifyCodeResponse 发送验证码响应
type SendVerifyCodeResponse struct {
	Message   string `json:"message"`
	Countdown int    `json:"countdown"` // 倒计时秒数
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	TurnstileToken string `json:"turnstile_token"`
}

// AuthResponse 认证响应格式（匹配前端期望）
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"` // 新增：Refresh Token
	ExpiresIn    int       `json:"expires_in,omitempty"`    // 新增：Access Token有效期（秒）
	TokenType    string    `json:"token_type"`
	User         *dto.User `json:"user"`
}

// respondWithTokenPair 生成 Token 对并返回认证响应
// 如果 Token 对生成失败，回退到只返回 Access Token（向后兼容）
func (h *AuthHandler) respondWithTokenPair(c *gin.Context, user *service.User) {
	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user, "")
	if err != nil {
		slog.Error("failed to generate token pair", "error", err, "user_id", user.ID)
		// 回退到只返回Access Token
		token, tokenErr := h.authService.GenerateToken(user)
		if tokenErr != nil {
			response.InternalError(c, "Failed to generate token")
			return
		}
		response.Success(c, AuthResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			User:        dto.UserFromService(user),
		})
		return
	}
	response.Success(c, AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         dto.UserFromService(user),
	})
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Turnstile 验证（邮箱验证码注册场景避免重复校验一次性 token）
	if err := h.authService.VerifyTurnstileForRegister(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c), req.VerifyCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Determine parentID: explicit request field takes priority, then reseller domain context
	parentID := req.ParentID
	if parentID == nil {
		if resellerCtx := middleware2.GetResellerDomainFromContext(c); resellerCtx != nil {
			parentID = &resellerCtx.ResellerID
		}
	}

	// 获取注册来源域名
	registerDomain := c.Request.Host
	// 去掉端口号
	if idx := strings.LastIndex(registerDomain, ":"); idx != -1 {
		registerDomain = registerDomain[:idx]
	}

	_, user, err := h.authService.RegisterWithVerification(c.Request.Context(), req.Email, req.Password, req.VerifyCode, req.PromoCode, req.InvitationCode, parentID, registerDomain)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.respondWithTokenPair(c, user)
}

// SendVerifyCode 发送邮箱验证码
// POST /api/v1/auth/send-verify-code
func (h *AuthHandler) SendVerifyCode(c *gin.Context) {
	var req SendVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Turnstile 验证
	if err := h.authService.VerifyTurnstile(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 构建邮件选项：商户域名覆盖发件人和站点名称，用户语言
	emailOpts := h.buildEmailOptions(c)

	result, err := h.authService.SendVerifyCodeAsync(c.Request.Context(), req.Email, emailOpts)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, SendVerifyCodeResponse{
		Message:   "Verification code sent successfully",
		Countdown: result.Countdown,
	})
}

// checkLoginDomain verifies that the user is logging in from the correct domain.
// Admin and reseller roles are exempt and can log in from anywhere.
// Regular users with no parent must log in from the main site.
// Regular users with a parent (merchant sub-users) must log in from their merchant's domain.
func (h *AuthHandler) checkLoginDomain(c *gin.Context, user *service.User) bool {
	if user.Role == service.RoleAdmin || user.Role == service.RoleReseller {
		return true
	}

	resellerCtx := middleware2.GetResellerDomainFromContext(c)

	if user.ParentID == nil {
		// Main-site user: must log in from main site (no reseller context)
		if resellerCtx != nil {
			response.Forbidden(c, "此账号只能在码驿站登录，请前往码驿站登录页")
			return false
		}
	} else {
		// Merchant sub-user: must log in from their own merchant's domain
		if resellerCtx == nil {
			response.Forbidden(c, "此账号只能在所属商户站点登录，请联系商户获取登录地址")
			return false
		}
		if resellerCtx.ResellerID != *user.ParentID {
			response.Forbidden(c, "此账号只能在所属商户站点登录，请使用您注册时使用的站点")
			return false
		}
	}

	return true
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Turnstile 验证
	if err := h.authService.VerifyTurnstile(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	_ = token // token 由 authService.Login 返回但此处由 respondWithTokenPair 重新生成

	// Enforce login domain restriction
	if !h.checkLoginDomain(c, user) {
		return
	}

	// Check if TOTP 2FA is enabled for this user
	if h.totpService != nil && h.settingSvc.IsTotpEnabled(c.Request.Context()) && user.TotpEnabled {
		// Create a temporary login session for 2FA
		tempToken, err := h.totpService.CreateLoginSession(c.Request.Context(), user.ID, user.Email)
		if err != nil {
			response.InternalError(c, "Failed to create 2FA session")
			return
		}

		response.Success(c, TotpLoginResponse{
			Requires2FA:     true,
			TempToken:       tempToken,
			UserEmailMasked: service.MaskEmail(user.Email),
		})
		return
	}

	h.respondWithTokenPair(c, user)
}

// TotpLoginResponse represents the response when 2FA is required
type TotpLoginResponse struct {
	Requires2FA     bool   `json:"requires_2fa"`
	TempToken       string `json:"temp_token,omitempty"`
	UserEmailMasked string `json:"user_email_masked,omitempty"`
}

// Login2FARequest represents the 2FA login request
type Login2FARequest struct {
	TempToken string `json:"temp_token" binding:"required"`
	TotpCode  string `json:"totp_code" binding:"required,len=6"`
}

// Login2FA completes the login with 2FA verification
// POST /api/v1/auth/login/2fa
func (h *AuthHandler) Login2FA(c *gin.Context) {
	var req Login2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	slog.Debug("login_2fa_request",
		"temp_token_len", len(req.TempToken),
		"totp_code_len", len(req.TotpCode))

	// Get the login session
	session, err := h.totpService.GetLoginSession(c.Request.Context(), req.TempToken)
	if err != nil || session == nil {
		tokenPrefix := ""
		if len(req.TempToken) >= 8 {
			tokenPrefix = req.TempToken[:8]
		}
		slog.Debug("login_2fa_session_invalid",
			"temp_token_prefix", tokenPrefix,
			"error", err)
		response.BadRequest(c, "Invalid or expired 2FA session")
		return
	}

	slog.Debug("login_2fa_session_found",
		"user_id", session.UserID,
		"email", session.Email)

	// Verify the TOTP code
	if err := h.totpService.VerifyCode(c.Request.Context(), session.UserID, req.TotpCode); err != nil {
		slog.Debug("login_2fa_verify_failed",
			"user_id", session.UserID,
			"error", err)
		response.ErrorFrom(c, err)
		return
	}

	// Delete the login session
	_ = h.totpService.DeleteLoginSession(c.Request.Context(), req.TempToken)

	// Get the user
	user, err := h.userService.GetByID(c.Request.Context(), session.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Enforce login domain restriction
	if !h.checkLoginDomain(c, user) {
		return
	}

	h.respondWithTokenPair(c, user)
}

// GetCurrentUser handles getting current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	type UserResponse struct {
		*dto.User
		RunMode                 string   `json:"run_mode"`
		ResellerPriceMultiplier *float64 `json:"reseller_price_multiplier,omitempty"`
		ResellerAgentEnabled    bool     `json:"reseller_agent_enabled,omitempty"`
		ResellerPayURL          string   `json:"reseller_pay_url,omitempty"`
		ResellerSellingPrice    *float64 `json:"reseller_selling_price,omitempty"`
	}

	runMode := config.RunModeStandard
	if h.cfg != nil {
		runMode = h.cfg.RunMode
	}

	resp := UserResponse{User: dto.UserFromService(user), RunMode: runMode}

	// If the user belongs to a reseller with merchant_mode enabled, expose the
	// reseller's price_multiplier so the payment service can apply the correct
	// balance ratio. Requires merchant_mode == "enabled" as the guard.
	if user.ParentID != nil && h.resellerSettingRepo != nil {
		if rs, err := h.resellerSettingRepo.GetAll(c.Request.Context(), *user.ParentID); err == nil && rs["merchant_mode"] == "enabled" {
			resp.ResellerAgentEnabled = true
			resp.ResellerPayURL = rs["pay_url"]
			if mult, err := strconv.ParseFloat(rs["price_multiplier"], 64); err == nil && mult > 0 {
				resp.ResellerPriceMultiplier = &mult
			}
			if sp, err := strconv.ParseFloat(rs["selling_price"], 64); err == nil && sp > 0 {
				resp.ResellerSellingPrice = &sp
			}
		}
	}

	response.Success(c, resp)
}

// ValidatePromoCodeRequest 验证优惠码请求
type ValidatePromoCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// ValidatePromoCodeResponse 验证优惠码/邀请码响应
type ValidatePromoCodeResponse struct {
	Valid        bool    `json:"valid"`
	BonusAmount  float64 `json:"bonus_amount,omitempty"`
	ErrorCode    string  `json:"error_code,omitempty"`
	Message      string  `json:"message,omitempty"`
	IsReferral   bool    `json:"is_referral,omitempty"`   // 是否为邀请码
	ReferrerName string  `json:"referrer_name,omitempty"` // 邀请人用户名（脱敏）
}

// ValidatePromoCode 验证优惠码或邀请码（公开接口，注册前调用）
// POST /api/v1/auth/validate-promo-code
func (h *AuthHandler) ValidatePromoCode(c *gin.Context) {
	var req ValidatePromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 首先检查是否为邀请码（以 R 开头的 8 位字符）
	if h.referralService != nil && h.referralService.IsReferralCode(req.Code) {
		referrer, err := h.referralService.ValidateReferralCodePublic(c.Request.Context(), req.Code)
		if err != nil {
			errorCode := "REFERRAL_CODE_INVALID"
			switch err {
			case service.ErrReferralCodeNotFound:
				errorCode = "REFERRAL_CODE_NOT_FOUND"
			case service.ErrReferralCodeInvalid:
				errorCode = "REFERRAL_CODE_INVALID"
			}
			response.Success(c, ValidatePromoCodeResponse{
				Valid:      false,
				IsReferral: true,
				ErrorCode:  errorCode,
			})
			return
		}

		// 邀请码有效，返回邀请人信息
		referrerName := service.MaskEmail(referrer.Email)
		if referrer.Username != "" {
			referrerName = referrer.Username
		}

		response.Success(c, ValidatePromoCodeResponse{
			Valid:        true,
			IsReferral:   true,
			ReferrerName: referrerName,
		})
		return
	}

	// 检查优惠码功能是否启用
	if h.settingSvc != nil && !h.settingSvc.IsPromoCodeEnabled(c.Request.Context()) {
		response.Success(c, ValidatePromoCodeResponse{
			Valid:     false,
			ErrorCode: "PROMO_CODE_DISABLED",
		})
		return
	}

	promoCode, err := h.promoService.ValidatePromoCode(c.Request.Context(), req.Code)
	if err != nil {
		// 根据错误类型返回对应的错误码
		errorCode := "PROMO_CODE_INVALID"
		switch err {
		case service.ErrPromoCodeNotFound:
			errorCode = "PROMO_CODE_NOT_FOUND"
		case service.ErrPromoCodeExpired:
			errorCode = "PROMO_CODE_EXPIRED"
		case service.ErrPromoCodeDisabled:
			errorCode = "PROMO_CODE_DISABLED"
		case service.ErrPromoCodeMaxUsed:
			errorCode = "PROMO_CODE_MAX_USED"
		case service.ErrPromoCodeAlreadyUsed:
			errorCode = "PROMO_CODE_ALREADY_USED"
		}

		response.Success(c, ValidatePromoCodeResponse{
			Valid:     false,
			ErrorCode: errorCode,
		})
		return
	}

	if promoCode == nil {
		response.Success(c, ValidatePromoCodeResponse{
			Valid:     false,
			ErrorCode: "PROMO_CODE_INVALID",
		})
		return
	}

	response.Success(c, ValidatePromoCodeResponse{
		Valid:       true,
		BonusAmount: promoCode.BonusAmount,
	})
}

// ValidateInvitationCodeRequest 验证邀请码请求
type ValidateInvitationCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// ValidateInvitationCodeResponse 验证邀请码响应
type ValidateInvitationCodeResponse struct {
	Valid     bool   `json:"valid"`
	ErrorCode string `json:"error_code,omitempty"`
}

// ValidateInvitationCode 验证邀请码（公开接口，注册前调用）
// POST /api/v1/auth/validate-invitation-code
func (h *AuthHandler) ValidateInvitationCode(c *gin.Context) {
	// 检查邀请码功能是否启用
	if h.settingSvc == nil || !h.settingSvc.IsInvitationCodeEnabled(c.Request.Context()) {
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_DISABLED",
		})
		return
	}

	var req ValidateInvitationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 验证邀请码
	redeemCode, err := h.redeemService.GetByCode(c.Request.Context(), req.Code)
	if err != nil {
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_NOT_FOUND",
		})
		return
	}

	// 检查类型和状态
	if redeemCode.Type != service.RedeemTypeInvitation {
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_INVALID",
		})
		return
	}

	if redeemCode.Status != service.StatusUnused {
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_USED",
		})
		return
	}

	response.Success(c, ValidateInvitationCodeResponse{
		Valid: true,
	})
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email          string `json:"email" binding:"required,email"`
	TurnstileToken string `json:"turnstile_token"`
}

// ForgotPasswordResponse 忘记密码响应
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ForgotPassword 请求密码重置
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Turnstile 验证
	if err := h.authService.VerifyTurnstile(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	frontendBaseURL := strings.TrimSpace(h.cfg.Server.FrontendURL)
	if frontendBaseURL == "" {
		slog.Error("server.frontend_url not configured; cannot build password reset link")
		response.InternalError(c, "Password reset is not configured")
		return
	}

	// 构建邮件选项：商户域名覆盖发件人和站点名称，用户语言
	emailOpts := h.buildEmailOptions(c)

	// Request password reset (async)
	// Note: This returns success even if email doesn't exist (to prevent enumeration)
	if err := h.authService.RequestPasswordResetAsync(c.Request.Context(), req.Email, frontendBaseURL, emailOpts); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ForgotPasswordResponse{
		Message: "If your email is registered, you will receive a password reset link shortly.",
	})
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ResetPassword 重置密码
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(c.Request.Context(), req.Email, req.Token, req.NewPassword); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ResetPasswordResponse{
		Message: "Your password has been reset successfully. You can now log in with your new password.",
	})
}

// ==================== Token Refresh Endpoints ====================

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // Access Token有效期（秒）
	TokenType    string `json:"token_type"`
}

// RefreshToken 刷新Token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	tokenPair, err := h.authService.RefreshTokenPair(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	})
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"` // 可选：撤销指定的Refresh Token
}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Message string `json:"message"`
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	// 允许空请求体（向后兼容）
	_ = c.ShouldBindJSON(&req)

	// 如果提供了Refresh Token，撤销它
	if req.RefreshToken != "" {
		if err := h.authService.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
			slog.Debug("failed to revoke refresh token", "error", err)
			// 不影响登出流程
		}
	}

	response.Success(c, LogoutResponse{
		Message: "Logged out successfully",
	})
}

// RevokeAllSessionsResponse 撤销所有会话响应
type RevokeAllSessionsResponse struct {
	Message string `json:"message"`
}

// RevokeAllSessions 撤销当前用户的所有会话
// POST /api/v1/auth/revoke-all-sessions
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.authService.RevokeAllUserSessions(c.Request.Context(), subject.UserID); err != nil {
		slog.Error("failed to revoke all sessions", "user_id", subject.UserID, "error", err)
		response.InternalError(c, "Failed to revoke sessions")
		return
	}

	response.Success(c, RevokeAllSessionsResponse{
		Message: "All sessions have been revoked. Please log in again.",
	})
}
