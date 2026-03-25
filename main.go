package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // pprof 端点注册到 DefaultServeMux

	"github.com/gin-gonic/gin"
	"github.com/gin-redis-demo/config"
)

// 增加redis连接的中间件
func RedisMiddleware(c *gin.Context) {
	redisClient := config.NewRedisClient("localhost:6379", "", 0)
	c.Set("redisClient", redisClient)
	c.Next()
}

func main() {
	r := gin.Default()

	// pprof 性能分析端点（独立端口，避免与业务路由冲突）
	go func() {
		log.Println("pprof server listening on :6060")
		log.Println("  - CPU profile:  http://localhost:6060/debug/pprof/")
		log.Println("  - Heap profile: http://localhost:6060/debug/pprof/heap")
		log.Println("  - Goroutine:    http://localhost:6060/debug/pprof/goroutine?debug=1")
		log.Println("  - Block profile:http://localhost:6060/debug/pprof/block?debug=1")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatalf("pprof server failed: %v", err)
		}
	}()

	log.Println("server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
