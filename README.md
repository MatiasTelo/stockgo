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
- **Cache**: Redis (opcional)
- **Mensajería**: RabbitMQ (opcional, futuras integraciones)
- **Validaciones**: go-playground/validator/v10

### Estructura de Capas
```
┌─────────────────┐
│   Handlers      │ ← API REST / HTTP
├─────────────────┤
│   Services      │ ← Lógica de negocio
├─────────────────┤
│  Repository     │ ← Acceso a datos
├─────────────────┤
│   Database      │ ← PostgreSQL
└─────────────────┘
```

## 📁 Estructura del Proyecto

```
stockgo/
├── cmd/
│   ├── main.go                    # Punto de entrada principal
│   └── migrate.go                 # Utilidad de migraciones
├── internal/
│   ├── config/
│   │   └── config.go              # Configuración de la aplicación
│   ├── database/
│   │   └── database.go            # Conexión a PostgreSQL
│   ├── handlers/                  # Controladores HTTP
│   │   ├── add_article.go         # Gestión de artículos
│   │   ├── replenish_stock.go     # Reabastecimiento
│   │   ├── deduct_stock.go        # Deducciones
│   │   ├── reserve_stock.go       # Reservas
│   │   ├── cancel_reservation.go  # Cancelaciones
│   │   └── low_stock.go           # Alertas de stock bajo
│   ├── models/                    # Modelos de datos
│   │   ├── stock.go               # Modelo Stock
│   │   └── stock_event.go         # Modelo StockEvent
│   ├── repository/                # Capa de acceso a datos
│   │   ├── stock_repository.go    # Operaciones de Stock
│   │   └── stock_event_repository.go # Operaciones de StockEvent
│   ├── service/
│   │   └── stock_service.go       # Lógica de negocio central
│   └── messaging/                 # RabbitMQ (futuro)
├── migrations/                    # Migraciones SQL
│   ├── 001_create_stocks_table.up.sql
│   ├── 002_create_stock_events_table.up.sql
│   └── ...
├── postman_collection.json        # Colección Postman
├── POSTMAN_EXAMPLES.md            # Ejemplos de uso de API
├── Dockerfile                     # Containerización
├── go.mod                         # Dependencias Go
└── README.md
```

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

## 🧪 Testing

### Con Postman
1. Importar `postman_collection.json`
2. Configurar variables: `base_url=http://localhost:8080`
3. Ejecutar los endpoints según `POSTMAN_EXAMPLES.md`

### Flujo de prueba completo
```bash
# 1. Crear artículo
curl -X POST http://localhost:8080/api/articles \
  -H "Content-Type: application/json" \
  -d '{"article_id":"TEST-001","quantity":100,"min_stock":10}'

# 2. Reservar stock
curl -X PUT http://localhost:8080/api/stock/reserve \
  -H "Content-Type: application/json" \
  -d '{"article_id":"TEST-001","quantity":5,"order_id":"ORDER-123"}'

# 3. Confirmar reserva
curl -X PUT http://localhost:8080/api/stock/confirm-reservation \
  -H "Content-Type: application/json" \
  -d '{"article_id":"TEST-001","order_id":"ORDER-123"}'
```

## � Comandos de Mantenimiento

```bash
# Ver estado de migraciones
go run cmd/migrate.go version

# Aplicar migraciones
go run cmd/migrate.go up

# Revertir migración
go run cmd/migrate.go down 1

# Compilar para producción
go build -o stockgo cmd/main.go
```

## 🐳 Docker

```dockerfile
# Construir imagen
docker build -t stockgo .

# Ejecutar contenedor
docker run -p 8080:8080 -e DB_HOST=host.docker.internal stockgo
```

## 📝 Validaciones Implementadas

- ✅ **Prevención de reservas duplicadas** por order_id
- ✅ **Validación de stock suficiente** antes de operaciones
- ✅ **Constrains de integridad** en base de datos
- ✅ **Validación de tipos de datos** en entrada
- ✅ **Auditoría completa** de todas las operaciones

## 🚀 Características de Producción

- ✅ **Connection pooling** optimizado para PostgreSQL
- ✅ **Logging estructurado** con trazabilidad
- ✅ **Health checks** para monitoreo
- ✅ **Índices de base de datos** optimizados
- ✅ **Manejo de errores** robusto
- ✅ **Documentación API** completa

