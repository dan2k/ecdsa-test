package  jwt;

import(
	"time"
	"github.com/golang-jwt/jwt/v4"
	"fmt"
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