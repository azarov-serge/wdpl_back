package users

// UserProfileResponse — ответ с профилем пользователя (GET /api/users/me, PUT /api/users/me).
type UserProfileResponse struct {
	UserID      string  `json:"userID"`
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarURL,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Locale      string  `json:"locale"`
	Timezone    string  `json:"timezone"`
}

// UpdateProfileRequest — тело запроса PUT /api/users/me.
type UpdateProfileRequest struct {
	DisplayName *string `json:"displayName" validate:"omitempty,min=2,max=100"`
	AvatarURL   *string `json:"avatarURL"`
	Bio         *string `json:"bio" validate:"omitempty,max=500"`
	Locale      *string `json:"locale"`
	Timezone    *string `json:"timezone"`
}
