package seed

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"gorm.io/gorm"
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

		db.Create(
			&tables.User{
				Username:  fmt.Sprintf("fake_%s", gofakeit.Username()),
				FullName:  gofakeit.Name(),
				Email:     gofakeit.Email(),
				Password:  pass,
				IsAdmin:   false,
				LinkToken: utils.GenerateRandomString(32),
			},
		)
	}
}
