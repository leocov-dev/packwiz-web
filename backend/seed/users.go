package seed

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"gorm.io/gorm"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
)

func CreateRandomUsers(db *gorm.DB, count int) {
	for i := 0; i < count; i++ {
		pass, _ := utils.HashPassword(gofakeit.Password(
			true,
			true,
			true,
			true,
			true,
			12,
		))

		user := tables.User{
			Username:  fmt.Sprintf("fake_%s", gofakeit.Username()),
			FullName:  gofakeit.Name(),
			Email:     gofakeit.Email(),
			Password:  pass,
			IsActive:  true,
			LinkToken: utils.GenerateLinkToken(16),
		}
		if db.Create(&user).Error == nil {
			_ = user_svc.AssignDefaultRole(db, user.ID) // seed data only; a missing role just leaves a plain user
		}
	}
}
