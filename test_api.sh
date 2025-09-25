#!/bin/bash

# Script para pruebas rápidas con curl
# Uso: ./test_api.sh

BASE_URL="http://localhost:8080"
ARTICLE_ID="ART-TEST-$(date +%s)"
ORDER_ID="ORDER-TEST-$(date +%s)"

echo "🧪 Iniciando pruebas del microservicio StockGO..."
echo "📍 Base URL: $BASE_URL"
echo "📦 Article ID: $ARTICLE_ID"
echo "🛒 Order ID: $ORDER_ID"
echo ""

# Health Check
echo "1️⃣  Health Check"
curl -s "$BASE_URL/health" | jq '.'
echo -e "\n"

# Crear artículo
echo "2️⃣  Creando artículo..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stock/articles" \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "'$ARTICLE_ID'",
    "name": "Producto de Prueba Automatizada",
    "description": "Producto creado por script de pruebas",
    "initial_stock": 100,
    "min_stock": 10,
    "max_stock": 500,
    "unit_price": 29.99,
    "location": "AUTO-TEST-LOC",
    "metadata": {
      "test": true,
      "created_by": "test_script"
    }
  }')
echo $CREATE_RESPONSE | jq '.'
echo -e "\n"

# Obtener stock inicial
echo "3️⃣  Consultando stock inicial..."
curl -s "$BASE_URL/api/stock/$ARTICLE_ID" | jq '.'
echo -e "\n"

# Reabastecer
echo "4️⃣  Reabasteciendo stock..."
REPLENISH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stock/replenish" \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "'$ARTICLE_ID'",
    "quantity": 50,
    "reason": "Test replenishment",
    "supplier": "Test Supplier",
    "batch_number": "BATCH-'$ORDER_ID'"
  }')
echo $REPLENISH_RESPONSE | jq '.'
echo -e "\n"

# Reservar stock
echo "5️⃣  Reservando stock..."
RESERVE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stock/reserve" \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "'$ARTICLE_ID'",
    "quantity": 25,
    "order_id": "'$ORDER_ID'",
    "customer_id": "TEST-CUSTOMER",
    "expiration_minutes": 30
  }')
echo $RESERVE_RESPONSE | jq '.'
echo -e "\n"

# Consultar stock después de reserva
echo "6️⃣  Stock después de reserva..."
curl -s "$BASE_URL/api/stock/$ARTICLE_ID" | jq '.'
echo -e "\n"

# Deducir stock
echo "7️⃣  Deduciendo stock..."
DEDUCT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stock/deduct" \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "'$ARTICLE_ID'",
    "quantity": 15,
    "reason": "Test sale",
    "transaction_id": "TXN-'$ORDER_ID'"
  }')
echo $DEDUCT_RESPONSE | jq '.'
echo -e "\n"

# Cancelar reserva
echo "8️⃣  Cancelando reserva..."
CANCEL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stock/cancel-reservation" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "'$ORDER_ID'",
    "article_id": "'$ARTICLE_ID'",
    "quantity": 25,
    "reason": "Test cancellation"
  }')
echo $CANCEL_RESPONSE | jq '.'
echo -e "\n"

# Stock final
echo "9️⃣  Stock final..."
FINAL_STOCK=$(curl -s "$BASE_URL/api/stock/$ARTICLE_ID")
echo $FINAL_STOCK | jq '.'
echo -e "\n"

# Low stock check
echo "🔟 Verificando artículos con bajo stock..."
curl -s "$BASE_URL/api/stock/low-stock?threshold=50" | jq '.'
echo -e "\n"

echo "✅ Pruebas completadas!"
echo "📊 Resumen final del artículo $ARTICLE_ID:"
echo $FINAL_STOCK | jq '{article_id, current_stock, available_stock, reserved_stock}'