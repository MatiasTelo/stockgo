# StockGO - Microservicio de Gestión de Inventario

![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-blue.svg)
![Fiber](https://img.shields.io/badge/Fiber-v2-green.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

**StockGO** es un microservicio especializado en la gestión de inventario y stock para sistemas de e-commerce, desarrollado en Go con arquitectura limpia y enfoque en alta disponibilidad.

## 🎯 Características Principales

- ✅ **Gestión completa de inventario** - Control total del stock de artículos
- ✅ **Operaciones de stock** - Reponer, deducir, reservar y cancelar reservas
- ✅ **Auditoría completa** - Registro de todos los eventos y movimientos
- ✅ **Alertas de stock bajo** - Monitoreo automático de niveles mínimos
- ✅ **API REST robusta** - Endpoints optimizados para alta concurrencia
- ✅ **Validaciones de integridad** - Constraints de base de datos y validaciones de negocio
- ✅ **Arquitectura limpia** - Separación clara de responsabilidades

## 🏗️ Arquitectura

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   API Client    │────│ StockGO API  │────│  PostgreSQL DB  │
└─────────────────┘    └──────────────┘    └─────────────────┘
                              │
                       ┌──────────────┐
                       │ Stock Events │
                       │  (Auditoría) │
                       └──────────────┘
```

### Stack Tecnológico

- **Backend**: Go 1.25+ con Fiber Framework
- **Base de Datos**: PostgreSQL 13+
- **Migraciones**: golang-migrate
- **Validaciones**: Validator v10
- **Testing**: Postman Collection incluida

## 🚀 Instalación y Configuración

### Prerequisitos

- Go 1.25 o superior
- PostgreSQL 13 o superior
- Git

### 1. Clonar el Repositorio

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