package model

type User struct {
	ID           int64  `db:"id"`
	UserName     string `db:"username"`
	Email        string `db:"email"`
	Password     string `db:"password"`
	RefreshToken string `db:"refresh_token"`
}
