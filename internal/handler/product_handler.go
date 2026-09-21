package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/response"
)

type ProductHandler struct{ svc *service.ProductService }
func NewProduct(s *service.ProductService) *ProductHandler { return &ProductHandler{svc:s} }

// List godoc
// @Summary List products
// @Tags products
// @Produce json
// @Param q query string false "search"
// @Param limit query int false "limit"
// @Success 200 {object} map[string]any
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	q:=c.Query("q")
	limit,_:=strconv.Atoi(c.DefaultQuery("limit","20"))
	offset,_:=strconv.Atoi(c.DefaultQuery("offset","0"))
	list,_:=h.svc.List(c.Request.Context(), q, limit, offset)
	response.OK(c,list)
}
// Get godoc
// @Summary Get product
// @Tags products
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Router /products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	p,err:=h.svc.Get(c.Request.Context(), c.Param("id"))
	if err!=nil{response.Err(c,404,err.Error());return}
	response.OK(c,p)
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
	if err:=c.ShouldBindJSON(&req);err!=nil{response.Err(c,400,err.Error());return}
	uid,_:=c.Get("userID")
	p:=&model.Product{SellerID: uid.(string), Title: req.Title, Description: req.Description, Price: req.Price, Condition: req.Condition, CategoryID: req.CategoryID, BrandID: req.BrandID, Status: "ACTIVE"}
	if err:=h.svc.Create(c.Request.Context(), p);err!=nil{response.Err(c,400,err.Error());return}
	response.Created(c,p)
}
// Reserve godoc
// @Summary Reserve product
// @Tags products
// @Security Bearer
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Router /products/{id}/reserve [post]
func (h *ProductHandler) Reserve(c *gin.Context) {
	if err:=h.svc.Reserve(c.Request.Context(), c.Param("id"));err!=nil{response.Err(c,400,err.Error());return}
	response.OK(c, gin.H{"reserved":true})
}
func (h *ProductHandler) Update(c *gin.Context) { response.OK(c, gin.H{"todo":true}) }
func (h *ProductHandler) Delete(c *gin.Context) { response.OK(c, gin.H{"deleted":true}) }
