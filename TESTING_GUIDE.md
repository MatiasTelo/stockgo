# StockGO Microservice - Guía de Pruebas con Postman

## 🚀 Inicio Rápido

### 1. Iniciar el Microservicio
```bash
cd "c:\Users\matia\Desktop\Facultad\Cuarto año\Microservicios\Ecommerce\stockgo-main"
go run cmd/main.go
```

El servidor se iniciará en `http://localhost:8080`

### 2. Importar Colección en Postman

1. Abrir Postman
2. Hacer clic en "Import" 
3. Seleccionar el archivo `postman_collection.json`
4. La colección "StockGO Microservice Tests" aparecerá en tu workspace

## 📋 Endpoints Disponibles

### Health Check
- **GET** `/health` - Verificar que el servicio está funcionando

### Gestión de Artículos
- **POST** `/api/stock/articles` - Crear nuevo artículo
- **GET** `/api/stock/{article_id}` - Obtener información de stock

### Operaciones de Stock  
- **POST** `/api/stock/replenish` - Reabastecer stock
- **POST** `/api/stock/deduct` - Deducir stock
- **POST** `/api/stock/reserve` - Reservar stock
- **POST** `/api/stock/cancel-reservation` - Cancelar reserva

### Consultas
- **GET** `/api/stock/low-stock` - Obtener artículos con bajo stock

## 🧪 Secuencia de Pruebas Recomendada

### Paso 1: Verificar Servicio
```bash
GET http://localhost:8080/health
```

### Paso 2: Crear Artículo
```json
POST /api/stock/articles
{
    "article_id": "ART-001",
    "name": "Producto de Prueba",
    "description": "Descripción del producto",
    "initial_stock": 100,
    "min_stock": 10,
    "max_stock": 500,
    "unit_price": 25.50,
    "location": "A1-B2-C3",
    "metadata": {
        "category": "electronics",
        "supplier": "Supplier ABC"
    }
}
```

### Paso 3: Verificar Stock Inicial
```bash
GET /api/stock/ART-001
```
**Respuesta esperada:**
```json
{
    "article_id": "ART-001",
    "name": "Producto de Prueba", 
    "current_stock": 100,
    "available_stock": 100,
    "reserved_stock": 0,
    "min_stock": 10,
    "max_stock": 500
}
```

### Paso 4: Reabastecer Stock
```json
POST /api/stock/replenish
{
    "article_id": "ART-001",
    "quantity": 50,
    "reason": "Reabastecimiento mensual",
    "supplier": "Supplier ABC",
    "batch_number": "BATCH-2025-001"
}
```

### Paso 5: Reservar Stock
```json
POST /api/stock/reserve
{
    "article_id": "ART-001",
    "quantity": 20,
    "order_id": "ORDER-001",
    "customer_id": "CUSTOMER-001",
    "expiration_minutes": 30
}
```

### Paso 6: Verificar Reserva
```bash
GET /api/stock/ART-001
```
**Respuesta esperada:**
```json
{
    "article_id": "ART-001",
    "current_stock": 150,
    "available_stock": 130,
    "reserved_stock": 20
}
```

### Paso 7: Deducir Stock
```json
POST /api/stock/deduct
{
    "article_id": "ART-001",
    "quantity": 10,
    "reason": "Venta directa",
    "transaction_id": "TXN-001"
}
```

### Paso 8: Cancelar Reserva
```json
POST /api/stock/cancel-reservation
{
    "order_id": "ORDER-001",
    "article_id": "ART-001",
    "quantity": 20,
    "reason": "Order cancelled by customer"
}
```

## 🔍 Casos de Prueba Avanzados

### Test de Stock Insuficiente
```json
POST /api/stock/reserve
{
    "article_id": "ART-001",
    "quantity": 99999,
    "order_id": "ORDER-FAIL"
}
```
**Respuesta esperada:** `400 Bad Request` con mensaje de error

### Test de Artículo Inexistente
```bash
GET /api/stock/INVALID-ARTICLE
```
**Respuesta esperada:** `404 Not Found`

### Test de Bajo Stock
```bash
GET /api/stock/low-stock?threshold=50
```

## 📊 Validaciones Automáticas

La colección de Postman incluye tests automáticos que verifican:

- ✅ Códigos de estado HTTP correctos
- ✅ Estructura de respuesta válida
- ✅ Consistencia de datos de stock
- ✅ Manejo correcto de errores
- ✅ Validación de reservas

### Ejecutar Todos los Tests
1. En Postman, seleccionar la colección
2. Hacer clic en "Run collection"
3. Configurar el número de iteraciones
4. Hacer clic en "Run StockGO Microservice Tests"

## 🔧 Variables de Entorno

La colección utiliza estas variables que puedes personalizar:

- `base_url`: URL del microservicio (default: `http://localhost:8080`)
- `article_id`: ID del artículo para pruebas (default: `ART-001`)
- `order_id`: ID de la orden para pruebas (default: `ORDER-001`)

### Cambiar Variables
1. En Postman, ir a la pestaña "Variables" de la colección
2. Modificar los valores según necesites
3. Guardar los cambios

## 📈 Monitoreo y Logs

El microservicio genera logs detallados. Para ver la actividad:

```bash
# Los logs aparecerán en la consola donde ejecutaste:
go run cmd/main.go
```

**Ejemplo de logs:**
```
2025/09/25 10:30:15 Creating article: ART-001
2025/09/25 10:30:20 Reserving 20 units for order ORDER-001
2025/09/25 10:30:25 Stock reserved successfully
```

## 🐛 Troubleshooting

### Error: "Connection refused"
- Verificar que el servicio esté ejecutándose
- Confirmar que el puerto 8080 esté disponible

### Error: "Database connection failed"  
- Verificar configuración de PostgreSQL
- Ejecutar migraciones si es necesario

### Error: "Redis/RabbitMQ warnings"
- Normal en desarrollo local sin estos servicios
- El microservicio funciona sin Redis/RabbitMQ para pruebas básicas

## 🎯 Casos de Uso Reales

### Flujo de E-commerce Completo
1. **Crear producto** → POST `/api/stock/articles`
2. **Cliente hace pedido** → POST `/api/stock/reserve` 
3. **Cliente paga** → Confirmar reserva (vía RabbitMQ)
4. **Cliente cancela** → POST `/api/stock/cancel-reservation`
5. **Restock periódico** → POST `/api/stock/replenish`
6. **Monitoreo** → GET `/api/stock/low-stock`

### Pruebas de Carga
- Usar Postman Runner con múltiples iteraciones
- Probar reservas concurrentes del mismo artículo
- Validar consistencia de stock bajo carga

¡Listo para probar tu microservicio! 🚀