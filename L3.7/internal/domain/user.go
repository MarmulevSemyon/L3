package domain

// Actor описывает пользователя, который выполняет действие.
//
// Эти данные нужны не только Go-коду, но и SQL-триггерам.
// Перед INSERT, UPDATE или DELETE репозиторий будет передавать их в PostgreSQL
// через session variables.
type Actor struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
