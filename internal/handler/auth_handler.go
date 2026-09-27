package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/response"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuth(s *service.AuthService) *AuthHandler { return &AuthHandler{svc: s} }

// Register godoc
// @Summary Register
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterReq true "register"
// @Success 201 {object} map[string]any
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.Username, req.FullName)
	if err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, u)
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginReq true "login"
// @Success 200 {object} map[string]any
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	a, r, u, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Err(c, 401, "invalid credentials")
		return
	}
	response.OK(c, gin.H{"access_token": a, "refresh_token": r, "user": u})
}

// Google godoc
// @Summary Login with Google (kirim Google ID token, auto-register kalau belum ada)
// @Tags auth
// @Accept json
// @Produce json
// @Param body body GoogleReq true "google"
// @Success 200 {object} map[string]any
// @Router /auth/google [post]
func (h *AuthHandler) Google(c *gin.Context) {
	var req GoogleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	a, r, u, err := h.svc.GoogleLogin(c.Request.Context(), req.IDToken)
	if err != nil {
		response.Err(c, 401, err.Error())
		return
	}
	response.OK(c, gin.H{"access_token": a, "refresh_token": r, "user": u})
}

// Refresh godoc
// @Summary Refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RefreshReq true "refresh"
// @Success 200 {object} map[string]any
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	a, r, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Err(c, 401, err.Error())
		return
	}
	response.OK(c, gin.H{"access_token": a, "refresh_token": r})
}

// Logout godoc
// @Summary Logout (revoke refresh token)
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LogoutReq true "logout"
// @Success 200 {object} map[string]any
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"logged_out": true})
}

// Me godoc
// @Summary Me
// @Tags auth
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	uid, _ := c.Get("userID")
	u, err := h.svc.Me(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 404, err.Error())
		return
	}
	response.OK(c, u)
}
