package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tukychat/internal/models"
)

var ErrProfileNotFound = errors.New("profile not found")
var ErrUsernameTaken = errors.New("username already taken")
var ErrProfileAlreadyExists = errors.New("profile already exists")

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, userID, username string) (*models.Profile, error) {
	username = strings.TrimSpace(username)
	usernameLower := strings.ToLower(username)

	query := `
		insert into public.profiles (id, username, username_lower)
		values ($1, $2, $3)
		returning id, username, avatar_url, bio, created_at, updated_at
	`

	var p models.Profile
	err := r.db.QueryRow(ctx, query, userID, username, usernameLower).Scan(
		&p.ID,
		&p.Username,
		&p.AvatarURL,
		&p.Bio,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if strings.Contains(pgErr.ConstraintName, "profiles_pkey") {
					return nil, ErrProfileAlreadyExists
				}
				return nil, ErrUsernameTaken
			}
		}
		return nil, err
	}

	return &p, nil
}

func (r *ProfileRepository) GetByID(ctx context.Context, userID string) (*models.Profile, error) {
	query := `
		select id, username, avatar_url, bio, created_at, updated_at
		from public.profiles
		where id = $1
	`

	var p models.Profile
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&p.ID,
		&p.Username,
		&p.AvatarURL,
		&p.Bio,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	return &p, nil
}

func (r *ProfileRepository) ListUsers(ctx context.Context, currentUserID, search string, limit, offset int) ([]models.UserListItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	search = strings.TrimSpace(strings.ToLower(search))
	searchPattern := "%"
	if search != "" {
		searchPattern = "%" + search + "%"
	}

	query := `
		select
			p.id,
			p.username,
			p.avatar_url,
			case
				when f.user_low is not null then 'friend'
				when fr_sent.id is not null then 'request_sent'
				when fr_received.id is not null then 'request_received'
				else 'none'
			end as relationship_status
		from public.profiles p
		left join public.friends f
			on (
				((f.user_low = $1 and f.user_high = p.id) or (f.user_high = $1 and f.user_low = p.id))
			)
		left join public.friend_requests fr_sent
			on fr_sent.from_user_id = $1
			and fr_sent.to_user_id = p.id
			and fr_sent.status = 'pending'
		left join public.friend_requests fr_received
			on fr_received.from_user_id = p.id
			and fr_received.to_user_id = $1
			and fr_received.status = 'pending'
		where p.id <> $1
		  and ($2 = '%' or p.username_lower like $2)
		order by p.username_lower asc
		limit $3 offset $4
	`

	rows, err := r.db.Query(ctx, query, currentUserID, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserListItem
	for rows.Next() {
		var u models.UserListItem
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.AvatarURL,
			&u.RelationshipStatus,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}