package controller

import (
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"
)

// BaseController holds useful common methods for controllers.
type BaseController struct{}

func (bc BaseController) getUserID(r *web.Ctx) (uuid.UUID, error) {
	return uuid.Parse(r.Session.UserID)
}
