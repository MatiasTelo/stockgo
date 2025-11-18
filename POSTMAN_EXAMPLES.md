# StockGO Microservice - Ejemplos de Requests para Postman

## 🎯 Requests de Ejemplo para Postman

### 1. Health Check
```http
GET http://localhost:8080/health
```

### 2. Crear Artículo
```http
POST http://localhost:8080/api/stock/articles
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 100,
    "min_stock": 10,
    "max_stock": 500,
    "location": "A1-B2-C3"
}
```

### 3. Obtener Stock de Artículo
```http
GET http://localhost:8080/api/stock/articles/ART-001
Authorization: Bearer YOUR_TOKEN_HERE
```

### 4. Obtener Todos los Artículos
```http
GET http://localhost:8080/api/stock/articles
Authorization: Bearer YOUR_TOKEN_HERE
```

### 5. Obtener Eventos de un Artículo
```http
GET http://localhost:8080/api/stock/articles/ART-001/events
Authorization: Bearer YOUR_TOKEN_HERE
```

**Nota:** Todos los endpoints GET requieren autenticación mediante token Bearer en el header `Authorization`.

### 6. Reabastecer Stock (JSON)
```http
PUT http://localhost:8080/api/stock/replenish
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 50,
    "reason": "Reposición mensual"
}
```

### 7. Deducir Stock (JSON)
```http
PUT http://localhost:8080/api/stock/deduct
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 15,
    "reason": "Venta directa"
}
```

### 8. Reservar Stock (JSON)
```http
PUT http://localhost:8080/api/stock/reserve
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 3,
    "order_id": "ORDER-123"
}
```

### 9. Cancelar Reserva (JSON)
```http
PUT http://localhost:8080/api/stock/cancel-reservation
Content-Type: application/json

{
    "article_id": "ART-001",
    "order_id": "ORDER-123",
    "reason": "Cliente canceló la orden"
}
```

### 10. Confirmar Reserva (JSON)
```http
PUT http://localhost:8080/api/stock/confirm-reservation
Content-Type: application/json

{
    "article_id": "ART-001",
    "order_id": "ORDER-789",
    "reason": "Pago confirmado"
}
```

### 11. Consultar Artículos con Bajo Stock
```http
GET http://localhost:8080/api/stock/low-stock
```

## 🔐 Autenticación

Los siguientes endpoints requieren autenticación mediante token Bearer:
- `GET /api/stock/articles` - Obtener todos los artículos
- `GET /api/stock/articles/:articleId` - Obtener un artículo específico  
- `GET /api/stock/articles/:articleId/events` - Obtener eventos de un artículo

### Cómo usar autenticación en Postman:
1. En la pestaña "Authorization" de la request
2. Selecciona "Bearer Token" en el dropdown
3. Pega tu token en el campo "Token"

O directamente en Headers:
```
Authorization: Bearer YOUR_TOKEN_HERE
```

**Nota:** Los endpoints POST y PUT no requieren autenticación por ahora.

## 🧪 Scenarios de Prueba Completos

### Scenario 1: Flujo E-commerce Exitoso
```bash
# 1. Crear artículo
POST /api/stock/articles
{
    "article_id": "LAPTOP-001",
    "quantity": 50,
    "min_stock": 5,
    "max_stock": 200,
    "location": "Almacén A"
}

# 2. Cliente reserva stock
PUT /api/stock/reserve
{
    "article_id": "LAPTOP-001",
    "quantity": 2,
    "order_id": "ORDER-SUCCESS-001"
}

# 3. Verificar stock reservado
GET /api/stock/articles/LAPTOP-001
Authorization: Bearer YOUR_TOKEN_HERE

# 4. Confirmar reserva (pago exitoso)
PUT /api/stock/confirm-reservation
{
    "article_id": "LAPTOP-001",
    "order_id": "ORDER-SUCCESS-001",
    "reason": "Pago confirmado"
}

# 5. Verificar stock final
GET /api/stock/articles/LAPTOP-001
Authorization: Bearer YOUR_TOKEN_HERE
```

### Scenario 2: Cancelación de Orden
```bash
# 1. Reservar stock
PUT /api/stock/reserve
{
    "article_id": "LAPTOP-001",
    "quantity": 1,
    "order_id": "ORDER-CANCEL-001"
}

# 2. Cancelar reserva
PUT /api/stock/cancel-reservation
{
    "article_id": "LAPTOP-001",
    "order_id": "ORDER-CANCEL-001",
    "reason": "Cliente canceló"
}

# 3. Verificar que el stock se liberó
GET /api/stock/articles/LAPTOP-001
Authorization: Bearer YOUR_TOKEN_HERE
```

### Scenario 3: Gestión de Inventario
```bash
# 1. Verificar stock actual
GET /api/stock/articles/LAPTOP-001
Authorization: Bearer YOUR_TOKEN_HERE

# 2. Reabastecer inventario
PUT /api/stock/replenish
{
    "article_id": "LAPTOP-001",
    "quantity": 30,
    "reason": "Recepción de mercadería"
}

# 3. Verificar artículos con bajo stock
GET /api/stock/low-stock

# 4. Venta directa
PUT /api/stock/deduct
{
    "article_id": "LAPTOP-001",
    "quantity": 1,
    "reason": "Venta en tienda"
}

# 5. Ver eventos del artículo
GET /api/stock/articles/LAPTOP-001/events
Authorization: Bearer YOUR_TOKEN_HERE
```

