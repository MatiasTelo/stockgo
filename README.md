# Microservicio de Stock - StockGO

![Go](https://img.shields.io/badge/Go-1.25+-blue.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-blue.svg)
![Fiber](https://img.shields.io/badge/Fiber-v2-green.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

Microservicio especializado en la gestión de inventario y stock desarrollado en Go con arquitectura limpia, diseñado para sistemas de e-commerce de alta disponibilidad.

## 🎯 Casos de Uso

### CU: Validación de stock de un artículo

**Precondición**: El sistema recibe una solicitud de reserva o venta de un artículo

**Camino normal**:
1. Buscar el Stock para el articleId solicitado
2. Comparar el stock disponible (currentStock - reserved) con la cantidad solicitada
3. Si hay suficiente stock disponible, proceder con la operación
4. Registrar el evento correspondiente (RESERVE, DEDUCT, etc.)
5. Actualizar los campos currentStock y/o reserved según corresponda

**Caminos alternativos**:
- Si no hay suficiente stock, retornar error con código 400
- Si el artículo no existe, retornar error 404

### CU: Reserva de stock para órdenes

**Precondición**: Se recibe una solicitud de reserva con order_id único

**Camino normal**:
1. Validar que no exista una reserva activa para el mismo order_id
2. Verificar stock disponible para el articleId
3. Incrementar el campo "reserved" del stock
4. Crear evento de tipo RESERVE en el historial
5. Retornar confirmación de reserva exitosa

**Caminos alternativos**:
- Si ya existe reserva activa para el order_id, retornar error 409 (Conflict)
- Si no hay stock suficiente, retornar error 400

### CU: Confirmación de reserva (conversión a venta)

**Precondición**: Existe una reserva activa para el order_id especificado

**Camino normal**:
1. Verificar que existe reserva activa para el order_id
2. Decrementar currentStock según cantidad reservada
3. Decrementar reserved según cantidad reservada
4. Crear evento CONFIRM_RESERVE en el historial
5. Marcar la reserva como confirmada

**Caminos alternativos**:
- Si no existe reserva activa, retornar error 404

### CU: Cancelación de reserva

**Precondición**: Existe una reserva activa para el order_id especificado

**Camino normal**:
1. Verificar que existe reserva activa para el order_id
2. Decrementar reserved según cantidad reservada
3. Crear evento CANCEL_RESERVE en el historial
4. Liberar el stock reservado para nuevas operaciones

### CU: Reabastecimiento de stock

**Precondición**: Se necesita incrementar el stock de un artículo

**Camino normal**:
1. Verificar que el artículo existe en el sistema
2. Incrementar currentStock según cantidad especificada
3. Crear evento REPLENISH en el historial
4. Verificar si el stock supera max_stock y generar alerta si corresponde

### CU: Consulta de stock de un artículo

**Precondición**: Se solicita información de stock para un articleId

**Camino normal**:
1. Buscar el registro de stock para el articleId
2. Calcular stock disponible (currentStock - reserved)
3. Retornar información completa del stock

**Caminos alternativos**:
- Si el artículo no existe, retornar error 404

### CU: Detección de stock bajo

**Precondición**: El sistema monitorea niveles de stock automáticamente

**Camino normal**:
1. Después de cada operación que reduzca stock, verificar si currentStock <= min_stock
2. Si se detecta stock bajo, crear evento LOW_STOCK
3. Incluir el artículo en la lista de alertas de stock bajo

## 📊 Modelo de Datos

### Stock
- **id**: UUID - Identificador único del registro
- **article_id**: VARCHAR(100) - ID del artículo (referencia externa)
- **quantity**: INTEGER - Stock total disponible (currentStock)
- **reserved**: INTEGER - Cantidad reservada pero no vendida
- **min_stock**: INTEGER - Nivel mínimo para alertas
- **max_stock**: INTEGER - Nivel máximo recomendado
- **location**: VARCHAR(255) - Ubicación física en almacén
- **created_at**: TIMESTAMP - Fecha de creación
- **updated_at**: TIMESTAMP - Última actualización

### StockEvent (MovStock)
- **id**: UUID - Identificador único del evento
- **article_id**: VARCHAR(100) - Artículo relacionado
- **event_type**: VARCHAR(50) - Tipo de movimiento [ADD|REPLENISH|DEDUCT|RESERVE|CANCEL_RESERVE|CONFIRM_RESERVE|LOW_STOCK]
- **quantity**: INTEGER - Cantidad del movimiento
- **order_id**: VARCHAR(100) - ID de orden (para reservas)
- **reason**: TEXT - Descripción o motivo del movimiento
- **metadata**: JSONB - Información adicional en formato JSON
- **created_at**: TIMESTAMP - Fecha y hora del evento

## 🚀 Interfaz REST

### Consulta de stock de un artículo

`GET /api/articles/{articleId}`

**Params path**
- articleId: ID del artículo para consultar

**Headers**
- Content-Type: application/json

**Response**

`200 OK` - Si existe el artículo
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "article_id": "LAPTOP-001",
  "quantity": 45,
  "reserved": 5,
  "available": 40,
  "min_stock": 10,
  "max_stock": 100,
  "location": "A1-B2-C3",
  "created_at": "2025-10-06T15:30:00Z",
  "updated_at": "2025-10-06T18:00:00Z"
}
```

`404 NOT FOUND` - Si no existe el artículo

### Crear artículo en inventario

`POST /api/articles`

**Body**
```json
{
  "article_id": "LAPTOP-001",
  "quantity": 50,
  "min_stock": 10,
  "max_stock": 100,
  "location": "A1-B2-C3"
}
```

**Response**
`201 CREATED` - Artículo creado exitosamente

### Reservar stock para una orden

`PUT /api/stock/reserve`

**Body**
```json
{
  "article_id": "LAPTOP-001",
  "quantity": 2,
  "order_id": "ORDER-12345",
  "reason": "Reserva para orden de compra"
}
```

**Response**
`200 OK` - Reserva exitosa
`400 BAD REQUEST` - Stock insuficiente
`409 CONFLICT` - Ya existe reserva para este order_id

### Cancelar reserva

`PUT /api/stock/cancel-reservation`

**Body**
```json
{
  "article_id": "LAPTOP-001",
  "order_id": "ORDER-12345",
  "reason": "Cliente canceló la orden"
}
```

### Confirmar reserva (conversión a venta)

`PUT /api/stock/confirm-reservation`

**Body**
```json
{
  "article_id": "LAPTOP-001", 
  "order_id": "ORDER-12345",
  "reason": "Pago confirmado"
}
```

### Reabastecer stock

`PUT /api/stock/replenish`

**Body**
```json
{
  "article_id": "LAPTOP-001",
  "quantity": 25,
  "reason": "Llegada de nuevo inventario"
}
```

### Deducir stock

`PUT /api/stock/deduct`

**Body**
```json
{
  "article_id": "LAPTOP-001",
  "quantity": 1,
  "reason": "Venta directa en tienda física"
}
```

### Consultar stock bajo

`GET /api/stock/low`

**Response**
```json
{
  "articles": [
    {
      "article_id": "LAPTOP-001",
      "current_stock": 8,
      "min_stock": 10,
      "deficit": 2
    }
  ],
  "total_articles": 1
}
```

### Consultar eventos de stock

`GET /api/stock/events` - Todos los eventos
`GET /api/stock/events/article/{articleId}` - Por artículo
`GET /api/stock/events/order/{orderId}` - Por orden

## 🏗️ Arquitectura

### Stack Tecnológico
- **Backend**: Go 1.25+ con Fiber Framework v2
- **Base de Datos**: PostgreSQL 13+
- **Migraciones**: golang-migrate/migrate
- **Cache**: Redis
- **Mensajería**: RabbitMQ


## 🚀 Instalación y Configuración

### Prerequisitos
- Go 1.25+
- PostgreSQL 13+
- Git

### 1. Clonar repositorio
```bash
git clone https://github.com/MatiasTelo/stockgo.git
cd stockgo
```

### 2. Configurar variables de entorno
```env
# Servidor
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=admin
DB_DATABASE=stockdb
DB_SSLMODE=disable

# Opcionales
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

### 3. Instalar dependencias
```bash
go mod download
```

### 4. Ejecutar migraciones
```bash
go run cmd/migrate.go up
```

### 5. Iniciar servicio
```bash
go run cmd/main.go
```

Servicio disponible en: `http://localhost:8080`

## 🐰 Interfaz Asíncrona (RabbitMQ)

### Exchange Principal
**Exchange**: `ecommerce` (tipo: topic)

---

### 📥 Consumers (Mensajes Recibidos)

#### 1. Procesamiento de Orden Creada
- **Consumer**: OrderPlacedConsumer
- **Queue**: `stock.order.placed`
- **Routing Key**: `order_placed`

**Body del mensaje**:
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

#### 2. Procesamiento de Orden Confirmada
- **Consumer**: OrderConfirmedConsumer
- **Queue**: `stock.order.confirmed`
- **Routing Key**: `order_confirmed`

**Body del mensaje**:
```json
{
  "orderId": "ORD-001",
  "cartId": "CART-123",
  "userId": "USER-456",
  "articles": [
    {
      "articleId": "ART-001",
      "quantity": 2
    }
  ],
  "confirmed_at": "2025-10-14T15:30:00Z"
}
```

#### 3. Procesamiento de Orden Cancelada
- **Consumer**: OrderCanceledConsumer
- **Queue**: `stock.order.canceled`
- **Routing Key**: `order_canceled`

**Body del mensaje**:
```json
{
  "orderId": "ORD-001",
  "cartId": "CART-123",
  "userId": "USER-456",
  "articles": [
    {
      "articleId": "ART-001",
      "quantity": 2
    }
  ],
  "canceled_at": "2025-10-14T15:35:00Z",
  "reason": "Payment failed"
}
```

---

### 📤 Publishers (Mensajes Enviados)

#### 1. Alerta de Stock Insuficiente
- **Publisher**: InsufficientStockPublisher
- **Routing Key**: `insufficient_stock`

**Body del mensaje**:
```json
{
  "order_id": "ORD-001",
  "article_ids": ["ART-001", "ART-003"]
}
```

#### 2. Alerta de Stock Bajo
- **Publisher**: LowStockPublisher
- **Routing Key**: `article_lowstock`

**Body del mensaje**:
```json
{
  "article_id": "ART-001",
  "current_quantity": 5,
  "min_quantity": 10,
  "location": "Warehouse A"
}
```



### Tipos de Eventos
- `ADD` - Creación de artículo
- `REPLENISH` - Reabastecimiento
- `DEDUCT` - Deducción directa (venta)
- `RESERVE` - Reserva de stock
- `CANCEL_RESERVE` - Cancelación de reserva
- `LOW_STOCK` - Alerta de stock bajo
