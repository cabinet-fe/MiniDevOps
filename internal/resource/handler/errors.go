package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bedrock/internal/pkg"
	"bedrock/internal/resource/service"
)

func writeServiceError(c *gin.Context, err error) {
	switch {
	case service.IsConflict(err):
		pkg.Error(c, http.StatusConflict, err.Error())
	case service.IsForbidden(err):
		pkg.Error(c, http.StatusForbidden, err.Error())
	case service.IsNotFound(err):
		pkg.Error(c, http.StatusNotFound, err.Error())
	default:
		pkg.Error(c, http.StatusBadRequest, err.Error())
	}
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err
}
