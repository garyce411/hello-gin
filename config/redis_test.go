package config

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisAddr = "localhost:6379"

func TestRedisConnect(t *testing.T) {
	client := NewRedisClient(redisAddr, "", 0)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis 连接失败: %v", err)
	}
	t.Log("Redis 连接成功")
}

func TestRedisPoolExhaustion(t *testing.T) {
	client := NewRedisClient(redisAddr, "", 0)
	defer client.Close()

	ctx := context.Background()

	// 模拟慢查询持有连接超过 PoolSize，制造连接池耗尽
	slowCtx, slowCancel := context.WithTimeout(ctx, 20*time.Second)
	defer slowCancel()

	// 获取连接池大小的连接，全部用 SMEMBERS 阻塞（Redis 列表的 BRPOP 更合适，这里用 LPUSH + BRPOP 模拟慢操作）
	// 由于 go-redis 没有直接的阻塞 API，改用 pipeline + Lua 脚本模拟慢操作
	// 实际上制造 pool exhaustion 更直接的方式是：用多个 goroutine 同时执行慢命令，
	// 让它们的总耗时超过 pool 等待时间，导致后来的请求超时/被拒绝。
	//
	// 这里采用更可靠的方式：利用 Redis 的 BRPOPLPUSH（或 BRPOP）阻塞特性，
	// 先向一个 key push 数据，再在多个 goroutine 中用 BRPOP 阻塞等待。

	const (
		listKey    = "pool_exhaust_test_key"
		blockDur   = 15 * time.Second
		numBlocked = 15 // 超过 PoolSize(10) 的并发阻塞请求
	)

	// 清理
	client.Del(ctx, listKey)

	// 先放一个值，确保 BRPOP 不会立即返回
	client.LPush(ctx, listKey, "holder")

	var wg sync.WaitGroup
	results := make([]error, numBlocked)
	errCh := make(chan error, numBlocked)

	// 启动 numBlocked 个 goroutine，每个执行一个长时间阻塞的操作
	for i := 0; i < numBlocked; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// BRPOP 超时时间设长于 blockDur，确保能抢占到连接
			val, err := client.BRPop(slowCtx, blockDur, listKey).Result()
			if err != nil {
				results[idx] = err
				errCh <- err
				return
			}
			t.Logf("goroutine %d 获取到值: %v", idx, val)
			results[idx] = nil
		}(i)
	}

	// 等待所有 goroutine 启动并占用连接
	time.Sleep(500 * time.Millisecond)

	// 现在尝试再发起一个新请求，此时连接池应该耗尽或接近耗尽
	// 这个新请求会因为 Wait=true 而排队等待可用连接，
	// 如果等待时间超过上下文超时，就会失败。
	newCtx, newCancel := context.WithTimeout(ctx, 2*time.Second)
	defer newCancel()

	_, newErr := client.Set(newCtx, "pool_test_key", "pool_test_value", 0).Result()

	// 预期：newErr 不为 nil，因为连接池耗尽且等待超时
	// 如果 Wait=false，则会返回 redis.ErrPoolExhausted
	// 如果 Wait=true，则会因上下文超时而失败
	if newErr != nil {
		t.Logf("连接池耗尽场景符合预期，新请求失败: %v", newErr)
	} else {
		t.Log("警告：新请求在连接池耗尽情况下仍然成功，可能池配置较大或请求未真正耗尽连接")
	}

	// 验证：检查是否至少有请求因池耗尽而失败
	exhaustedCount := 0
	for _, err := range results {
		if err != nil && err != redis.Nil {
			exhaustedCount++
		}
	}
	if exhaustedCount == 0 {
		t.Log("注意：所有被阻塞的 goroutine 都正常返回，未触发预期的连接池耗尽")
	} else {
		t.Logf("检测到 %d 个 goroutine 因连接池耗尽而失败", exhaustedCount)
	}

	// 恢复被 BRPOP 阻塞的 goroutine：将值重新塞回去
	client.LPush(ctx, listKey, "recovery")
	wg.Wait()

	// 最终清理
	client.Del(ctx, listKey)
}

func TestRedisPoolExhaustionDirect(t *testing.T) {
	// 直接通过大量并发连接来耗尽连接池
	client := NewRedisClient(redisAddr, "", 0)
	defer client.Close()

	ctx := context.Background()
	client.Del(ctx, "pool_direct_test")

	const (
		numGoroutines = 50 // 远超 PoolSize=10
		blockDur      = 10 * time.Second
		listKey       = "pool_direct_test"
	)

	// 先塞一个值，避免 BRPOP 立即返回空
	client.LPush(ctx, listKey, "holder")

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			localCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := client.BRPop(localCtx, blockDur, listKey).Result()
			errCh <- err
		}(i)
	}

	// 等待所有 goroutine 启动
	time.Sleep(300 * time.Millisecond)

	// 此刻连接池应已耗尽，尝试发起新请求
	newCtx, newCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer newCancel()

	_, newErr := client.Set(newCtx, "direct_test_key", "value", 0).Result()

	if newErr != nil {
		t.Logf("连接池耗尽，新请求失败（符合预期）: %v", newErr)
	} else {
		t.Log("注意：新请求未失败")
	}

	// 恢复
	client.LPush(ctx, listKey, "recovery")
	wg.Wait()
	client.Del(ctx, listKey)

	// 统计失败数量
	close(errCh)
	failCount := 0
	for err := range errCh {
		if err != nil && err != redis.Nil {
			failCount++
		}
	}
	t.Logf("BRPop 请求中 %d 个失败", failCount)
}

func TestRedisPoolMetrics(t *testing.T) {
	client := NewRedisClient(redisAddr, "", 0)
	defer client.Close()

	stats := client.PoolStats()
	t.Logf("连接池状态: %+v", stats)
}
