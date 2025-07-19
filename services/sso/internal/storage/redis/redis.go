package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sso/internal/storage"
	"time"

	"github.com/redis/go-redis/v9"
	"sso/internal/domain/models"
)

type Redis struct {
	client *redis.Client
}

func New(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) SavePendingUser(ctx context.Context, user models.PendingUser) error {
	id := uuid.New().String()

	data, err := json.Marshal(user)
	if err != nil {
		return storage.ErrInvalidUserData
	}

	key := fmt.Sprintf("pending_user:%s", id)
	err = r.client.Set(ctx, key, data, 24*7*time.Hour).Err()
	if err != nil {
		return storage.ErrInvalidUserData
	}

	return nil
}

func (r *Redis) DeletePendingUser(ctx context.Context, id string) error {
	key := fmt.Sprintf("pending_user:%s", id)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return storage.ErrNoValueForKey
		}

		return err
	}
	return nil
}

func (r *Redis) PendingUser(ctx context.Context, userID string) (models.PendingUser, error) {
	val, err := r.client.Get(ctx, userID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.PendingUser{}, storage.ErrNoValueForKey
		}

		return models.PendingUser{}, err
	}

	var user models.PendingUser
	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		return models.PendingUser{}, storage.ErrInvalidUserData
	}

	return user, nil
}

func (r *Redis) ListPendingUsers(ctx context.Context) ([]models.PendingUser, error) {
	keys, err := r.client.Keys(ctx, "pending_user:*").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	var users []models.PendingUser
	for _, key := range keys {

		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, storage.ErrNoValueForKey
			}

			return nil, err
		}

		var user models.PendingUser

		err = json.Unmarshal([]byte(val), &user)
		if err != nil {
			return nil, storage.ErrInvalidUserData
		}

		users = append(users, user)
	}
	return users, nil
}
