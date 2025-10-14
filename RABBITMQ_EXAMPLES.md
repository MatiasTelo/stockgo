# Ejemplos de mensajes para probar RabbitMQ con StockGo

## Configuración de RabbitMQ

**Exchange:** `ecommerce`  
**Tipo:** `topic`  
**Colas y Routing Keys:**
- `stock.order.placed` → routing key: `order_placed`
- `stock.order.confirmed` → routing key: `order_confirmed`  
- `stock.order.canceled` → routing key: `order_canceled`

## 1. Mensaje de Orden Creada (order_placed)

**Routing Key:** `order_placed`  
**Cola:** `stock.order.placed`

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

**Routing Key:** `order_confirmed`  
**Cola:** `stock.order.confirmed`

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

**Routing Key:** `order_canceled`  
**Cola:** `stock.order.canceled`

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

### Usando rabbitmqctl (desde línea de comandos)

```bash
# Publicar orden creada
rabbitmqctl publish_message ecommerce order_placed '{"orderId":"ORD-001","cartId":"CART-123","userId":"USER-456","articles":[{"articleId":"ART-001","quantity":2},{"articleId":"ART-002","quantity":1}]}'

# Publicar orden confirmada  
rabbitmqctl publish_message ecommerce order_confirmed '{"orderId":"ORD-001","cartId":"CART-123","userId":"USER-456","articles":[{"articleId":"ART-001","quantity":2},{"articleId":"ART-002","quantity":1}],"confirmed_at":"2025-10-13T18:30:00Z"}'

# Publicar orden cancelada
rabbitmqctl publish_message ecommerce order_canceled '{"orderId":"ORD-001","cartId":"CART-123","userId":"USER-456","articles":[{"articleId":"ART-001","quantity":2},{"articleId":"ART-002","quantity":1}],"canceled_at":"2025-10-13T18:35:00Z","reason":"Payment failed"}'
```

### Usando RabbitMQ Management Web UI

1. Ir a http://localhost:15672 (usuario: guest, password: guest)
2. Ir a "Exchanges"
3. Hacer click en "ecommerce" 
4. En "Publish message":
   - **Routing key:** usar `order_placed`, `order_confirmed` o `order_canceled`
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
OrderPlacedConsumer: Starting to consume order.placed messages
OrderConfirmedConsumer: Starting to consume order.confirmed messages  
OrderCanceledConsumer: Starting to consume order.canceled messages
OrderPlacedConsumer: Successfully reserved 2 units of article ART-001 for order ORD-001
LowStockPublisher: Publishing low stock alert for article ART-001
```