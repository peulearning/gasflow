package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// Authenticate valida o Bearer token e injeta os claims no contexto.
func (s *Service) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, `{"error":"token ausente"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"formato inválido: use Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		claims, err := s.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"error":"token inválido ou expirado"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorize retorna um middleware que exige um dos roles fornecidos.
func (s *Service) Authorize(roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromCtx(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"não autenticado"}`, http.StatusUnauthorized)
				return
			}
			if err := RequireRole(claims, roles...); err != nil {
				http.Error(w, `{"error":"acesso negado"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromCtx extrai os claims do contexto.
func ClaimsFromCtx(ctx context.Context) *Claims {
	v, _ := ctx.Value(ClaimsKey).(*Claims)
	return v
}