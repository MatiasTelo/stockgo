# Integración con Microservicio de Autenticación

## 📋 Descripción

El servicio de Stock se integra con el microservicio de autenticación (AuthGO) para validar tokens en los endpoints GET. Utiliza **Redis como caché** para mejorar el rendimiento y reducir las llamadas al servicio de autenticación.

## 🔐 Endpoints Protegidos

Los siguientes endpoints requieren autenticación mediante token Bearer:

- `GET /api/stock/articles` - Obtener todos los artículos
- `GET /api/stock/articles/:articleId` - Obtener un artículo específico
- `GET /api/stock/articles/:articleId/events` - Obtener eventos de un artículo

## ⚙️ Configuración

### Variables de Entorno

Agrega la siguiente variable en tu archivo `.env`:

```bash
AUTH_SERVICE_URL=http://localhost:3000
```

**Valores por defecto:**
- Si no se especifica, usa: `http://localhost:3000`
- Ajusta el puerto según tu configuración del servicio de autenticación

## 🧪 Cómo Probar

### 1. Asegúrate de que Redis esté corriendo
```bash
redis-server
# O si usas Docker:
docker run -d -p 6379:6379 redis:latest
```

### 2. Asegúrate de que el servicio de Auth esté corriendo
```

## 🔄 Flujo de Autenticación

1. **Cliente** envía request con token en el header:
   ```
   Authorization: Bearer <token>
   ```

2. **StockGO Middleware** extrae el token y busca en **Redis Cache**:
   - Si encuentra el token en caché → retorna los datos del usuario inmediatamente
   - Si NO encuentra el token → continúa al paso 3

3. **StockGO** llama al servicio de autenticación:
   ```
   GET http://localhost:3000/users/current
   Authorization: Bearer <token>
   ```

4. **AuthGO** valida el token y responde con los datos del usuario:
   ```json
   {
       "id": "user-123",
       "username": "john_doe",
       "email": "john@example.com",
       "role": "admin"
   }
   ```

5. **StockGO** guarda los datos en Redis (TTL: 10 minutos) y continúa con la petición

6. Si el token es inválido, responde con error 401
