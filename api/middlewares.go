package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/jwt"
)

func (s *Server) statsdMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start).Milliseconds()

		// Send metrics to statsd
		_ = s.sdClient.Incr("http.requests", []string{"path:" + c.Path()}, 1)
		_ = s.sdClient.Timing("http.response_time", time.Duration(duration)*time.Millisecond, []string{"path:" + c.Path()}, 1)
		_ = s.sdClient.Incr("http.status."+fmt.Sprint(c.Response().Status), []string{"path:" + c.Path(), "method:" + c.Request().Method}, 1)

		return err
	}
}

func (s *Server) userAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg, err := config.ReadConfig("config-verifier")
		if err != nil {
			s.logger.Error("Failed to read verifier config: ", err)
		}

		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing token"})
		}

		tokenStr := authHeader[len("Bearer "):]

		// parse and validate JWT
		userID, err := jwt.ValidateJWT(tokenStr, cfg.Server.UserAuth.JwtSecret)
		if err != nil {
			s.logger.Error("Failed to parse jwt: ", err)
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
		}

		user, err := s.db.FindUserById(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "User not found"})
		}

		// store a pointer to authenticated user model for usage in next handlers
		c.Set("user", user)

		return next(c)
	}
}

func (s *Server) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing Authorization header"})
		}

		tokenStr := authHeader[len("Bearer "):]
		_, err := s.authService.ValidateToken(tokenStr)
		if err != nil {
			s.logger.Warnf("fail to validate token, err: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		s.logger.Info("Token validated successfully")
		return next(c)
	}
}
