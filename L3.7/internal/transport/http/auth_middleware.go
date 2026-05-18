package http

import (
	stdhttp "net/http"
	"strings"

	"warehousecontrol/internal/auth"
	"warehousecontrol/internal/domain"

	"github.com/wb-go/wbf/ginext"
)

const actorContextKey = "actor"

// AuthMiddleware проверяет JWT и кладёт пользователя в Gin context.
func AuthMiddleware(secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginext.H{
				"error": "authorization header is required",
			})
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginext.H{
				"error": "authorization header must start with Bearer",
			})
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if tokenString == "" {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginext.H{
				"error": "token is required",
			})
			return
		}

		actor, err := auth.ParseToken(tokenString, secret)
		if err != nil {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginext.H{
				"error": "invalid token",
			})
			return
		}

		c.Set(actorContextKey, actor)
		c.Next()
	}
}

// ActorFromContext достаёт пользователя из Gin context.
func ActorFromContext(c *ginext.Context) (domain.Actor, bool) {
	value, exists := c.Get(actorContextKey)
	if !exists {
		return domain.Actor{}, false
	}

	actor, ok := value.(domain.Actor)
	if !ok {
		return domain.Actor{}, false
	}

	return actor, true
}
