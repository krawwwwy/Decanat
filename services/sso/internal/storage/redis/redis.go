package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"sso/internal/domain/models"
)

type Storage struct {
	client *redis.Client
}

func New(client *redis.Client) *Storage {
	return &Storage{client: client}
}

func (r *Storage) SavePendingUser(ctx context.Context, id string, user *models.PendingUser) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("pending_user:%s", id)
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (r *Storage) GetPendingUser(ctx context.Context, id string) (*models.PendingUser, error) {
	key := fmt.Sprintf("pending_user:%s", id)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user models.PendingUser
	err = json.Unmarshal([]byte(val), &user)
	return &user, err
}

func (r *Storage) DeletePendingUser(ctx context.Context, id string) error {
	key := fmt.Sprintf("pending_user:%s", id)
	return r.client.Del(ctx, key).Err()
}

func (r *Storage) ListPendingUsers(ctx context.Context) ([]*models.PendingUser, error) {
	keys, err := r.client.Keys(ctx, "pending_user:*").Result()
	if err != nil {
		return nil, err
	}

	var users []*models.PendingUser
	for _, key := range keys {
		val, _ := r.client.Get(ctx, key).Result()
		var user models.PendingUser
		json.Unmarshal([]byte(val), &user)
		users = append(users, &user)
	}
	return users, nil
}
