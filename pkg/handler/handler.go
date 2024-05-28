package handler

import (
	"encoding/json"
	"net/http"
	"time"
	 "github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	JWT "mymodule/pkg/jwt"
	"mymodule/pkg/middleware"
	"mymodule/pkg/ecdsa"
)

// User struct สำหรับเก็บข้อมูลผู้ใช้
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignRequest struct {
	Data string `json:"data"`
}
type VerifyRequest struct {
	Data         string `json:"data"`
	SignatureHex string `json:"signatureHex"`
	PublicKeyHex string `json:"publicKeyHex"`
}
func SignHandler(c *fiber.Ctx) error {
	// Parse the request body into the SignRequest struct
	var request SignRequest
	if err := c.BodyParser(&request); err != nil {
		return err
	}

	// Call the sign function with the data from the request
	signatureHex, publicKeyHex, privateKeyHex, err := ecdsa.Sign(request.Data)
	if err != nil {
		return err
	}

	// Return the signature, public key, and private key
	return c.JSON(fiber.Map{
		"signatureHex": signatureHex,
		"publicKeyHex": publicKeyHex,
		"privateKeyHex": privateKeyHex,
	})
}

func VerifyHandler(c *fiber.Ctx) error {
	// Parse the request body into the VerifyRequest struct
	var request VerifyRequest
	if err := c.BodyParser(&request); err != nil {
		return err
	}
	// Call the verify function with the data from the request
	valid, err := ecdsa.Verify(request.Data, request.SignatureHex, request.PublicKeyHex)
	if err != nil {
		return err
	}

	// Return the result of the verification
	return c.JSON(fiber.Map{
		"valid": valid,
	})
}

// Login route
func login(c *fiber.Ctx) error {
	var user User
	if err := json.Unmarshal(c.Body(), &user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ตรวจสอบ Username และ Password
	// (ควรใช้ DB สำหรับเก็บข้อมูล User และตรวจสอบในส่วนนี้)
	if user.Username != "test" || user.Password != "test" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// สร้าง Access Token
	accessToken, err := JWT.GenerateToken(user.Username, "access", time.Minute*15)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate access token"})
	}

	// สร้าง Refresh Token
	refreshToken, err := JWT.GenerateToken(user.Username, "refresh", time.Hour*24*7)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate refresh token"})
	}

	// ส่งกลับ Access Token & Refresh Token
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}



// Refresh token route
func refreshToken(c *fiber.Ctx) error {
	refreshTokenString := c.Get("Authorization")
	if refreshTokenString == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}
	refreshTokenString = refreshTokenString[7:] // Remove "Bearer " prefix
	// ตรวจสอบ Refresh Token
	refreshToken, err := JWT.ParseToken(refreshTokenString)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}
	// ตรวจสอบว่า Refresh Token หมดอายุหรือไม่
	if !refreshToken.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh token expired"})
	}
	claims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	// ดึงข้อมูล username จาก claims
	username := claims["username"].(string)

	// สร้าง Access Token ใหม่
	accessToken, err := JWT.GenerateToken(username, "access", time.Minute*15)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate access token"})
	}

	// ส่งกลับ Access Token ใหม่
	return c.Status(http.StatusOK).JSON(fiber.Map{"access_token": accessToken})
}
func testHandler(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}
func Init() {
	// สร้าง Fiber app
	app := fiber.New()

	// ใช้ middleware สำหรับ logging และ recover
	app.Use(logger.New())
	app.Use(recover.New())

	// สร้าง Group สำหรับ route ที่ต้องการ authenticate
	api := app.Group("/api", middleware.Authenticate)
	// api.Get("/verify", verifyToken)
	api.Post("/refresh", refreshToken)
	api.Post("/sign",SignHandler)
	api.Post("/verify",VerifyHandler)
	// Define routes อื่น ๆ
	app.Post("/login", login)
	app.Get("/",testHandler)

	// Start server
	app.Listen(":3000")
}







