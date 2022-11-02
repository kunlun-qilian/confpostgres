package tests

//go:generate tools gen model2 User --database DB
// @def primary ID
// @def unique_index I_user_id UserID
type User struct {
	PrimaryID
	RefUser
}

type PrimaryID struct {
	ID uint `db:"F_id,autoincrement" json:"-"`
}

type RefUser struct {
	UserID string `db:"F_user_id,size=100"`
}
