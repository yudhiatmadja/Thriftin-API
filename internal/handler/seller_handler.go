package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/response"
)

type SellerHandler struct {
	seller *service.SellerService
	banned *service.BannedService
}

func NewSeller(s *service.SellerService, b *service.BannedService) *SellerHandler {
	return &SellerHandler{seller: s, banned: b}
}

// Apply godoc
// @Summary Daftar jadi seller (isi data diri + jenis barang)
// @Tags seller
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body SellerApplyReq true "apply"
// @Success 201 {object} map[string]any
// @Router /seller/apply [post]
func (h *SellerHandler) Apply(c *gin.Context) {
	var req SellerApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	err := h.seller.Apply(c.Request.Context(), model.SellerApplication{
		UserID: uid.(string), StoreName: req.StoreName, Phone: req.Phone,
		Address: req.Address, IDNumber: req.IDNumber,
		ProductTypes: req.ProductTypes, Description: &req.Description,
	})
	if err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, gin.H{"status": "PENDING", "message": "Pengajuan terkirim, tunggu verifikasi admin"})
}

// MyStatus godoc
// @Summary Status pengajuan seller saya
// @Tags seller
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /seller/status [get]
func (h *SellerHandler) MyStatus(c *gin.Context) {
	uid, _ := c.Get("userID")
	a, err := h.seller.MyStatus(c.Request.Context(), uid.(string))
	if err != nil {
		response.OK(c, gin.H{"status": "NONE"})
		return
	}
	verified := h.seller.IsVerified(c.Request.Context(), uid.(string))
	response.OK(c, gin.H{"application": a, "seller_verified": verified})
}

// ListPending godoc
// @Summary Admin: list pengajuan seller pending
// @Tags admin
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /admin/sellers/pending [get]
func (h *SellerHandler) ListPending(c *gin.Context) {
	list, err := h.seller.ListPending(c.Request.Context())
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// Decide godoc
// @Summary Admin: approve/reject seller
// @Tags admin
// @Security Bearer
// @Accept json
// @Param id path string true "application id"
// @Param body body SellerDecideReq true "decision"
// @Success 200 {object} map[string]any
// @Router /admin/sellers/{id}/decide [post]
func (h *SellerHandler) Decide(c *gin.Context) {
	var req SellerDecideReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	if err := h.seller.Decide(c.Request.Context(), c.Param("id"), req.Status, req.AdminNote); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"decided": req.Status})
}

// BannedList godoc
// @Summary List kata terlarang (filter nama produk)
// @Tags seller
// @Success 200 {object} map[string]any
// @Router /banned-words [get]
func (h *SellerHandler) BannedList(c *gin.Context) {
	list, err := h.banned.List(c.Request.Context())
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// BannedAdd godoc
// @Summary Admin: tambah kata terlarang
// @Tags admin
// @Security Bearer
// @Success 201 {object} map[string]any
// @Router /admin/banned-words [post]
func (h *SellerHandler) BannedAdd(c *gin.Context) {
	var req BannedWordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	if err := h.banned.Add(c.Request.Context(), req.Word); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, gin.H{"added": req.Word})
}

// BannedRemove godoc
// @Summary Admin: hapus kata terlarang
// @Tags admin
// @Security Bearer
// @Param word path string true "word"
// @Success 200 {object} map[string]any
// @Router /admin/banned-words/{word} [delete]
func (h *SellerHandler) BannedRemove(c *gin.Context) {
	if err := h.banned.Remove(c.Request.Context(), c.Param("word")); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"removed": true})
}