---

📚 **Documentación adicional**: Ver `POSTMAN_EXAMPLES.md` para ejemplos detallados de uso de la API.

```bash
git clone https://github.com/MatiasTelo/stockgo.git
cd stockgo
```

### 2. Configurar Variables de Entorno

Crea el archivo `.env` en el directorio raíz:

```env
# Servidor
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Base de Datos PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=admin
DB_DATABASE=stockdb
DB_SSLMODE=disable

# Opcionales (para futuras integraciones)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=ecommerce
RABBITMQ_QUEUE=stock_events
```

### 3. Configurar PostgreSQL

```sql
-- Conectar como usuario postgres
psql -U postgres

-- Crear la base de datos
CREATE DATABASE stockdb;

-- Salir
\q
```

### 4. Instalar Dependencias

```bash
go mod download
```

### 5. Ejecutar Migraciones

```bash
go run cmd/migrate_runner.go up
```

### 6. Iniciar el Servidor

```bash
go run cmd/main.go
```

El servidor estará disponible en `http://localhost:8080`

## 📋 API Endpoints

### Health Check
- `GET /health` - Estado del servicio

### Gestión de Artículos

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/api/stock/articles` | Crear nuevo artículo en inventario |
| `GET` | `/api/stock/articles` | Listar todos los artículos |
| `GET` | `/api/stock/articles/{articleId}` | Obtener información de stock |
| `GET` | `/api/stock/articles/{articleId}/events` | Obtener historial de eventos |

### Operaciones de Stock

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `PUT` | `/api/stock/replenish` | Reabastecer stock (JSON body) |
| `PUT` | `/api/stock/articles/{articleId}/replenish` | Reabastecer stock específico |
| `PUT` | `/api/stock/deduct` | Deducir stock (JSON body) |
| `PUT` | `/api/stock/articles/{articleId}/deduct` | Deducir stock específico |

### Reservas

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/api/stock/reserve` | Reservar stock para orden |
| `POST` | `/api/stock/articles/{articleId}/reserve` | Reservar stock específico |
| `DELETE` | `/api/stock/reservations` | Cancelar reserva (JSON body) |
| `DELETE` | `/api/stock/orders/{orderId}/reservations/{articleId}` | Cancelar reserva específica |
| `POST` | `/api/stock/reservations/confirm` | Confirmar reserva (convierte a venta) |
| `POST` | `/api/stock/orders/{orderId}/reservations/{articleId}/confirm` | Confirmar reserva específica |

### Consultas y Monitoreo

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `GET` | `/api/stock/low-stock` | Artículos con stock bajo |
| `GET` | `/api/stock/alerts/summary` | Resumen de alertas de stock |

## 💾 Modelo de Datos

### Tabla `stocks`
```sql
id           UUID PRIMARY KEY         -- ID único del registro
article_id   VARCHAR(100) UNIQUE      -- ID del artículo (del catálogo)
quantity     INTEGER NOT NULL         -- Cantidad total disponible
reserved     INTEGER DEFAULT 0        -- Cantidad reservada
min_stock    INTEGER DEFAULT 0        -- Stock mínimo (alerta)
max_stock    INTEGER DEFAULT 0        -- Stock máximo
location     VARCHAR(255)             -- Ubicación en almacén
created_at   TIMESTAMP                -- Fecha de creación
updated_at   TIMESTAMP                -- Última actualización
```

### Tabla `stock_events`
```sql
id           UUID PRIMARY KEY         -- ID del evento
article_id   VARCHAR(100)             -- ID del artículo
event_type   VARCHAR(50)              -- Tipo: ADD, REPLENISH, DEDUCT, RESERVE, etc.
quantity     INTEGER                  -- Cantidad del movimiento
order_id     VARCHAR(100)             -- ID de orden (si aplica)
reason       TEXT                     -- Motivo del movimiento
metadata     JSONB                    -- Información adicional
created_at   TIMESTAMP                -- Fecha del evento
```

### Tipos de Eventos
- `ADD` - Creación de artículo
- `REPLENISH` - Reabastecimiento
- `DEDUCT` - Deducción directa (venta)
- `RESERVE` - Reserva de stock
- `CANCEL_RESERVE` - Cancelación de reserva
- `LOW_STOCK` - Alerta de stock bajo

