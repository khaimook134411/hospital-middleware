package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/util"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.WriteError(c, util.ErrValidation(err.Error()))
		return
	}
	token, appErr := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if appErr != nil {
		util.WriteError(c, appErr)
		return
	}
	util.WriteSuccess(c, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.cfg.JWTTTL.Seconds()),
	})
}
