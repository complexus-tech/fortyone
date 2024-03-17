package projects

type dbProjecwt struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Owner       string `db:"owner"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
	DeletedAt   string `db:"deleted_at"`
}