## 🧪 Testing

### Usando Postman

1. Importa la colección `postman_collection.json`
2. Consulta `POSTMAN_EXAMPLES.md` para ejemplos detallados
3. Configura las variables:
   - `base_url`: `http://localhost:8080`
   - `article_id`: `ART-001`
   - `order_id`: `ORDER-123`

### Ejemplo de Flujo Completo

```bash
# 1. Health Check
curl http://localhost:8080/health

# 2. Crear artículo
curl -X POST http://localhost:8080/api/stock/articles \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "LAPTOP-001",
    "quantity": 50,
    "min_stock": 5,
    "max_stock": 200,
    "location": "A1-B2-C3"
  }'

# 3. Consultar stock
curl http://localhost:8080/api/stock/articles/LAPTOP-001

# 4. Reservar stock
curl -X POST http://localhost:8080/api/stock/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "LAPTOP-001",
    "quantity": 2,
    "order_id": "ORDER-123"
  }'

# 5. Confirmar reserva (venta)
curl -X POST http://localhost:8080/api/stock/reservations/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "LAPTOP-001",
    "quantity": 2,
    "order_id": "ORDER-123",
    "reason": "Pago confirmado"
  }'
```

## 🔧 Comandos de Migración

```bash
# Aplicar migraciones
go run cmd/migrate_runner.go up

# Revertir migraciones
go run cmd/migrate_runner.go down

# Ver versión actual
go run cmd/migrate_runner.go version

# Forzar versión específica
go run cmd/migrate_runner.go force 1
```

## 📊 Estructura del Proyecto

```
stockgo/
├── cmd/
│   ├── main.go              # Aplicación principal
│   └── migrate_runner.go    # Script de migraciones
├── internal/
│   ├── config/              # Configuración de la aplicación
│   ├── database/            # Conexiones a base de datos
│   ├── handlers/            # Handlers REST por funcionalidad
│   │   ├── add_article.go       # Gestión de artículos
│   │   ├── replenish_stock.go   # Reabastecimiento
│   │   ├── deduct_stock.go      # Deducciones
│   │   ├── reserve_stock.go     # Reservas
│   │   ├── cancel_reservation.go # Cancelaciones
│   │   └── low_stock.go         # Alertas de stock bajo
│   ├── messaging/           # RabbitMQ (futuro)
│   ├── models/              # Modelos de datos
│   ├── repository/          # Acceso a datos
│   └── service/             # Lógica de negocio
├── migrations/              # Migraciones SQL
│   ├── 001_create_stocks_table.up.sql
│   ├── 002_create_stock_events_table.up.sql
│   └── ...
├── .env.example            # Variables de entorno de ejemplo
├── postman_collection.json # Colección de Postman
├── POSTMAN_EXAMPLES.md     # Ejemplos detallados de uso
├── Dockerfile              # Para containerización
├── go.mod                  # Dependencias de Go
└── README.md
```

## 🚀 Despliegue con Docker

```bash
# Construir imagen
docker build -t stockgo .

# Ejecutar contenedor
docker run -p 8080:8080 \
  -e DB_HOST=your-postgres-host \
  -e DB_PASSWORD=your-password \
  stockgo
```

## 🔍 Consideraciones de Producción

- ✅ **Conexión pool** de PostgreSQL configurado
- ✅ **Logging estructurado** con request IDs
- ✅ **Health checks** para monitoreo
- ✅ **Constraints de BD** para integridad de datos
- ✅ **Validaciones** de entrada robustas
- ✅ **Índices optimizados** para consultas frecuentes
- ✅ **Auditoría completa** con stock_events

## 🤝 Contribuir

1. Fork el repositorio
2. Crea tu feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la branch (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📚 Documentación

- **API Examples**: Ver `POSTMAN_EXAMPLES.md`
- **Testing Guide**: Ver `TESTING_GUIDE.md`
- **Postman Collection**: Importar `postman_collection.json`

## 📝 Licencia

Este proyecto está bajo la Licencia MIT. Ver `LICENSE` para más detalles.

## 📞 Soporte

- **Issues**: [GitHub Issues](https://github.com/MatiasTelo/stockgo/issues)
- **Ejemplos**: Ver `POSTMAN_EXAMPLES.md`

---

⭐ **¡No olvides dar una estrella si este proyecto te fue útil!**