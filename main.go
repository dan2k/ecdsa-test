package main
import (
	"mymodule/pkg/handler"
	"mymodule/pkg/jwt"
)
func main() {
	handler.Init();
	defer jwt.RedisClient.Close()
	// สร้าง goroutine สำหรับลบ token ที่หมดอายุแล้ว
	go jwt.RemoveExpiredTokens()
}

