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

var ErrCannotFriendSelf = errors.New("cannot send friend request to self")
var ErrTargetProfileNotFound = errors.New("target profile not found")
var ErrAlreadyFriends = errors.New("already friends")
var ErrPendingRequestAlreadyExists = errors.New("pending friend request already exists")

type FriendRequestRepository struct {
	db *pgxpool.Pool
}

func NewFriendRequestRepository(db *pgxpool.Pool) *FriendRequestRepository {
	return &FriendRequestRepository{db: db}
}

func orderedPair(a, b string) (string, string) {
	if strings.Compare(a, b) < 0 {
		return a, b
	}
	return b, a
}

func (r *FriendRequestRepository) Create(ctx context.Context, fromUserID, toUserID string) (string, error) {
	if fromUserID == toUserID {
		return "", ErrCannotFriendSelf
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// Verificar que exista el perfil destino
	var exists bool
	err = tx.QueryRow(ctx, `
		select exists(
			select 1
			from public.profiles
			where id = $1
		)
	`, toUserID).Scan(&exists)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrTargetProfileNotFound
	}

	low, high := orderedPair(fromUserID, toUserID)

	// Verificar si ya son amigos
	err = tx.QueryRow(ctx, `
		select exists(
			select 1
			from public.friends
			where user_low = $1 and user_high = $2
		)
	`, low, high).Scan(&exists)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrAlreadyFriends
	}

	// Insertar solicitud pendiente
	var requestID string
	err = tx.QueryRow(ctx, `
		insert into public.friend_requests (
			from_user_id,
			to_user_id,
			pair_user_low,
			pair_user_high,
			status
		)
		values ($1, $2, $3, $4, 'pending')
		returning id
	`, fromUserID, toUserID, low, high).Scan(&requestID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// unique index parcial de pending
			if pgErr.Code == "23505" {
				return "", ErrPendingRequestAlreadyExists
			}
		}
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return requestID, nil
}

func (r *FriendRequestRepository) ListReceived(ctx context.Context, userID string) ([]models.FriendRequestItem, error) {

	query := `
		select
			fr.id,
			fr.from_user_id,
			p.username,
			p.avatar_url,
			fr.status,
			fr.created_at
		from public.friend_requests fr
		join public.profiles p
			on p.id = fr.from_user_id
		where fr.to_user_id = $1
		  and fr.status = 'pending'
		order by fr.created_at desc
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.FriendRequestItem

	for rows.Next() {
		var item models.FriendRequestItem

		err := rows.Scan(
			&item.ID,
			&item.FromID,
			&item.Username,
			&item.AvatarURL,
			&item.Status,
			&item.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		requests = append(requests, item)
	}

	return requests, nil
}

func (r *FriendRequestRepository) Accept(ctx context.Context, requestID, currentUserID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var fromUserID, toUserID, status string

	err = tx.QueryRow(ctx, `
		select from_user_id, to_user_id, status
		from public.friend_requests
		where id = $1
		for update
	`, requestID).Scan(&fromUserID, &toUserID, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTargetProfileNotFound
		}
		return err
	}

	if toUserID != currentUserID {
		return errors.New("not allowed to accept this request")
	}

	if status != "pending" {
		return errors.New("request already processed")
	}

	low, high := orderedPair(fromUserID, toUserID)

	var alreadyFriends bool
	err = tx.QueryRow(ctx, `
		select exists(
			select 1
			from public.friends
			where user_low = $1 and user_high = $2
		)
	`, low, high).Scan(&alreadyFriends)
	if err != nil {
		return err
	}
	if alreadyFriends {
		return ErrAlreadyFriends
	}

	_, err = tx.Exec(ctx, `
		update public.friend_requests
		set status = 'accepted',
		    responded_at = now()
		where id = $1
	`, requestID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		insert into public.friends (user_low, user_high)
		values ($1, $2)
	`, low, high)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}