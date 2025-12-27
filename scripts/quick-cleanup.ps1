# Скрипт для быстрой очистки базы данных (PowerShell версия для Windows)
# Использование: .\scripts\quick-cleanup.ps1 [view|reset|full|kafka|redis|all]

param(
    [string]$Command = "help"
)

switch ($Command) {
    "view" {
        Write-Host "📊 Показываю активные аренды..." -ForegroundColor Cyan
        docker exec postgres psql -U user -d bikerent -c @"
SELECT 
    r.id as rent_id,
    r.user_id,
    b.name as bike_name,
    b.status as bike_status,
    r.status as rent_status
FROM rents r
JOIN bikes b ON r.bike_id = b.id
WHERE r.status = 'active';
"@
        Write-Host ""
        Write-Host "🚲 Статус всех велосипедов:" -ForegroundColor Cyan
        docker exec postgres psql -U user -d bikerent -c "SELECT name, status, location FROM bikes ORDER BY name;"
        Write-Host ""
        Write-Host "📨 Kafka топики:" -ForegroundColor Cyan
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --list 2>$null
        Write-Host ""
        Write-Host "💾 Redis ключи:" -ForegroundColor Cyan
        docker exec redis redis-cli KEYS "*"
    }
    
    "reset" {
        Write-Host "🔄 Завершаю все активные аренды и освобождаю велосипеды..." -ForegroundColor Yellow
        docker exec postgres psql -U user -d bikerent -c @"
UPDATE rents SET status = 'completed', end_time = NOW() WHERE status = 'active';
UPDATE bikes SET status = 'available';
"@
        Write-Host "✅ Готово! Все велосипеды освобождены." -ForegroundColor Green
    }
    
    "full" {
        Write-Host "🗑️  ПОЛНАЯ ОЧИСТКА: удаляю все аренды и сбрасываю велосипеды..." -ForegroundColor Red
        docker exec postgres psql -U user -d bikerent -c @"
TRUNCATE TABLE rents CASCADE;
UPDATE bikes SET status = 'available';
"@
        Write-Host "✅ База данных полностью очищена!" -ForegroundColor Green
    }
    
    "kafka" {
        Write-Host "🧹 Очистка Kafka топиков..." -ForegroundColor Yellow
        Write-Host "Удаляю топики bike-rent-events и bike-status-events..." -ForegroundColor Gray
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-rent-events 2>$null
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-status-events 2>$null
        Start-Sleep -Seconds 2
        Write-Host "Пересоздаю топики..." -ForegroundColor Gray
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-rent-events --partitions 1 --replication-factor 1 2>$null
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-status-events --partitions 1 --replication-factor 1 2>$null
        Write-Host "✅ Kafka топики очищены!" -ForegroundColor Green
    }
    
    "redis" {
        Write-Host "🧹 Очистка Redis..." -ForegroundColor Yellow
        docker exec redis redis-cli FLUSHDB
        Write-Host "✅ Redis очищен!" -ForegroundColor Green
    }
    
    "all" {
        Write-Host "🗑️  ПОЛНАЯ ОЧИСТКА ВСЕГО ПРОЕКТА..." -ForegroundColor Red
        Write-Host ""
        Write-Host "1. Очистка PostgreSQL..." -ForegroundColor Yellow
        docker exec postgres psql -U user -d bikerent -c @"
TRUNCATE TABLE rents CASCADE;
UPDATE bikes SET status = 'available';
"@
        Write-Host "✅ PostgreSQL очищен" -ForegroundColor Green
        Write-Host ""
        Write-Host "2. Очистка Kafka..." -ForegroundColor Yellow
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-rent-events 2>$null
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-status-events 2>$null
        Start-Sleep -Seconds 2
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-rent-events --partitions 1 --replication-factor 1 2>$null
        docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-status-events --partitions 1 --replication-factor 1 2>$null
        Write-Host "✅ Kafka очищен" -ForegroundColor Green
        Write-Host ""
        Write-Host "3. Очистка Redis..." -ForegroundColor Yellow
        docker exec redis redis-cli FLUSHDB
        Write-Host "✅ Redis очищен" -ForegroundColor Green
        Write-Host ""
        Write-Host "🎉 ВСЁ ОЧИЩЕНО! Проект в исходном состоянии." -ForegroundColor Green
    }
    
    default {
        Write-Host "📖 Использование: .\scripts\quick-cleanup.ps1 [команда]" -ForegroundColor White
        Write-Host ""
        Write-Host "Команды:" -ForegroundColor Yellow
        Write-Host "  view   - Посмотреть все активные аренды, велосипеды, Kafka топики и Redis ключи"
        Write-Host "  reset  - Завершить все аренды и освободить велосипеды (мягкая очистка)"
        Write-Host "  full   - Полная очистка PostgreSQL (удалить все аренды)"
        Write-Host "  kafka  - Очистить все топики Kafka (пересоздать)"
        Write-Host "  redis  - Очистить все ключи в Redis"
        Write-Host "  all    - Полная очистка всего проекта (PostgreSQL + Kafka + Redis)"
        Write-Host ""
        Write-Host "Примеры:" -ForegroundColor Cyan
        Write-Host "  .\scripts\quick-cleanup.ps1 view    # Посмотреть текущее состояние"
        Write-Host "  .\scripts\quick-cleanup.ps1 reset   # Освободить все велосипеды"
        Write-Host "  .\scripts\quick-cleanup.ps1 kafka   # Очистить Kafka"
        Write-Host "  .\scripts\quick-cleanup.ps1 redis   # Очистить Redis"
        Write-Host "  .\scripts\quick-cleanup.ps1 all     # Очистить ВСЁ"
    }
}

