package middleware

import (
	"net/http"
	"strings"

	"sme_fin_backend/utils"
)

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.SendErrorResponse(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		
		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.SendErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		
		token := parts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			utils.SendErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		
		// Store claims in request context
		r.Header.Set("X-User-ID", claims.UserID.String())
		r.Header.Set("X-User-Email", claims.Email)
		
		next.ServeHTTP(w, r)
	})
}

