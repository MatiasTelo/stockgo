@echo off
REM Script para Windows PowerShell - Pruebas del microservicio StockGO
REM Uso: test_api.bat

set BASE_URL=http://localhost:8080
set ARTICLE_ID=ART-TEST-%RANDOM%
set ORDER_ID=ORDER-TEST-%RANDOM%

echo 🧪 Iniciando pruebas del microservicio StockGO...
echo 📍 Base URL: %BASE_URL%
echo 📦 Article ID: %ARTICLE_ID%
echo 🛒 Order ID: %ORDER_ID%
echo.

REM Health Check
echo 1️⃣  Health Check
curl -s "%BASE_URL%/health"
echo.
echo.

REM Crear artículo
echo 2️⃣  Creando artículo...
curl -s -X POST "%BASE_URL%/api/stock/articles" ^
  -H "Content-Type: application/json" ^
  -d "{\"article_id\":\"%ARTICLE_ID%\",\"name\":\"Producto de Prueba Automatizada\",\"description\":\"Producto creado por script de pruebas\",\"initial_stock\":100,\"min_stock\":10,\"max_stock\":500,\"unit_price\":29.99,\"location\":\"AUTO-TEST-LOC\",\"metadata\":{\"test\":true,\"created_by\":\"test_script\"}}"
echo.
echo.

REM Obtener stock inicial
echo 3️⃣  Consultando stock inicial...
curl -s "%BASE_URL%/api/stock/%ARTICLE_ID%"
echo.
echo.

REM Reabastecer
echo 4️⃣  Reabasteciendo stock...
curl -s -X POST "%BASE_URL%/api/stock/replenish" ^
  -H "Content-Type: application/json" ^
  -d "{\"article_id\":\"%ARTICLE_ID%\",\"quantity\":50,\"reason\":\"Test replenishment\",\"supplier\":\"Test Supplier\",\"batch_number\":\"BATCH-%ORDER_ID%\"}"
echo.
echo.

REM Reservar stock
echo 5️⃣  Reservando stock...
curl -s -X POST "%BASE_URL%/api/stock/reserve" ^
  -H "Content-Type: application/json" ^
  -d "{\"article_id\":\"%ARTICLE_ID%\",\"quantity\":25,\"order_id\":\"%ORDER_ID%\",\"customer_id\":\"TEST-CUSTOMER\",\"expiration_minutes\":30}"
echo.
echo.

REM Consultar stock después de reserva
echo 6️⃣  Stock después de reserva...
curl -s "%BASE_URL%/api/stock/%ARTICLE_ID%"
echo.
echo.

REM Deducir stock
echo 7️⃣  Deduciendo stock...
curl -s -X POST "%BASE_URL%/api/stock/deduct" ^
  -H "Content-Type: application/json" ^
  -d "{\"article_id\":\"%ARTICLE_ID%\",\"quantity\":15,\"reason\":\"Test sale\",\"transaction_id\":\"TXN-%ORDER_ID%\"}"
echo.
echo.

REM Cancelar reserva
echo 8️⃣  Cancelando reserva...
curl -s -X POST "%BASE_URL%/api/stock/cancel-reservation" ^
  -H "Content-Type: application/json" ^
  -d "{\"order_id\":\"%ORDER_ID%\",\"article_id\":\"%ARTICLE_ID%\",\"quantity\":25,\"reason\":\"Test cancellation\"}"
echo.
echo.

REM Stock final
echo 9️⃣  Stock final...
curl -s "%BASE_URL%/api/stock/%ARTICLE_ID%"
echo.
echo.

REM Low stock check
echo 🔟 Verificando artículos con bajo stock...
curl -s "%BASE_URL%/api/stock/low-stock?threshold=50"
echo.
echo.

echo ✅ Pruebas completadas!
pause