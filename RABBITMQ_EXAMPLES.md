# Ejemplos de mensajes para probar RabbitMQ con StockGo

## Configuración de RabbitMQ

**Exchanges:** Tipo `fanout` (cada evento tiene su propio exchange)
- `orders_placed` → Cola: `orders_placed_stock`
- `orders_confirmed` → Cola: `orders_confirmed_stock`
- `orders_canceled` → Cola: `orders_canceled_stock`

**Nota:** Los exchanges tipo **fanout** no usan routing keys, envían mensajes a todas las colas vinculadas.

## 1. Mensaje de Orden Creada (order_placed)

**Exchange:** `orders_placed` (fanout)  
**Cola:** `orders_placed_stock`  
**Routing Key:** (no requerido para fanout)

```json
{
  "orderId": "ORD-001",
  "cartId": "CART-123",
  "userId": "USER-456",
  "articles": [
    {
      "articleId": "ART-001",
      "quantity": 2
    },
    {
      "articleId": "ART-002", 
      "quantity": 1
    }
  ]
}
```

## 2. Mensaje de Orden Confirmada (order_confirmed)

**Exchange:** `orders_confirmed` (fanout)  
**Cola:** `orders_confirmed_stock`  
**Routing Key:** (no requerido para fanout)

```json
{
  "orderId": "ORD-001",
  "cartId": "CART-123", 
  "userId": "USER-456",
  "articles": [
    {
      "articleId": "ART-001",
      "quantity": 2
    },
    {
      "articleId": "ART-002",
      "quantity": 1
    }
  ],
  "confirmed_at": "2025-10-13T18:30:00Z"
}
```

## 3. Mensaje de Orden Cancelada (order_canceled)

**Exchange:** `orders_canceled` (fanout)  
**Cola:** `orders_canceled_stock`  
**Routing Key:** (no requerido para fanout)

```json
{
  "orderId": "ORD-001",
  "cartId": "CART-123",
  "userId": "USER-456", 
  "articles": [
    {
      "articleId": "ART-001",
      "quantity": 2
    },
    {
      "articleId": "ART-002",
      "quantity": 1
    }
  ],
  "canceled_at": "2025-10-13T18:35:00Z",
  "reason": "Payment failed"
}
```

## 4. Comandos para enviar mensajes usando RabbitMQ Management o CLI

### Usando rabbitmqadmin (recomendado para fanout)

```bash
# Publicar orden creada
rabbitmqadmin publish exchange=orders_placed payload='{"orderId":"ORD-001","cartId":"CART-123","userId":"USER-456","articles":[{"articleId":"ART-001","quantity":2},{"articleId":"ART-002","quantity":1}]}'

# Publicar orden confirmada  
rabbitmqadmin publish exchange=orders_confirmed payload='{"orderId":"ORD-001","cartId":"CART-123","userId":"USER-456","articles":[{"articleId":"ART-001","quantity":2},{"articleId":"ART-002","quantity":1}],"confirmed_at":"2025-10-13T18:30:00Z"}'

# Publicar orden cancelada
rabbitmqadmin publish exchange=orders_canceled payload='{"orderId":"ORD-001","cartId":"CART-123","userId":"USER-456","articles":[{"articleId":"ART-001","quantity":2},{"articleId":"ART-002","quantity":1}],"canceled_at":"2025-10-13T18:35:00Z","reason":"Payment failed"}'
```

### Usando RabbitMQ Management Web UI

1. Ir a http://localhost:15672 (usuario: guest, password: guest)
2. Ir a "Exchanges"
3. Hacer click en el exchange correspondiente (`orders_placed`, `orders_confirmed` o `orders_canceled`) 
4. En "Publish message":
   - **Routing key:** (dejar vacío, fanout no lo usa)
   - **Payload:** copiar el JSON del ejemplo
   - Hacer click en "Publish message"

## 5. Ejemplo de mensajes de Low Stock (enviados por el microservicio)

El microservicio enviará automáticamente estos mensajes cuando el stock esté bajo:

**Routing Key:** `article_lowstock`
**Exchange:** `ecommerce`

```json
{
  "article_id": "ART-001",
  "current_quantity": 5,
  "min_quantity": 10,
  "location": "Warehouse A"
}
```

## 6. APIs REST disponibles para crear datos de prueba

### Agregar artículos (POST /articles)
```bash
curl -X POST http://localhost:8080/articles \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "ART-001",
    "quantity": 100,
    "min_quantity": 10,
    "location": "Warehouse A"
  }'
```

### Reservar stock manualmente (POST /reserve)
```bash
curl -X POST http://localhost:8080/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "ART-001", 
    "quantity": 5,
    "order_id": "ORD-TEST-001"
  }'
```

### Ver stock actual (GET /stock/{article_id})
```bash
curl http://localhost:8080/stock/ART-001
```

### Ver stock bajo (GET /low-stock)
```bash
curl http://localhost:8080/low-stock
```

## 7. Flujo de prueba completo

1. **Preparación:**
   - Ejecutar PostgreSQL y RabbitMQ
   - Ejecutar el microservicio: `go run cmd/main.go`
   - Agregar algunos artículos via API REST

2. **Probar flujo de orden:**
   - Enviar mensaje `order_placed` → Debe reservar stock
   - Enviar mensaje `order_confirmed` → Debe confirmar y descontar stock  
   - O enviar `order_canceled` → Debe liberar las reservas

3. **Verificar:**
   - Chequear logs del microservicio
   - Verificar stock via API GET
   - Si el stock queda bajo el mínimo, debe enviarse mensaje `article_lowstock`

## 8. Logs esperados

Cuando funcione correctamente verás logs como:
```
OrderPlacedConsumer: Waiting for orders_placed messages...
OrderConfirmedConsumer: Waiting for orders_confirmed messages...
OrderCanceledConsumer: Waiting for orders_canceled messages...
OrderPlacedConsumer: Successfully reserved 2 units of article ART-001 for order ORD-001
LowStockPublisher: Publishing low stock alert for article ART-001
```