package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	us      UserRetrierver
	authKey []byte
	passKey string
}

func NewAuthHandler(us UserRetrierver, authKey string, passKey string) *AuthHandler {
	return &AuthHandler{
		authKey: []byte(authKey),
		passKey: passKey,
		us:      us,
	}
}

func (h *AuthHandler) InitRoute(app *fiber.App) {
	router := app.Group("/login")
	router.Post("/", h.Login)
	app.Use(h.AuthMiddleware) // todo. 위치 이동 시키자
}

// Claims represents the JWT claims
type Claims struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {

	var req LoginRequest
	err := c.BodyParser(&req)
	if err != nil {
		return err
	}

	user, err := h.us.User(req.Username)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return err
	}

	// Create token expiration time (24 hours from now)
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.authKey)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(JWTResponse{
		Token:  tokenString,
		Expiry: expirationTime.Unix(),
	})
}

func (h *AuthHandler) AuthMiddleware(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return errors.New("authorization header missing")
	} else if authHeader == h.passKey { // bot 혹은 HTTP Req 테스트용 키
		return c.Next()
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return errors.New("invalid authorization format")
	}

	tokenString := tokenParts[1]

	// Parse and validate the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.authKey, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	if c.Method() != "GET" && !claims.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden",
		})
	}

	return c.Next()
}
