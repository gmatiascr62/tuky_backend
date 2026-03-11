package models

type UserListItem struct {
	ID                 string  `json:"id"`
	Username           string  `json:"username"`
	AvatarURL          *string `json:"avatar_url"`
	RelationshipStatus string  `json:"relationship_status"`
}