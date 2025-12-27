#!/bin/bash

# Скрипт для быстрой очистки базы данных
# Использование: ./scripts/quick-cleanup.sh [view|reset|full|kafka|redis|all]

COMMAND=${1:-help}

case $COMMAND in
  view)
    echo "📊 Показываю активные аренды..."
    docker exec postgres psql -U user -d bikerent -c "
      SELECT 
        r.id as rent_id,
        r.user_id,
        b.name as bike_name,
        b.status as bike_status,
        r.status as rent_status
      FROM rents r
      JOIN bikes b ON r.bike_id = b.id
      WHERE r.status = 'active';
    "
    echo ""
    echo "🚲 Статус всех велосипедов:"
    docker exec postgres psql -U user -d bikerent -c "
      SELECT name, status, location FROM bikes ORDER BY name;
    "
    echo ""
    echo "📨 Kafka топики:"
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null
    echo ""
    echo "💾 Redis ключи:"
    docker exec redis redis-cli KEYS "*"
    ;;
    
  reset)
    echo "🔄 Завершаю все активные аренды и освобождаю велосипеды..."
    docker exec postgres psql -U user -d bikerent -c "
      UPDATE rents SET status = 'completed', end_time = NOW() WHERE status = 'active';
      UPDATE bikes SET status = 'available';
    "
    echo "✅ Готово! Все велосипеды освобождены."
    ;;
    
  full)
    echo "🗑️  ПОЛНАЯ ОЧИСТКА: удаляю все аренды и сбрасываю велосипеды..."
    docker exec postgres psql -U user -d bikerent -c "
      TRUNCATE TABLE rents CASCADE;
      UPDATE bikes SET status = 'available';
    "
    echo "✅ База данных полностью очищена!"
    ;;
    
  kafka)
    echo "🧹 Очистка Kafka топиков..."
    echo "Удаляю топики bike-rent-events и bike-status-events..."
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-rent-events 2>/dev/null
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-status-events 2>/dev/null
    sleep 2
    echo "Пересоздаю топики..."
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-rent-events --partitions 1 --replication-factor 1 2>/dev/null
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-status-events --partitions 1 --replication-factor 1 2>/dev/null
    echo "✅ Kafka топики очищены!"
    ;;
    
  redis)
    echo "🧹 Очистка Redis..."
    docker exec redis redis-cli FLUSHDB
    echo "✅ Redis очищен!"
    ;;
    
  all)
    echo "🗑️  ПОЛНАЯ ОЧИСТКА ВСЕГО ПРОЕКТА..."
    echo ""
    echo "1. Очистка PostgreSQL..."
    docker exec postgres psql -U user -d bikerent -c "
      TRUNCATE TABLE rents CASCADE;
      UPDATE bikes SET status = 'available';
    "
    echo "✅ PostgreSQL очищен"
    echo ""
    echo "2. Очистка Kafka..."
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-rent-events 2>/dev/null
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-status-events 2>/dev/null
    sleep 2
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-rent-events --partitions 1 --replication-factor 1 2>/dev/null
    docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-status-events --partitions 1 --replication-factor 1 2>/dev/null
    echo "✅ Kafka очищен"
    echo ""
    echo "3. Очистка Redis..."
    docker exec redis redis-cli FLUSHDB
    echo "✅ Redis очищен"
    echo ""
    echo "🎉 ВСЁ ОЧИЩЕНО! Проект в исходном состоянии."
    ;;
    
  *)
    echo "📖 Использование: $0 [команда]"
    echo ""
    echo "Команды:"
    echo "  view   - Посмотреть все активные аренды, велосипеды, Kafka топики и Redis ключи"
    echo "  reset  - Завершить все аренды и освободить велосипеды (мягкая очистка)"
    echo "  full   - Полная очистка PostgreSQL (удалить все аренды)"
    echo "  kafka  - Очистить все топики Kafka (пересоздать)"
    echo "  redis  - Очистить все ключи в Redis"
    echo "  all    - Полная очистка всего проекта (PostgreSQL + Kafka + Redis)"
    echo ""
    echo "Примеры:"
    echo "  $0 view    # Посмотреть текущее состояние"
    echo "  $0 reset   # Освободить все велосипеды"
    echo "  $0 kafka   # Очистить Kafka"
    echo "  $0 redis   # Очистить Redis"
    echo "  $0 all     # Очистить ВСЁ"
    ;;
esac

