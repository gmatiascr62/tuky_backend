package models

type SupabaseUser struct {
	ID               string                 `json:"id"`
	Email            string                 `json:"email"`
	Role             string                 `json:"role"`
	Aud              string                 `json:"aud"`
	AppMetadata      map[string]any         `json:"app_metadata"`
	UserMetadata     map[string]any         `json:"user_metadata"`
	EmailConfirmedAt string                 `json:"email_confirmed_at"`
	Phone            string                 `json:"phone"`
	ConfirmedAt      string                 `json:"confirmed_at"`
	LastSignInAt     string                 `json:"last_sign_in_at"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}