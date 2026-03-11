package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"tukychat/internal/models"
)

type FriendRepository struct {
	db *pgxpool.Pool
}

func NewFriendRepository(db *pgxpool.Pool) *FriendRepository {
	return &FriendRepository{db: db}
}

func (r *FriendRepository) ListByUserID(ctx context.Context, userID string) ([]models.FriendItem, error) {
	query := `
		select
			p.id,
			p.username,
			p.avatar_url
		from public.friends f
		join public.profiles p
			on p.id = case
				when f.user_low = $1 then f.user_high
				else f.user_low
			end
		where f.user_low = $1 or f.user_high = $1
		order by p.username_lower asc
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := make([]models.FriendItem, 0)

	for rows.Next() {
		var item models.FriendItem
		if err := rows.Scan(
			&item.ID,
			&item.Username,
			&item.AvatarURL,
		); err != nil {
			return nil, err
		}
		friends = append(friends, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return friends, nil
}