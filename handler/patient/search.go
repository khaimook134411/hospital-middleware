package patient

import (
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	"github.com/khaimook/hospital-middleware/util"
)

var nationalIDRe = regexp.MustCompile(`^\d{13}$`)

type searchRequest struct {
	NationalID  string `form:"national_id"`
	PassportID  string `form:"passport_id" binding:"omitempty,max=20"`
	FirstName   string `form:"first_name" binding:"omitempty,max=100"`
	MiddleName  string `form:"middle_name" binding:"omitempty,max=100"`
	LastName    string `form:"last_name" binding:"omitempty,max=100"`
	DateOfBirth string `form:"date_of_birth"`
	PhoneNumber string `form:"phone_number" binding:"omitempty,max=20"`
	Email       string `form:"email" binding:"omitempty,email"`
	Limit       int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset      int    `form:"offset" binding:"omitempty,min=0"`
}

func (h *Handler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		util.WriteError(c, util.ErrValidation(err.Error()))
		return
	}

	req.NationalID = strings.TrimSpace(req.NationalID)
	if req.NationalID != "" && !nationalIDRe.MatchString(req.NationalID) {
		util.WriteError(c, util.ErrValidation("national_id must be 13 digits"))
		return
	}

	req.DateOfBirth = strings.TrimSpace(req.DateOfBirth)
	if req.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", req.DateOfBirth); err != nil {
			util.WriteError(c, util.ErrValidation("date_of_birth must be YYYY-MM-DD"))
			return
		}
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	claims := util.GetClaims(c)

	params := patientrepo.SearchParams{
		NationalID:  req.NationalID,
		PassportID:  strings.TrimSpace(req.PassportID),
		FirstName:   strings.TrimSpace(req.FirstName),
		MiddleName:  strings.TrimSpace(req.MiddleName),
		LastName:    strings.TrimSpace(req.LastName),
		DateOfBirth: req.DateOfBirth,
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
		Email:       strings.TrimSpace(req.Email),
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	patients, appErr := h.svc.Search(c.Request.Context(), *claims.HospitalID, claims.HospitalCode, params)
	if appErr != nil {
		util.WriteError(c, appErr)
		return
	}

	util.WriteSuccessWithMeta(c, patients, gin.H{
		"limit":  req.Limit,
		"offset": req.Offset,
		"count":  len(patients),
	})
}
