package  jwt;

import(
	"time"
	"github.com/golang-jwt/jwt/v4"
	"github.com/go-redis/redis/v8"
	"fmt"
	"errors"
	"context"
)
// JWT claims struct
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// JWT key
var secretKey = []byte("your-secret-key")
// สร้าง JWT token
func GenerateToken(username, tokenType string, duration time.Duration) (string, error) {
	_ = tokenType
	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // ใช้ jwt.NewWithClaims()

	// กำหนด key สำหรับสร้าง token
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
// ตรวจสอบและแปลง JWT token
func ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}
// Verify token route
// func verifyToken(c *fiber.Ctx) error {
// 	tokenString := c.Get("Authorization")
// 	if tokenString == "" {
// 		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
// 	}

// 	tokenString = tokenString[7:] // Remove "Bearer " prefix

// 	// ตรวจสอบ token
// 	token, err := parseToken(tokenString)
// 	if err != nil {
// 		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
// 	}

// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok || !token.Valid {
// 		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
// 	}

// 	// ดึงข้อมูล username จาก claims
// 	username := claims["username"].(string)

// 	// ส่งกลับ username
// 	return c.Status(http.StatusOK).JSON(fiber.Map{"username": username})
// }

// เชื่อมต่อ Redis
var RedisClient *redis.Client
// ฟังก์ชันสำหรับ blacklist token
func BlacklistToken(token string) error {
	// ดึงข้อมูล expiry จาก token
	parsedToken, err := ParseToken(token)
	if err != nil {
		return err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token")
	}

	expiry := time.Unix(int64(claims["exp"].(float64)), 0).Sub(time.Now())

	// เก็บ token ใน Redis blacklist พร้อมกำหนด expiry
	err = RedisClient.SetEX(context.Background(), token, "blacklisted", expiry).Err()
	if err != nil {
		return err
	}

	return nil
}

// ฟังก์ชันสำหรับตรวจสอบว่า token อยู่ใน blacklist หรือไม่
func IsTokenBlacklisted(token string) bool {
	exists, err := RedisClient.Exists(context.Background(), token).Result()
	if err != nil {
		return false // สมมติว่า token ไม่ได้อยู่ใน blacklist หากเกิด error
	}
	return exists == 1
}
// ฟังก์ชันสำหรับลบ token ออกจาก blacklist
func RemoveFromBlacklist(token string) error {
	return RedisClient.Del(context.Background(), token).Err()
}
func RemoveExpiredTokens() {
	for {
		// ดึง key ทั้งหมดที่ตรงกับ pattern "blacklisted:*"
		keys, err := RedisClient.Keys(context.Background(), "blacklisted:*").Result()
		if err != nil {
			fmt.Println("Error fetching keys:", err)
			// handle error appropriately
			time.Sleep(time.Minute) // รอ 1 นาทีก่อนลองใหม่
			continue
		}

		// Loop ผ่าน keys
		for _, key := range keys {
			// ดึง expiry time ของ key
			ttl, err := RedisClient.TTL(context.Background(), key).Result()
			if err != nil {
				fmt.Printf("Error fetching TTL for key %s: %v\n", key, err)
				// handle error appropriately
				continue
			}

			// ถ้า token หมดอายุแล้ว ให้ลบออกจาก blacklist
			if ttl <= 0 {
				err := RemoveFromBlacklist(key)
				if err != nil {
					fmt.Printf("Error removing key %s from blacklist: %v\n", key, err)
					// handle error appropriately
				} else {
					fmt.Printf("Removed expired token %s from blacklist\n", key)
				}
			}
		}

		// รอ 1 นาทีก่อนทำการลบ token ที่หมดอายุแล้วครั้งถัดไป
		time.Sleep(time.Minute)
	}
}