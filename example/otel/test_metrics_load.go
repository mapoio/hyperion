package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// 生成 100 条订单请求，用于测试 metrics 数据
func main() {
	const (
		serverURL    = "http://localhost:8090/api/orders"
		totalOrders  = 100
		concurrency  = 10 // 并发请求数
	)

	fmt.Printf("🚀 开始发送 %d 条订单请求...\n", totalOrders)
	fmt.Printf("📊 服务器地址: %s\n", serverURL)
	fmt.Printf("⚡ 并发数: %d\n\n", concurrency)

	start := time.Now()

	// 使用信号量控制并发
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// 统计结果
	var (
		successCount int
		failCount    int
		mu           sync.Mutex
	)

	// 产品列表
	products := []string{"laptop", "phone", "tablet", "monitor", "keyboard", "mouse"}

	for i := 1; i <= totalOrders; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(orderNum int) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			// 构造请求数据
			req := map[string]interface{}{
				"user_id":    fmt.Sprintf("user-%d", rand.Intn(20)+1), // 20 个不同用户
				"product_id": products[rand.Intn(len(products))],      // 随机产品
				"amount":     float64(rand.Intn(500)+50) / 100.0,      // 0.50 到 5.50 之间
			}

			reqBody, _ := json.Marshal(req)

			// 发送 HTTP POST 请求
			resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				fmt.Printf("❌ [%3d] 请求失败: %v\n", orderNum, err)
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				successCount++
				fmt.Printf("✅ [%3d] 订单创建成功 - User: %s, Product: %s, Amount: %.2f\n",
					orderNum, req["user_id"], req["product_id"], req["amount"])
			} else {
				failCount++
				fmt.Printf("❌ [%3d] 订单创建失败 - Status: %d\n", orderNum, resp.StatusCode)
			}
			mu.Unlock()

			// 随机延迟 10-100ms，模拟真实场景
			time.Sleep(time.Duration(rand.Intn(90)+10) * time.Millisecond)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// 打印统计结果
	fmt.Printf("\n%s\n", "============================================================")
	fmt.Printf("📈 测试完成!\n")
	fmt.Printf("⏱️  总耗时: %v\n", elapsed)
	fmt.Printf("✅ 成功: %d\n", successCount)
	fmt.Printf("❌ 失败: %d\n", failCount)
	fmt.Printf("📊 成功率: %.2f%%\n", float64(successCount)/float64(totalOrders)*100)
	fmt.Printf("🚀 平均 QPS: %.2f\n", float64(totalOrders)/elapsed.Seconds())
	fmt.Printf("%s\n", "============================================================")
}
