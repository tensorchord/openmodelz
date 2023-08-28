package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Create the image cache.
// @Description Create the image cache.
// @Tags        image-cache
// @Accept      json
// @Produce     json
// @Param       body body     types.ImageCache true "image-cache"
// @Success     201  {object} types.ImageCache
// @Router      /system/image-cache [post]
func (s *Server) handleImageCacheCreate(c *gin.Context) error {
	var req types.ImageCache
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(
			http.StatusBadRequest, err, "image-cache-create")
	}

	if err := s.validator.ValidateImageCacheRequest(&req); err != nil {
		return NewError(
			http.StatusBadRequest, err, "image-cache-create")
	}
	inference, err := s.runtime.InferenceGetCRD(req.Namespace, req.Name)
	if err != nil {
		return errFromErrDefs(err, "inference-instance-list")
	}

	if err := s.runtime.ImageCacheCreate(c.Request.Context(), req, inference); err != nil {
		return errFromErrDefs(err, "image-cache-create")
	}
	c.JSON(http.StatusOK, req)
	return nil
}
