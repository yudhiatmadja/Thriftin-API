package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thriftin/api/pkg/response"
)

var allowImg = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
var allowVid = map[string]bool{".mp4": true, ".webm": true, ".mov": true}

// Upload godoc
// @Summary Upload images/video (multipart files[], max 6, 15MB each)
// @Tags uploads
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} map[string]any
// @Router /uploads [post]
func Upload(c *gin.Context) {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		response.Err(c, 400, "field 'files' required (multipart)")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		response.Err(c, 400, "no files uploaded")
		return
	}
	if len(files) > 6 {
		response.Err(c, 400, "max 6 files per upload")
		return
	}
	urls := []string{}
	for _, f := range files {
		if f.Size > 15<<20 {
			response.Err(c, 400, fmt.Sprintf("%s exceeds 15MB", f.Filename))
			return
		}
		ext := strings.ToLower(filepath.Ext(f.Filename))
		if !allowImg[ext] && !allowVid[ext] {
			response.Err(c, 400, fmt.Sprintf("%s: only jpg/png/webp/gif/mp4/webm/mov allowed", f.Filename))
			return
		}
		name := uuid.NewString() + ext
		dst := filepath.Join("uploads", name)
		if err := c.SaveUploadedFile(f, dst); err != nil {
			response.Err(c, 500, err.Error())
			return
		}
		urls = append(urls, "/uploads/"+name)
	}
	response.OK(c, gin.H{"urls": urls})
}
