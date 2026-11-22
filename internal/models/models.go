package models

import "time"

type User struct {
	ID         int64     `json:"id"`
	Email      string    `json:"email"`
	Timezone   string    `json:"timezone"`
	UILanguage string    `json:"ui_language"`
	CreatedAt  time.Time `json:"created_at"`
}