## 🔍 Casos de Error para Probar

### Error 400: Stock Insuficiente para Reserva
```http
PUT http://localhost:8080/api/stock/reserve
Content-Type: application/json

{
    "article_id": "LAPTOP-001",
    "quantity": 999999,
    "order_id": "ORDER-FAIL-001"
}
```

### Error 404: Artículo No Existe
```http
GET http://localhost:8080/api/stock/articles/ARTICULO-INEXISTENTE
Authorization: Bearer YOUR_TOKEN_HERE
```

### Error 401: Sin Token de Autenticación
```http
GET http://localhost:8080/api/stock/articles/ART-001
```
**Respuesta esperada:**
```json
{
    "error": "Authorization header is required"
}
```

### Error 401: Token Inválido
```http
GET http://localhost:8080/api/stock/articles/ART-001
Authorization: Bearer INVALID_TOKEN
```
**Respuesta esperada:**
```json
{
    "error": "Invalid authorization header format. Expected: Bearer <token>"
}
```

### Error 400: Datos Inválidos
```http
POST http://localhost:8080/api/stock/articles
Content-Type: application/json

{
    "article_id": "",
    "quantity": -10
}
```

### Error 400: Deducir Más Stock del Disponible
```http
PUT http://localhost:8080/api/stock/deduct
Content-Type: application/json

{
    "article_id": "LAPTOP-001",
    "quantity": 999999,
    "reason": "Venta"
}
```

## 📊 Validación de Respuestas Esperadas

### Respuesta Exitosa - Health Check
```json
{
    "status": "healthy",
    "service": "stock-service",
    "version": "1.0.0",
    "timestamp": 1696250400
}
```

### Respuesta Exitosa - Crear Artículo
```json
{
    "message": "Article created successfully",
    "article_id": "ART-001",
    "initial_quantity": 100
}
```

### Respuesta Exitosa - Reservar Stock
```json
{
    "message": "Stock reserved successfully",
    "reserved": {
        "article_id": "ART-001",
        "quantity": 3,
        "order_id": "ORDER-123"
    },
    "stock": {
        "id": "uuid-123-456",
        "article_id": "ART-001",
        "quantity": 97,
        "reserved": 3,
        "min_stock": 10,
        "max_stock": 500,
        "location": "A1-B2-C3"
    }
}
```

### Respuesta Exitosa - Cancelar Reserva
```json
{
    "message": "Reservation cancelled successfully",
    "cancelled_reservation": {
        "article_id": "ART-001",
        "order_id": "ORDER-123"
    },
    "stock": {
        "id": "uuid-123-456",
        "article_id": "ART-001",
        "quantity": 97,
        "reserved": 0,
        "min_stock": 10,
        "max_stock": 500,
        "location": "A1-B2-C3"
    }
}
```

### Respuesta Exitosa - Confirmar Reserva
```json
{
    "message": "Reservation confirmed successfully",
    "confirmed_reservation": {
        "article_id": "ART-001",
        "order_id": "ORDER-789"
    },
    "stock": {
        "id": "uuid-123-456",
        "article_id": "ART-001",
        "quantity": 94,
        "reserved": 0,
        "min_stock": 10,
        "max_stock": 500,
        "location": "A1-B2-C3"
    }
}
```
```json
{
    "id": "uuid-123-456",
    "article_id": "ART-001",
    "quantity": 97,
    "reserved": 3,
    "min_stock": 10,
    "max_stock": 500,
    "location": "A1-B2-C3",
    "created_at": "2025-10-02T10:30:00Z",
    "updated_at": "2025-10-02T15:45:00Z"
}
```

### Respuesta Exitosa - Lista de Artículos
```json
[
    {
        "id": "uuid-123-456",
        "article_id": "ART-001",
        "quantity": 97,
        "reserved": 3,
        "min_stock": 10,
        "max_stock": 500,
        "location": "A1-B2-C3"
    },
    {
        "id": "uuid-789-012",
        "article_id": "ART-002",
        "quantity": 25,
        "reserved": 0,
        "min_stock": 5,
        "max_stock": 100,
        "location": "B1-C2-D3"
    }
]
```

### Respuesta Exitosa - Bajo Stock
```json
[
    {
        "article_id": "ART-003",
        "quantity": 8,
        "reserved": 2,
        "min_stock": 10,
        "available": 6,
        "shortage": 4
    }
]
```

### Respuesta Error - Stock Insuficiente
```json
{
    "error": "Insufficient stock available",
    "details": "Requested: 999999, Available: 94",
    "article_id": "LAPTOP-001",
    "available_quantity": 94,
    "requested_quantity": 999999
}
```

### Respuesta Error - Artículo No Encontrado
```json
{
    "error": "Article not found",
    "article_id": "ARTICULO-INEXISTENTE"
}
```