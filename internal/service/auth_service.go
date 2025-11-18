package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type AuthService struct {
	redis          *redis.Client
	authServiceURL string
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role,omitempty"`
}

func NewAuthService(redisClient *redis.Client, authServiceURL string) *AuthService {
	return &AuthService{
		redis:          redisClient,
		authServiceURL: authServiceURL,
	}
}

// ValidateToken valida un token con el servicio de autenticación
// Primero busca en caché de Redis, si no encuentra llama al servicio
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*UserResponse, error) {
	// 1. Intentar obtener del caché
	cacheKey := fmt.Sprintf("auth:token:%s", token)
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// Token encontrado en caché
		var user UserResponse
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			return &user, nil
		}
	}

	// 2. Si no está en caché, validar con el servicio de auth
	user, err := s.callAuthService(token)
	if err != nil {
		return nil, err
	}

	// 3. Guardar en caché (TTL de 10 minutos)
	userData, _ := json.Marshal(user)
	s.redis.Set(ctx, cacheKey, userData, 10*time.Minute)

	return user, nil
}

// callAuthService llama al microservicio de autenticación para validar el token
func (s *AuthService) callAuthService(token string) (*UserResponse, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// Configurar request
	req.SetRequestURI(s.authServiceURL + "/users/current")
	req.Header.SetMethod("GET")
	req.Header.Set("Authorization", "Bearer "+token)

	// Cliente con timeout
	client := &fasthttp.Client{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Realizar petición
	err := client.Do(req, resp)
	if err != nil {
		return nil, fmt.Errorf("auth service error: %w", err)
	}

	// Verificar status code
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Parsear respuesta
	var user UserResponse
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}

	return &user, nil
}

// InvalidateToken invalida un token del caché
func (s *AuthService) InvalidateToken(ctx context.Context, token string) error {
	cacheKey := fmt.Sprintf("auth:token:%s", token)
	return s.redis.Del(ctx, cacheKey).Err()
}
