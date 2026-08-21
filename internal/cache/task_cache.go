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

	"github.com/team-task-api/internal/config"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/circuitbreaker"
)

const (
	taskListPrefix = "tasks:list:"
	teamKeysPrefix = "tasks:team:"
)

type TaskCache struct {
	rdb *redis.Client
	cb  *circuitbreaker.CircuitBreaker
	ttl time.Duration
}

func NewTaskCache(rdb *redis.Client, cfg config.CacheConfig) *TaskCache {
	return &TaskCache{
		rdb: rdb,
		cb:  circuitbreaker.New(cfg.CBFailures, cfg.CBSuccesses, time.Duration(cfg.CBTimeoutSec)*time.Second),
		ttl: time.Duration(cfg.TTLMin) * time.Minute,
	}
}

type ListKey struct {
	TeamIDs   []string
	Status    string
	AssigneeID string
	Limit     int
	Offset    int
}

func (k ListKey) String() string {
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
	return strings.Join(parts, "|")
}

func (c *TaskCache) buildKey(k ListKey) string {
	raw := k.String()
	hash := sha256.Sum256([]byte(raw))
	return taskListPrefix + fmt.Sprintf("%x", hash)
}

func (c *TaskCache) Get(ctx context.Context, k ListKey) ([]model.Task, error) {
	key := c.buildKey(k)

	var tasks []model.Task
	err := c.cb.Execute(func() error {
		data, err := c.rdb.Get(ctx, key).Bytes()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("redis get: %w", err)
		}
		return json.Unmarshal(data, &tasks)
	})

	if err == circuitbreaker.ErrOpen {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *TaskCache) Set(ctx context.Context, k ListKey, tasks []model.Task) error {
	key := c.buildKey(k)
	data, err := json.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}

	err = c.cb.Execute(func() error {
		pipe := c.rdb.TxPipeline()
		pipe.Set(ctx, key, data, c.ttl)
		for _, teamID := range k.TeamIDs {
			pipe.SAdd(ctx, teamKeysPrefix+teamID, key)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("cache set: %w", err)
		}
		return nil
	})

	if err == circuitbreaker.ErrOpen {
		return nil
	}
	return err
}

var invalidateScript = redis.NewScript(`
local setKey = KEYS[1]
local members = redis.call('smembers', setKey)
if #members > 0 then
	redis.call('del', unpack(members))
end
redis.call('del', setKey)
return #members
`)

func (c *TaskCache) InvalidateTeam(ctx context.Context, teamID string) error {
	setKey := teamKeysPrefix + teamID

	err := c.cb.Execute(func() error {
		if _, err := invalidateScript.Run(ctx, c.rdb, []string{setKey}).Result(); err != nil {
			return fmt.Errorf("invalidate team cache: %w", err)
		}
		return nil
	})

	if err == circuitbreaker.ErrOpen {
		return nil
	}
	return err
}
