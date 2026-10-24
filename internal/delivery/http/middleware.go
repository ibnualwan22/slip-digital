package http

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/ibnualwan/bisyaroh/pkg/response"
)

// JWTMiddleware validates the JWT token in the Authorization header
func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Error(c, http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return response.Error(c, http.StatusUnauthorized, "invalid authorization format")
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			}

			// Token valid: paparkan klaim identitas ke context untuk
			// middleware peran dan handler di hilir.
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if v, ok := claims["role"].(string); ok {
					c.Set("authRole", v)
				}
				if v, ok := claims["employee_id"].(string); ok {
					c.Set("authEmployeeID", v)
				}
				if v, ok := claims["siakad_id"].(string); ok {
					c.Set("authSiakadID", v)
				}
				if v, ok := claims["name"].(string); ok {
					c.Set("authName", v)
				}
				if v, ok := claims["sub"].(string); ok {
					c.Set("authUsername", v)
				}
			}

			return next(c)
		}
	}
}

// RequireRole menolak token yang perannya tidak sesuai.
// Inilah pagar server antara akun admin dan akun pengajar:
// token pengajar TIDAK PERNAH bisa melewati grup admin, dan sebaliknya.
func RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("authRole") != role {
				return response.Error(c, http.StatusForbidden, "akses ditolak untuk peran ini")
			}
			return next(c)
		}
	}
}
