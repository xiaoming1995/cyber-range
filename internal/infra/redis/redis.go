package redis

import (
	"context"
	"cyber-range/pkg/config"
	"cyber-range/pkg/logger"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// InitRedis initializes Redis client
func InitRedis(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	Client = client
	logger.Info(ctx, "Redis client initialized successfully")
	return client, nil
}

// Close closes Redis connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Instance state keys
const (
	KeyInstancePrefix      = "instance:"         // instance:{id}
	KeyUserInstancesPrefix = "user_instances:"   // user_instances:{user_id} (SET)
	KeyExpiredInstancesSet = "expired_instances" // ZSET sorted by expiry time
)

// StoreInstance stores instance metadata in Redis with TTL
func StoreInstance(ctx context.Context, instanceID, userID, challengeID, containerID, flag string, port int, expiresAt time.Time) error {
	key := KeyInstancePrefix + instanceID
	data := map[string]interface{}{
		"user_id":      userID,
		"challenge_id": challengeID,
		"container_id": containerID,
		"flag":         flag,
		"port":         port,
		"expires_at":   expiresAt.Unix(),
	}

	// 检查过期时间有效性
	if time.Until(expiresAt) <= 0 {
		return fmt.Errorf("expiry time is in the past")
	}

	pipe := Client.Pipeline()
	pipe.HSet(ctx, key, data)
	// 注意：不设置 TTL，由 Reaper 通过 ZSET 追踪过期并显式删除

	// Add to user's instances set
	userKey := KeyUserInstancesPrefix + userID
	pipe.SAdd(ctx, userKey, instanceID)

	// Add to expiry tracking (sorted set)
	pipe.ZAdd(ctx, KeyExpiredInstancesSet, redis.Z{
		Score:  float64(expiresAt.Unix()),
		Member: instanceID,
	})

	_, err := pipe.Exec(ctx)
	return err
}

// GetInstance retrieves instance metadata
func GetInstance(ctx context.Context, instanceID string) (map[string]string, error) {
	key := KeyInstancePrefix + instanceID
	return Client.HGetAll(ctx, key).Result()
}

// GetUserActiveInstances returns all active instance IDs for a user
func GetUserActiveInstances(ctx context.Context, userID string) ([]string, error) {
	key := KeyUserInstancesPrefix + userID
	return Client.SMembers(ctx, key).Result()
}

// DeleteInstance removes instance from Redis
func DeleteInstance(ctx context.Context, instanceID, userID string) error {
	pipe := Client.Pipeline()
	pipe.Del(ctx, KeyInstancePrefix+instanceID)
	pipe.SRem(ctx, KeyUserInstancesPrefix+userID, instanceID)
	pipe.ZRem(ctx, KeyExpiredInstancesSet, instanceID)
	_, err := pipe.Exec(ctx)
	return err
}

// RemoveFromExpiredSet 仅从过期追踪 ZSET 中移除实例（用于清理残留记录）
func RemoveFromExpiredSet(ctx context.Context, instanceID string) error {
	return Client.ZRem(ctx, KeyExpiredInstancesSet, instanceID).Err()
}

// GetExpiredInstances returns instances that have expired
func GetExpiredInstances(ctx context.Context) ([]string, error) {
	now := time.Now().Unix()

	// Debug: 查看 sorted set 中所有实例
	allInstances, _ := Client.ZRangeWithScores(ctx, KeyExpiredInstancesSet, 0, -1).Result()
	fmt.Printf("[Reaper Debug] Current time: %d, All instances in set: %d\n", now, len(allInstances))
	for _, item := range allInstances {
		member := item.Member.(string)
		score := int64(item.Score)
		expired := score <= now
		fmt.Printf("[Reaper Debug]   Instance: %s, ExpiresAt: %d, Expired: %v\n", member, score, expired)
	}

	return Client.ZRangeByScore(ctx, KeyExpiredInstancesSet, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", now),
	}).Result()
}

// GetInstanceByUserAndChallenge 检查用户是否已有该题目的运行实例
// 返回：实例数据, 错误（如果不存在返回nil, nil）
func GetInstanceByUserAndChallenge(ctx context.Context, userID, challengeID string) (map[string]string, error) {
	// 获取用户所有实例
	instanceIDs, err := GetUserActiveInstances(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 遍历检查每个实例是否属于该题目
	for _, instanceID := range instanceIDs {
		data, err := GetInstance(ctx, instanceID)
		if err != nil {
			continue // 跳过已过期或不存在的实例
		}

		// 检查challenge_id是否匹配
		if data["challenge_id"] == challengeID {
			return data, nil // 找到了该题目的实例
		}
	}

	return nil, nil // 没有找到该题目的实例
}
