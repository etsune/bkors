package pages

import "github.com/etsune/bkors/server/models"

type PageOptions struct {
	user *models.DBUser
}

func NewPageOptions(user *models.DBUser) *PageOptions {
	return &PageOptions{user}
}
