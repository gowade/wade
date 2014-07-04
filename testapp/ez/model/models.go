package model

const (
	RoleAnonymous = 1 << iota
	RoleUser      = 1 << iota
	RoleModerator = 1 << iota
	RoleAdmin     = 1 << iota
)

type Post struct {
	Id      int64  `db:"post_id"`
	Title   string `db:"title"`
	Content string `db:"content"`
}

type User struct {
	Id       int64  `db:"user_id"`
	Username string `db:"username"  json:"username"`
	Role     uint32 `db:"role"`
	Token    string `db:"token" json:"token"`
	Password string `db:"password" json:"password"`
}
