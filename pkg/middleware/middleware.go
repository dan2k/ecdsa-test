package middleware

import (
	JWT "mymodule/pkg/jwt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// middleware สำหรับตรวจสอบ token
func Authenticate(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}
	tokenString = tokenString[7:] // Remove "Bearer " prefix
	// ตรวจสอบ token
	token, err := JWT.ParseToken(tokenString)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
	if !token.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
	// ดึงข้อมูล username จาก claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
	// ตรวจสอบว่า token อยู่ใน blacklist หรือไม่
	if JWT.IsTokenBlacklisted(tokenString) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Token is blacklisted"})
	}

	username := claims["username"].(string)
	// บันทึก username ลงใน context
	c.Locals("username", username)

	return c.Next()
}