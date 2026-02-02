package users

import "time"

// UserProfile — доменная модель профиля пользователя.
type UserProfile struct {
	UserID      string
	DisplayName string
	AvatarURL   *string
	Bio         *string
	Locale      string
	Timezone    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
