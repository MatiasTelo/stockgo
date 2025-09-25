# StockGO Microservice - Ejemplos de Requests

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
    "article_id": "LAPTOP-DELL-001",
    "name": "Laptop Dell Inspiron 15",
    "description": "Laptop Dell Inspiron 15 con 8GB RAM y 256GB SSD",
    "initial_stock": 50,
    "min_stock": 5,
    "max_stock": 200,
    "unit_price": 899.99,
    "location": "A1-B3-C2",
    "metadata": {
        "category": "electronics",
        "brand": "Dell",
        "model": "Inspiron 15",
        "warranty_months": 24,
        "supplier": "Dell Technologies",
        "color": "silver",
        "weight": "2.1kg"
    }
}
```

### 3. Obtener Stock de Artículo
```http
GET http://localhost:8080/api/stock/LAPTOP-DELL-001
```

### 4. Reabastecer Stock
```http
POST http://localhost:8080/api/stock/replenish
Content-Type: application/json

{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 25,
    "reason": "Restock weekly shipment",
    "supplier": "Dell Technologies",
    "batch_number": "DELL-2025-W39",
    "expiration_date": "2027-09-25T00:00:00Z",
    "unit_cost": 750.00,
    "metadata": {
        "purchase_order": "PO-2025-1234",
        "shipment_id": "SHIP-DELL-456",
        "quality_check": "passed",
        "receiving_employee": "Juan Perez"
    }
}
```

### 5. Reservar Stock
```http
POST http://localhost:8080/api/stock/reserve
Content-Type: application/json

{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 3,
    "order_id": "ORD-2025-001234",
    "customer_id": "CUST-789456",
    "expiration_minutes": 45,
    "metadata": {
        "sales_channel": "ecommerce_web",
        "customer_tier": "premium",
        "promotion_code": "BACK2SCHOOL",
        "sales_rep": "Maria Rodriguez",
        "priority": "high"
    }
}
```

### 6. Deducir Stock (Venta Directa)
```http
POST http://localhost:8080/api/stock/deduct
Content-Type: application/json

{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 2,
    "reason": "Walk-in customer purchase",
    "transaction_id": "TXN-POS-2025-9876",
    "customer_id": "CUST-WALK-001",
    "metadata": {
        "sales_channel": "physical_store",
        "store_location": "Store Downtown",
        "cashier_id": "EMP-001",
        "payment_method": "credit_card",
        "receipt_number": "RCP-001234"
    }
}
```

### 7. Cancelar Reserva
```http
POST http://localhost:8080/api/stock/cancel-reservation
Content-Type: application/json

{
    "order_id": "ORD-2025-001234",
    "article_id": "LAPTOP-DELL-001",
    "quantity": 3,
    "reason": "Customer requested cancellation",
    "metadata": {
        "cancelled_by": "customer",
        "cancellation_reason": "found_better_deal",
        "refund_amount": 2699.97,
        "processed_by": "customer_service"
    }
}
```

### 8. Consultar Artículos con Bajo Stock
```http
GET http://localhost:8080/api/stock/low-stock?threshold=20
```

## 🧪 Scenarios de Prueba Completos

### Scenario 1: Flujo E-commerce Exitoso
```bash
# 1. Cliente agrega productos al carrito
POST /api/stock/reserve
{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 2,
    "order_id": "ORD-SUCCESS-001",
    "customer_id": "CUST-001"
}

# 2. Cliente procede al checkout (stock reservado)
GET /api/stock/LAPTOP-DELL-001
# Verificar: reserved_stock = 2

# 3. Pago exitoso - confirmar reserva
# (Esto se haría via RabbitMQ en producción)
# Por ahora simularemos deduciendo el stock reservado
POST /api/stock/deduct
{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 2,
    "reason": "Order confirmed and paid",
    "transaction_id": "TXN-SUCCESS-001"
}
```

### Scenario 2: Cancelación de Orden
```bash
# 1. Cliente reserva stock
POST /api/stock/reserve
{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 1,
    "order_id": "ORD-CANCEL-001",
    "customer_id": "CUST-002"
}

# 2. Cliente cancela la orden
POST /api/stock/cancel-reservation
{
    "order_id": "ORD-CANCEL-001",
    "article_id": "LAPTOP-DELL-001",
    "quantity": 1,
    "reason": "Customer cancellation"
}

# 3. Verificar que el stock se liberó
GET /api/stock/LAPTOP-DELL-001
# Verificar: reserved_stock disminuyó en 1
```

### Scenario 3: Gestión de Inventario
```bash
# 1. Verificar stock actual
GET /api/stock/LAPTOP-DELL-001

# 2. Reabastecer cuando llega mercadería
POST /api/stock/replenish
{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 30,
    "reason": "Monthly restock"
}

# 3. Verificar artículos que necesitan reabastecimiento
GET /api/stock/low-stock?threshold=15

# 4. Venta directa en tienda física
POST /api/stock/deduct
{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 1,
    "reason": "Walk-in sale",
    "transaction_id": "TXN-STORE-001"
}
```

## 🔍 Casos de Error para Probar

### Error 400: Stock Insuficiente
```http
POST http://localhost:8080/api/stock/reserve
Content-Type: application/json

{
    "article_id": "LAPTOP-DELL-001",
    "quantity": 999999,
    "order_id": "ORD-FAIL-001"
}
```

### Error 404: Artículo No Existe
```http
GET http://localhost:8080/api/stock/ARTICULO-INEXISTENTE
```

### Error 400: Datos Inválidos
```http
POST http://localhost:8080/api/stock/articles
Content-Type: application/json

{
    "article_id": "",
    "name": "",
    "initial_stock": -10
}
```

## 📊 Validación de Respuestas

### Respuesta Exitosa - Crear Artículo
```json
{
    "message": "Article created successfully",
    "article_id": "LAPTOP-DELL-001",
    "initial_stock": 50
}
```

### Respuesta Exitosa - Consultar Stock
```json
{
    "article_id": "LAPTOP-DELL-001",
    "name": "Laptop Dell Inspiron 15",
    "current_stock": 45,
    "available_stock": 42,
    "reserved_stock": 3,
    "min_stock": 5,
    "max_stock": 200,
    "unit_price": 899.99,
    "location": "A1-B3-C2",
    "metadata": {
        "category": "electronics",
        "brand": "Dell"
    },
    "last_updated": "2025-09-25T15:30:00Z"
}
```

### Respuesta Error - Stock Insuficiente
```json
{
    "error": "Insufficient available stock. Requested: 999999, Available: 42",
    "code": "INSUFFICIENT_STOCK",
    "available_stock": 42,
    "requested_quantity": 999999
}
```

## 🎯 Tips para Pruebas Efectivas

### 1. Usar Variables de Entorno
- Configura `{{base_url}}` = `http://localhost:8080`
- Usa `{{article_id}}` para reutilizar IDs
- Usa `{{order_id}}` para rastrear órdenes

### 2. Verificar Consistencia
Después de cada operación, consulta el stock para verificar:
- `current_stock` = `available_stock` + `reserved_stock`
- Los cambios se reflejan correctamente

### 3. Pruebas de Concurrencia
- Ejecuta múltiples reservas del mismo artículo
- Verifica que no hay over-reservation

### 4. Monitorear Logs
Observa los logs del servidor para:
- Confirmar que las operaciones se ejecutan
- Detectar errores o warnings
- Validar el flujo de eventos

¡Listo para probar tu microservicio de stock! 🚀