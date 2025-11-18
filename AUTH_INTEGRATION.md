# Integración con Microservicio de Autenticación

## 📋 Descripción

El servicio de Stock ahora se integra con el microservicio de autenticación (AuthGO) para validar tokens en los endpoints GET. Utiliza **Redis como caché** para mejorar el rendimiento y reducir las llamadas al servicio de autenticación.

## 🔐 Endpoints Protegidos

Los siguientes endpoints requieren autenticación mediante token Bearer:

- `GET /api/stock/articles` - Obtener todos los artículos
- `GET /api/stock/articles/:articleId` - Obtener un artículo específico
- `GET /api/stock/articles/:articleId/events` - Obtener eventos de un artículo

## ⚙️ Configuración

### Variables de Entorno

Agrega la siguiente variable en tu archivo `.env`:

```bash
AUTH_SERVICE_URL=http://localhost:3000
```

**Valores por defecto:**
- Si no se especifica, usa: `http://localhost:3000`
- Ajusta el puerto según tu configuración del servicio de autenticación

### Ejemplo de archivo .env completo

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=admin
DB_PASSWORD=admin
DB_DATABASE=stockdb
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=ecommerce
RABBITMQ_QUEUE=stock_events

# Auth Service Configuration
# Auth Service Configuration
AUTH_SERVICE_URL=http://localhost:3000
```

## 🧪 Cómo Probar

### 1. Asegúrate de que Redis esté corriendo
```bash
redis-server
# O si usas Docker:
docker run -d -p 6379:6379 redis:latest
```

### 2. Asegúrate de que el servicio de Auth esté corriendo
```

## 🔄 Flujo de Autenticación

1. **Cliente** envía request con token en el header:
   ```
   Authorization: Bearer <token>
   ```

2. **StockGO Middleware** extrae el token y busca en **Redis Cache**:
   - Si encuentra el token en caché → retorna los datos del usuario inmediatamente
   - Si NO encuentra el token → continúa al paso 3

3. **StockGO** llama al servicio de autenticación:
   ```
   GET http://localhost:3000/users/current
   Authorization: Bearer <token>
   ```

4. **AuthGO** valida el token y responde con los datos del usuario:
   ```json
   {
       "id": "user-123",
       "username": "john_doe",
       "email": "john@example.com",
       "role": "admin"
   }
   ```

5. **StockGO** guarda los datos en Redis (TTL: 10 minutos) y continúa con la petición

6. Si el token es inválido, responde con error 401

## 📦 Caché con Redis

### Ventajas
- ✅ Reduce latencia en validaciones repetidas
- ✅ Disminuye carga en el servicio de autenticación
- ✅ TTL de 10 minutos (configurable)
- ✅ Invalidación manual disponible

### Formato de Caché
```
Key: auth:token:<token>
Value: {"id":"user-123","username":"john_doe","email":"john@example.com","role":"admin"}
TTL: 600 segundos (10 minutos)
```

## 📝 Respuestas de Error

### Error 401: Sin Header de Autorización o Token Inválido
```json
{
    "error": "Unauthorized"
}
```

**Posibles causas:**
- No se envió el header `Authorization`
- El formato del header es incorrecto (debe ser `Bearer <token>`)
- El token está vacío
- El token es inválido o expiró
- El servicio de autenticación rechazó el token

## 🧪 Cómo Probar

### 2. Asegúrate de que el servicio de Auth esté corriendo
```bash
# En la carpeta de AuthGO
npm start  # o el comando que uses para iniciar tu servicio Node.js
```

### 3. Ejecuta el servicio de Stock
```bash
# En la carpeta de StockGO
go run cmd/main.go
```

### 4. Obtén un token válido del servicio de Auth
```bash
POST http://localhost:3000/auth/login
Content-Type: application/json

{
    "username": "admin",
    "password": "admin123"
}
```

### 5. Usa el token en las peticiones a Stock
```bash
GET http://localhost:8080/api/stock/articles
Authorization: Bearer <tu_token_aquí>
```

## 🔍 Acceso a la Información del Usuario en los Handlers

Si necesitas acceder a la información del usuario autenticado en tus handlers:

```go
func (h *YourHandler) Handle(c *fiber.Ctx) error {
    // Obtener el user_id del contexto
    userID := c.Locals("user_id").(string)
    
    // Obtener el username
    username := c.Locals("username").(string)
    
    // Obtener el objeto user completo
    user := c.Locals("user").(*service.UserResponse)
    
    // Usar la información en tu lógica
    log.Printf("Request from user: %s (ID: %s)", username, userID)
    log.Printf("User role: %s", user.Role)
    
    // ... resto del código
}
```

### Datos Disponibles en el Contexto
- `token` (string) - El token JWT completo
- `user_id` (string) - ID del usuario
- `username` (string) - Nombre de usuario
- `user` (*service.UserResponse) - Objeto completo con: ID, Username, Email, Role

## 🚀 Endpoints que NO Requieren Autenticación

Los siguientes endpoints siguen siendo públicos:

- `POST /api/stock/articles` - Crear artículo
- `PUT /api/stock/replenish` - Reabastecer stock
- `PUT /api/stock/deduct` - Deducir stock
- `PUT /api/stock/reserve` - Reservar stock
- `PUT /api/stock/cancel-reservation` - Cancelar reserva
- `PUT /api/stock/confirm-reservation` - Confirmar reserva
- `GET /api/stock/low-stock` - Consultar bajo stock
- `GET /health` - Health check

## 🛠️ Configuración del Servicio de Auth

El middleware espera que el servicio de autenticación tenga el siguiente endpoint:

**Endpoint:** `GET /users/current`

**Headers requeridos:**
```
Authorization: Bearer <token>
```

**Respuesta exitosa (200):**
```json
{
    "id": "user-123",
    "username": "john_doe",
    "email": "john@example.com",
    "role": "admin"
}
```

**Respuesta de error (401):**
Cualquier status code diferente a 200 será tratado como un token inválido.

## ⏱️ Timeout y Caché

- **Timeout de llamada al servicio:** 5 segundos
- **TTL del caché en Redis:** 10 minutos (600 segundos)
- **Verificación del caché:** Se hace antes de cada llamada al servicio de auth

## 🔄 Invalidación Manual del Caché

Si necesitas invalidar manualmente un token del caché (por ejemplo, al hacer logout):

```go
// En tu handler de logout o donde necesites
authService.InvalidateToken(context.Background(), token)
```

## 🧪 Verificar el Caché en Redis

Para verificar que los tokens se están guardando correctamente:

```bash
# Conectar a Redis CLI
redis-cli

# Ver todas las claves de auth
KEYS auth:token:*

# Ver un token específico
GET auth:token:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Ver el TTL de un token
TTL auth:token:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```
