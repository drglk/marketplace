package mapper

import (
	"marketplace/internal/entities"
	"marketplace/internal/models"
)

func UserByEntity(rawUser *entities.User) *models.User {
	return &models.User{
		ID:       rawUser.ID,
		Login:    rawUser.Login,
		PassHash: rawUser.PassHash,
	}
}
