package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
)

type contextKey struct{}

// UserClaims holds the parsed JWT claims relevant to the application.
type UserClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

// UserClaimsFromContext retrieves UserClaims from the context, or nil if not present.
func UserClaimsFromContext(ctx context.Context) *UserClaims {
	claims, _ := ctx.Value(contextKey{}).(*UserClaims)
	return claims
}

// NewAuthMiddleware creates a gorilla/mux middleware that validates OIDC JWT Bearer tokens.
func NewAuthMiddleware(ctx context.Context, issuerURL, clientID string) (mux.MiddlewareFunc, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc provider discovery failed for %q: %w", issuerURL, err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := extractBearerToken(r)
			if !ok {
				unauthorized(w, "missing authorization header")
				return
			}

			idToken, err := verifier.Verify(r.Context(), rawToken)
			if err != nil {
				slog.WarnContext(r.Context(), "JWT verification failed", slog.String("error", err.Error()))
				unauthorized(w, "invalid or expired token")
				return
			}

			var claims UserClaims
			if err := idToken.Claims(&claims); err != nil {
				slog.WarnContext(r.Context(), "JWT claims extraction failed", slog.String("error", err.Error()))
				unauthorized(w, "invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), contextKey{}, &claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}

func extractBearerToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="printyl"`)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "unauthorized",
		"message": message,
	})
}
