package staff

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/util"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9_.]+$`)

type createRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Hospital string `json:"hospital" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.WriteError(c, util.ErrValidation(err.Error()))
		return
	}
	if !usernameRe.MatchString(req.Username) {
		util.WriteError(c, util.ErrValidation("username must contain only a-z, 0-9, _ or ."))
		return
	}

	resp, appErr := h.svc.Create(c.Request.Context(), req.Username, req.Password, req.Hospital)
	if appErr != nil {
		util.WriteError(c, appErr)
		return
	}
	util.WriteSuccess(c, http.StatusCreated, resp)
}
