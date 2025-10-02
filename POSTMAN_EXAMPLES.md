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
```

### 4. Obtener Todos los Artículos
```http
GET http://localhost:8080/api/stock/articles
```

### 5. Obtener Eventos de un Artículo
```http
GET http://localhost:8080/api/stock/articles/ART-001/events
```

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

### 7. Reabastecer Stock (Por URL)
```http
PUT http://localhost:8080/api/stock/articles/ART-001/replenish
Content-Type: application/json

{
    "quantity": 25,
    "reason": "Restock urgente"
}
```

### 8. Deducir Stock (JSON)
```http
PUT http://localhost:8080/api/stock/deduct
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 15,
    "reason": "Venta directa"
}
```

### 9. Deducir Stock (Por URL)
```http
PUT http://localhost:8080/api/stock/articles/ART-001/deduct
Content-Type: application/json

{
    "quantity": 5,
    "reason": "Muestra de producto"
}
```

### 10. Reservar Stock (JSON)
```http
POST http://localhost:8080/api/stock/reserve
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 3,
    "order_id": "ORDER-123"
}
```

### 11. Reservar Stock (Por URL)
```http
POST http://localhost:8080/api/stock/articles/ART-001/reserve
Content-Type: application/json

{
    "quantity": 2,
    "order_id": "ORDER-456"
}
```

### 12. Cancelar Reserva (JSON)
```http
DELETE http://localhost:8080/api/stock/reservations
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 3,
    "order_id": "ORDER-123",
    "reason": "Cliente canceló la orden"
}
```

### 13. Cancelar Reserva (Por URL)
```http
DELETE http://localhost:8080/api/stock/orders/ORDER-456/reservations/ART-001
Content-Type: application/json

{
    "quantity": 2,
    "reason": "Timeout de reserva"
}
```

### 14. Confirmar Reserva (JSON)
```http
POST http://localhost:8080/api/stock/reservations/confirm
Content-Type: application/json

{
    "article_id": "ART-001",
    "quantity": 3,
    "order_id": "ORDER-789",
    "reason": "Pago confirmado"
}
```

### 15. Confirmar Reserva (Por URL)
```http
POST http://localhost:8080/api/stock/orders/ORDER-789/reservations/ART-001/confirm
Content-Type: application/json

{
    "quantity": 3,
    "reason": "Procesamiento de pago exitoso"
}
```

### 16. Consultar Artículos con Bajo Stock
```http
GET http://localhost:8080/api/stock/low-stock
```

### 17. Obtener Resumen de Alertas
```http
GET http://localhost:8080/api/stock/alerts/summary
```

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
POST /api/stock/reserve
{
    "article_id": "LAPTOP-001",
    "quantity": 2,
    "order_id": "ORDER-SUCCESS-001"
}

# 3. Verificar stock reservado
GET /api/stock/articles/LAPTOP-001

# 4. Confirmar reserva (pago exitoso)
POST /api/stock/reservations/confirm
{
    "article_id": "LAPTOP-001",
    "quantity": 2,
    "order_id": "ORDER-SUCCESS-001",
    "reason": "Pago confirmado"
}

# 5. Verificar stock final
GET /api/stock/articles/LAPTOP-001
```

### Scenario 2: Cancelación de Orden
```bash
# 1. Reservar stock
POST /api/stock/reserve
{
    "article_id": "LAPTOP-001",
    "quantity": 1,
    "order_id": "ORDER-CANCEL-001"
}

# 2. Cancelar reserva
DELETE /api/stock/reservations
{
    "article_id": "LAPTOP-001",
    "quantity": 1,
    "order_id": "ORDER-CANCEL-001",
    "reason": "Cliente canceló"
}

# 3. Verificar que el stock se liberó
GET /api/stock/articles/LAPTOP-001
```

### Scenario 3: Gestión de Inventario
```bash
# 1. Verificar stock actual
GET /api/stock/articles/LAPTOP-001

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
```

## 🔍 Casos de Error para Probar

### Error 400: Stock Insuficiente para Reserva
```http
POST http://localhost:8080/api/stock/reserve
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

### Respuesta Exitosa - Consultar Stock
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

## 🎯 Tips para Pruebas Efectivas

### 1. Usar Variables de Entorno en Postman
- `{{base_url}}` = `http://localhost:8080`
- `{{article_id}}` = `ART-001`
- `{{order_id}}` = `ORDER-123`

### 2. Verificar Consistencia de Stock
Después de cada operación, verificar:
- `quantity` ≥ `reserved` (siempre)
- `quantity - reserved` = stock disponible
- Los eventos se registran correctamente

### 3. Flujo de Prueba Recomendado
1. **Health Check** → Verificar servicio activo
2. **Crear Artículo** → Establecer inventario inicial
3. **Consultar Stock** → Verificar datos
4. **Reservar** → Simular orden de cliente
5. **Verificar Reserva** → Confirmar cambios
6. **Confirmar o Cancelar** → Completar flujo
7. **Verificar Estado Final** → Validar consistencia

### 4. Casos de Prueba por Cubrir
- ✅ Stock suficiente vs insuficiente
- ✅ Artículos existentes vs inexistentes
- ✅ Datos válidos vs inválidos
- ✅ Reservas exitosas vs fallidas
- ✅ Operaciones concurrentes
- ✅ Umbrales de stock mínimo/máximo

¡Listo para probar tu microservicio de stock paso a paso! 🚀