package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/team-task-api/internal/model"
)

const (
	taskListPrefix = "tasks:list:"
	teamKeysPrefix = "tasks:team:"
	ttl            = 5 * time.Minute
)

type TaskCache struct {
	rdb *redis.Client
}

func NewTaskCache(rdb *redis.Client) *TaskCache {
	return &TaskCache{rdb: rdb}
}

type ListKey struct {
	TeamIDs   []string
	Status    string
	AssigneeID string
	Limit     int
	Offset    int
}

func (c *TaskCache) buildKey(k ListKey) string {
	sorted := make([]string, len(k.TeamIDs))
	copy(sorted, k.TeamIDs)
	sort.Strings(sorted)

	parts := []string{
		"teams=" + strings.Join(sorted, ","),
		"status=" + k.Status,
		"assignee=" + k.AssigneeID,
		"limit=" + strconv.Itoa(k.Limit),
		"offset=" + strconv.Itoa(k.Offset),
	}
	raw := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(raw))
	return taskListPrefix + fmt.Sprintf("%x", hash)
}

func (c *TaskCache) Get(ctx context.Context, k ListKey) ([]model.Task, error) {
	key := c.buildKey(k)
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var tasks []model.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}
	return tasks, nil
}

func (c *TaskCache) Set(ctx context.Context, k ListKey, tasks []model.Task) error {
	key := c.buildKey(k)
	data, err := json.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	pipe := c.rdb.Pipeline()
	for _, teamID := range k.TeamIDs {
		pipe.SAdd(ctx, teamKeysPrefix+teamID, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("track cache keys: %w", err)
	}

	return nil
}

func (c *TaskCache) InvalidateTeam(ctx context.Context, teamID string) error {
	setKey := teamKeysPrefix + teamID
	members, err := c.rdb.SMembers(ctx, setKey).Result()
	if err != nil {
		return fmt.Errorf("get team cache keys: %w", err)
	}

	if len(members) > 0 {
		if err := c.rdb.Del(ctx, members...).Err(); err != nil {
			return fmt.Errorf("delete cache keys: %w", err)
		}
	}

	if err := c.rdb.Del(ctx, setKey).Err(); err != nil {
		return fmt.Errorf("delete team key set: %w", err)
	}

	return nil
}
