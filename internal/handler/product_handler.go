package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/response"
)

type ProductHandler struct{ svc *service.ProductService }

func NewProduct(s *service.ProductService) *ProductHandler { return &ProductHandler{svc: s} }

// List godoc
// @Summary List products
// @Tags products
// @Produce json
// @Param q query string false "search"
// @Param category_id query string false "category"
// @Param brand_id query string false "brand"
// @Param condition query string false "condition"
// @Param min_price query int false "min price"
// @Param max_price query int false "max price"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} map[string]any
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	minP, _ := strconv.ParseInt(c.Query("min_price"), 10, 64)
	maxP, _ := strconv.ParseInt(c.Query("max_price"), 10, 64)
	f := model.ProductFilter{
		Q: c.Query("q"), CategoryID: c.Query("category_id"), BrandID: c.Query("brand_id"),
		Condition: c.Query("condition"), MinPrice: minP, MaxPrice: maxP, Limit: limit, Offset: offset,
	}
	list, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// Get godoc
// @Summary Get product
// @Tags products
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Router /products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Err(c, 404, err.Error())
		return
	}
	response.OK(c, p)
}

// Create godoc
// @Summary Create product
// @Tags products
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body ProductReq true "product"
// @Success 201 {object} map[string]any
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req ProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}
	p := &model.Product{SellerID: uid.(string), Title: req.Title, Description: req.Description, Price: req.Price,
		Condition: req.Condition, CategoryID: req.CategoryID, BrandID: req.BrandID,
		Size: req.Size, Color: req.Color, Material: req.Material, Location: req.Location, Status: "ACTIVE"}
	if err := h.svc.Create(c.Request.Context(), p, req.Images, r); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	p, _ = h.svc.Get(c.Request.Context(), p.ID)
	response.Created(c, p)
}

// Update godoc
// @Summary Update my product
// @Tags products
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body ProductUpdateReq true "update"
// @Success 200 {object} map[string]any
// @Router /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	var req ProductUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.Update(c.Request.Context(), c.Param("id"), uid.(string), req.Title, req.Description, req.Price, req.Status); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	p, _ := h.svc.Get(c.Request.Context(), c.Param("id"))
	response.OK(c, p)
}

// Delete godoc
// @Summary Delete (archive) my product
// @Tags products
// @Security Bearer
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}
	if err := h.svc.Delete(c.Request.Context(), c.Param("id"), uid.(string), r); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

// MyProducts godoc
// @Summary My products (seller)
// @Tags products
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /products/mine [get]
func (h *ProductHandler) Mine(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.MyProducts(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// AddImage godoc
// @Summary Add image URL to my product
// @Tags products
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body ProductImageReq true "image"
// @Success 200 {object} map[string]any
// @Router /products/{id}/images [post]
func (h *ProductHandler) AddImage(c *gin.Context) {
	var req ProductImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.AddImage(c.Request.Context(), c.Param("id"), uid.(string), req.URL); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"added": true})
}
 
// Reserve godoc
// @Summary Reserve product
// @Tags products
// @Security Bearer
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Router /products/{id}/reserve [post]
func (h *ProductHandler) Reserve(c *gin.Context) {
	if err := h.svc.Reserve(c.Request.Context(), c.Param("id")); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"reserved": true})
}
