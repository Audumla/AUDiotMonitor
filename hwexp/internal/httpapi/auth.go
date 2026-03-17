package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Token struct {
	ID     string   `yaml:"id"`
	Token  string   `yaml:"token"`
	Scopes []string `yaml:"scopes"`
}

type AuthStore struct {
	tokens map[string]Token
}

func LoadAuthStore(path string) (*AuthStore, error) {
	if path == "" {
		return &AuthStore{tokens: make(map[string]Token)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	var wrapper struct {
		Tokens []Token `yaml:"tokens"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse tokens file: %w", err)
	}

	store := &AuthStore{tokens: make(map[string]Token)}
	for _, t := range wrapper.Tokens {
		store.tokens[t.Token] = t
	}

	return store, nil
}

func (s *AuthStore) Middleware(next http.HandlerFunc, requiredScope string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.jsonError(w, http.StatusUnauthorized, "HTTP_UNAUTHORIZED", "Missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			s.jsonError(w, http.StatusUnauthorized, "HTTP_UNAUTHORIZED", "Invalid authorization header format")
			return
		}

		token, exists := s.tokens[parts[1]]
		if !exists {
			s.jsonError(w, http.StatusUnauthorized, "HTTP_UNAUTHORIZED", "Invalid token")
			return
		}

		if requiredScope != "" {
			hasScope := false
			for _, s := range token.Scopes {
				if s == requiredScope {
					hasScope = true
					break
				}
			}
			if !hasScope {
				s.jsonError(w, http.StatusForbidden, "HTTP_FORBIDDEN", "Missing required scope: "+requiredScope)
				return
			}
		}

		next(w, r)
	}
}

func (s *AuthStore) jsonError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
