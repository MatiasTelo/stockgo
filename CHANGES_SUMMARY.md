# Resumen de Cambios - Integración de Autenticación con Caché Redis

## ✅ Archivos Creados

### 1. `internal/service/auth_service.go`
Nuevo servicio que maneja la autenticación con caché en Redis:
- **`ValidateToken()`**: Valida tokens primero buscando en caché, luego en el servicio de auth
- **`callAuthService()`**: Llama a `GET /users/current` del microservicio de auth
- **`InvalidateToken()`**: Permite invalidar manualmente un token del caché
- **Caché**: TTL de 10 minutos en Redis con formato `auth:token:<token>`

### 2. `AUTH_INTEGRATION.md`
Documentación completa de la integración con instrucciones de uso

## 🔄 Archivos Modificados

### 1. `internal/middleware/auth.go`
- Simplificado para usar el `AuthService`
- Extrae token con función `extractToken()`
- Almacena en contexto: `token`, `user_id`, `username`, `user`
- Respuesta unificada de error: `"Unauthorized"`

### 2. `internal/config/config.go`
- Agregado `AuthConfig` struct con `ServiceURL`
- Variable de entorno: `AUTH_SERVICE_URL` (default: `http://localhost:3000`)

### 3. `cmd/main.go`
- Creado `authService` con Redis y URL del servicio de auth
- Middleware ahora recibe `authService` en lugar de URL
- Rutas GET protegidas usan `middleware.AuthMiddleware(authService)`

### 4. `.env.example`
- Agregada variable `AUTH_SERVICE_URL=http://localhost:3000`

## 🎯 Características Principales

### Caché en Redis
```
✅ Reduce latencia en validaciones repetidas
✅ TTL configurable (10 minutos por defecto)
✅ Formato: auth:token:<token>
✅ Invalidación manual disponible
✅ Timeout de 5 segundos en llamadas al servicio de auth
```

### Flujo de Autenticación
1. Cliente envía token Bearer
2. Middleware busca en caché Redis
   - **Si existe**: Retorna datos inmediatamente
   - **Si NO existe**: Llama al servicio de auth
3. Servicio de auth valida con `GET /users/current`
4. Respuesta se guarda en Redis y se almacena en contexto
5. Handler accede a datos del usuario desde contexto

### Datos Almacenados en Contexto
```go
c.Locals("token")      // string - Token JWT completo
c.Locals("user_id")    // string - ID del usuario
c.Locals("username")   // string - Nombre de usuario
c.Locals("user")       // *service.UserResponse - Objeto completo
```

## 🔧 Estructura UserResponse
```go
type UserResponse struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role,omitempty"`
}
```

## 📋 Endpoints del Servicio de Auth Requeridos

### GET /users/current
**Headers:**
```
Authorization: Bearer <token>
```

**Respuesta (200):**
```json
{
    "id": "user-123",
    "username": "john_doe",
    "email": "john@example.com",
    "role": "admin"
}
```

## 🧪 Testing

### Ver tokens en Redis
```bash
redis-cli
KEYS auth:token:*
GET auth:token:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
TTL auth:token:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Probar endpoint protegido
```bash
GET http://localhost:8080/api/stock/articles
Authorization: Bearer <your_token>
```

## 🚀 Para Iniciar

1. Asegúrate que Redis esté corriendo:
   ```bash
   redis-server
   ```

2. Configura la variable de entorno (opcional si usas el default):
   ```bash
   AUTH_SERVICE_URL=http://localhost:3000
   ```

3. Inicia el servicio:
   ```bash
   go run cmd/main.go
   ```

## 🔍 Diferencias con la Implementación de Node.js

| Node.js | Go (StockGO) |
|---------|--------------|
| `node-cache` | Redis |
| TTL: 600s | TTL: 600s (10 min) |
| Caché en memoria | Caché distribuido |
| `axios` | `fasthttp` |
| Timeout: 5s | Timeout: 5s |
| Endpoint: `/users/current` | Endpoint: `/users/current` |

## ✨ Ventajas de la Implementación

1. **Caché distribuido**: Redis permite escalar horizontalmente
2. **Performance**: `fasthttp` es más rápido que `net/http`
3. **Simplicidad**: Middleware limpio y fácil de mantener
4. **Flexibilidad**: Fácil cambiar TTL y configuraciones
5. **Monitoring**: Puedes inspeccionar el caché en Redis directamente
