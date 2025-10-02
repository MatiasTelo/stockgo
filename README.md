# Stock Microservice

Microservicio de gestión de stock para un sistema de ecommerce implementado en Go con PostgreSQL, Redis y RabbitMQ.

## Características

### Casos de Uso
1. **Agregar nuevo artículo** - Crear un nuevo artículo en el inventario
2. **Reponer stock** - Aumentar la cantidad de stock existente
3. **Descontar stock** - Reducir stock manualmente
4. **Reservar stock** - Reservar stock para órdenes pendientes
5. **Cancelar reserva de stock** - Liberar stock reservado

### Entidades Principales
1. **Stock** - Información principal del inventario por artículo
2. **StockEvent** - Historial de eventos/movimientos de stock
3. **StockReservation** - Reservas activas de stock

### Tecnologías
- **Lenguaje**: Go 1.25+
- **Base de datos**: PostgreSQL
- **Cache**: Redis
- **Mensajería**: RabbitMQ
- **API**: REST con Fiber framework

## Estructura del Proyecto

```
stockgo/
├── cmd/
│   ├── main.go              # Aplicación principal
│   └── migrate.go           # Script de migraciones
├── internal/
│   ├── config/              # Configuración
│   ├── database/            # Conexiones a BD
│   ├── handlers/            # Handlers REST (separados)
│   │   ├── add_article.go
│   │   ├── replenish_stock.go
│   │   ├── deduct_stock.go
│   │   ├── reserve_stock.go
│   │   ├── cancel_reservation.go
│   │   └── low_stock.go
│   ├── messaging/           # RabbitMQ
│   ├── models/              # Modelos de datos
│   ├── repository/          # Acceso a datos
│   └── service/             # Lógica de negocio
├── migrations/              # Migraciones SQL
├── .env.example            # Variables de entorno
├── Dockerfile
├── go.mod
└── README.md
```

## Configuración

1. **Copiar archivo de configuración**:
```bash
cp .env.example .env
```

2. **Configurar variables de entorno** en `.env`:
```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_DATABASE=stockdb

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=ecommerce
```

## Instalación y Ejecución

1. **Instalar dependencias**:
```bash
go mod tidy
```

2. **Ejecutar migraciones**:
```bash
go run cmd/migrate.go up
```

3. **Iniciar el servicio**:
```bash
go run cmd/main.go
```

## API Endpoints

### Gestión de Artículos
- `POST /api/stock/articles` - Crear nuevo artículo
- `GET /api/stock/articles` - Listar todos los artículos
- `GET /api/stock/articles/:articleId` - Obtener artículo específico
- `GET /api/stock/articles/:articleId/events` - Historial de eventos

### Operaciones de Stock
- `PUT /api/stock/replenish` - Reponer stock
- `PUT /api/stock/articles/:articleId/replenish` - Reponer stock por ID
- `PUT /api/stock/deduct` - Descontar stock
- `PUT /api/stock/articles/:articleId/deduct` - Descontar stock por ID

### Reservas
- `POST /api/stock/reserve` - Reservar stock
- `POST /api/stock/articles/:articleId/reserve` - Reservar por ID
- `DELETE /api/stock/reservations` - Cancelar reserva
- `DELETE /api/stock/orders/:orderId/reservations/:articleId` - Cancelar por IDs
- `POST /api/stock/reservations/confirm` - Confirmar reserva
- `POST /api/stock/orders/:orderId/reservations/:articleId/confirm` - Confirmar por IDs

### Alertas
- `GET /api/stock/low-stock` - Artículos con stock bajo
- `GET /api/stock/alerts/summary` - Resumen de alertas

### Health Check
- `GET /health` - Estado del servicio

## Integración con RabbitMQ

### Mensajes que Escucha
- `order.created` - Reserva stock automáticamente para nuevas órdenes
- `order.status.changed` - Confirma o cancela reservas según estado de la orden

### Mensajes que Publica
- `stock.alert.low` - Alerta cuando un artículo llega al stock mínimo

### Formato de Mensajes

**Order Created**:
```json
{
  "order_id": "ORD-12345",
  "items": [
    {
      "article_id": "ART-001",
      "quantity": 2,
      "price": 29.99
    }
  ],
  "status": "CREATED",
  "created_at": "2024-01-01T12:00:00Z"
}
```

**Order Status Changed**:
```json
{
  "order_id": "ORD-12345",
  "status": "CONFIRMED",
  "updated_at": "2024-01-01T12:30:00Z",
  "items": [...]
}
```

**Low Stock Alert**:
```json
{
  "article_id": "ART-001",
  "current_quantity": 5,
  "min_stock": 10,
  "alerted_at": "2024-01-01T12:00:00Z"
}
```

## Ejemplos de Uso

### Crear un nuevo artículo
```bash
curl -X POST http://localhost:8080/api/stock/articles \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "ART-001",
    "quantity": 100,
    "min_stock": 10,
    "max_stock": 500,
    "location": "Warehouse A"
  }'
```

### Reponer stock
```bash
curl -X PUT http://localhost:8080/api/stock/replenish \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "ART-001",
    "quantity": 50,
    "reason": "Weekly replenishment"
  }'
```

### Reservar stock
```bash
curl -X POST http://localhost:8080/api/stock/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "ART-001",
    "quantity": 2,
    "order_id": "ORD-12345"
  }'
```

## Desarrollo

### Ejecutar migraciones
```bash
# Aplicar migraciones
go run cmd/migrate.go up

# Rollback
go run cmd/migrate.go down

# Ver versión actual
go run cmd/migrate.go version

# Forzar versión específica
go run cmd/migrate.go force 1
```

### Docker
```bash
# Construir imagen
docker build -t stockgo .

# Ejecutar contenedor
docker run -p 8080:8080 stockgo
```

## Monitoreo

El servicio incluye:
- Logs estructurados con request ID
- Health check endpoint
- Métricas de stock bajo
- Historial completo de eventos
- Cache con Redis para mejor performance

## Próximas Funcionalidades

- [ ] Métricas con Prometheus
- [ ] Reservas con expiración automática
- [ ] Notificaciones por email/webhook
- [ ] API de reportes avanzados
- [ ] Soporte para múltiples ubicaciones
- [ ] Integración con sistemas de forecasting