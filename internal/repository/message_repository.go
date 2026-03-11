package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"tukychat/internal/models"
)

var ErrNotFriends = errors.New("users are not friends")

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func orderedConversation(a, b string) (string, string) {
	if strings.Compare(a, b) < 0 {
		return a, b
	}
	return b, a
}

func (r *MessageRepository) areFriends(ctx context.Context, userA, userB string) (bool, error) {
	low, high := orderedConversation(userA, userB)

	var exists bool
	err := r.db.QueryRow(ctx, `
		select exists(
			select 1
			from public.friends
			where user_low = $1 and user_high = $2
		)
	`, low, high).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *MessageRepository) Create(ctx context.Context, fromUserID, toUserID, content string) (*models.MessageItem, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("empty_content")
	}

	ok, err := r.areFriends(ctx, fromUserID, toUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFriends
	}

	low, high := orderedConversation(fromUserID, toUserID)

	query := `
		insert into public.messages (
			from_user_id,
			to_user_id,
			conversation_low,
			conversation_high,
			content
		)
		values ($1, $2, $3, $4, $5)
		returning id, from_user_id, to_user_id, content, created_at, read_at
	`

	var msg models.MessageItem
	err = r.db.QueryRow(ctx, query, fromUserID, toUserID, low, high, content).Scan(
		&msg.ID,
		&msg.FromUserID,
		&msg.ToUserID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.ReadAt,
	)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *MessageRepository) ListConversation(ctx context.Context, currentUserID, friendID string, limit, offset int) ([]models.MessageItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	ok, err := r.areFriends(ctx, currentUserID, friendID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFriends
	}

	low, high := orderedConversation(currentUserID, friendID)

	query := `
		select
			id,
			from_user_id,
			to_user_id,
			content,
			created_at,
			read_at
		from public.messages
		where conversation_low = $1
		  and conversation_high = $2
		order by created_at desc
		limit $3 offset $4
	`

	rows, err := r.db.Query(ctx, query, low, high, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.MessageItem, 0)

	for rows.Next() {
		var msg models.MessageItem
		if err := rows.Scan(
			&msg.ID,
			&msg.FromUserID,
			&msg.ToUserID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.ReadAt,
		); err != nil {
			return nil, err
		}
		items = append(items, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, currentUserID, friendID string) (int64, error) {
	ok, err := r.areFriends(ctx, currentUserID, friendID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrNotFriends
	}

	tag, err := r.db.Exec(ctx, `
		update public.messages
		set read_at = now()
		where from_user_id = $1
		  and to_user_id = $2
		  and read_at is null
	`, friendID, currentUserID)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

func (r *MessageRepository) GetUnreadCounts(ctx context.Context, currentUserID string) ([]models.UnreadCountItem, error) {
	query := `
		select
			from_user_id as user_id,
			count(*)::int as unread_count
		from public.messages
		where to_user_id = $1
		  and read_at is null
		group by from_user_id
		order by count(*) desc
	`

	rows, err := r.db.Query(ctx, query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.UnreadCountItem, 0)

	for rows.Next() {
		var item models.UnreadCountItem
		if err := rows.Scan(&item.UserID, &item.UnreadCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}