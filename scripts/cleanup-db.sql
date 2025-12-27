-- ============================================
-- СКРИПТ ДЛЯ ОЧИСТКИ И УПРАВЛЕНИЯ ДАННЫМИ
-- PostgreSQL, Kafka, Redis
-- ============================================

-- ==========================================
-- POSTGRESQL КОМАНДЫ
-- ==========================================

-- 1. ПОСМОТРЕТЬ ВСЕ АКТИВНЫЕ АРЕНДЫ (с ID велосипедов и пользователей)
-- --------------------------------------------
SELECT 
    r.id as rent_id,
    r.user_id,
    r.bike_id,
    b.name as bike_name,
    b.location,
    r.start_time,
    r.status as rent_status,
    b.status as bike_status
FROM rents r
JOIN bikes b ON r.bike_id = b.id
WHERE r.status = 'active'
ORDER BY r.start_time DESC;

-- 2. ПОСМОТРЕТЬ СТАТУС ВСЕХ ВЕЛОСИПЕДОВ
-- --------------------------------------------
SELECT 
    id,
    name,
    status,
    location,
    created_at
FROM bikes
ORDER BY name;

-- 3. ЗАВЕРШИТЬ ВСЕ АКТИВНЫЕ АРЕНДЫ
-- --------------------------------------------
UPDATE rents 
SET status = 'completed', 
    end_time = NOW() 
WHERE status = 'active';

-- 4. СБРОСИТЬ ВСЕ ВЕЛОСИПЕДЫ В СТАТУС "AVAILABLE"
-- --------------------------------------------
UPDATE bikes 
SET status = 'available';

-- 5. ПОЛНАЯ ОЧИСТКА (осторожно! удаляет все данные)
-- --------------------------------------------
-- TRUNCATE TABLE rents CASCADE;
-- DELETE FROM bikes;
-- INSERT INTO bikes (name, status, location) VALUES
--     ('Bike 1', 'available', 'Location A'),
--     ('Bike 2', 'available', 'Location A'),
--     ('Bike 3', 'available', 'Location B'),
--     ('Bike 4', 'available', 'Location B'),
--     ('Bike 5', 'available', 'Location C');

-- 6. ЗАВЕРШИТЬ КОНКРЕТНУЮ АРЕНДУ ПО ID
-- --------------------------------------------
-- UPDATE rents 
-- SET status = 'completed', end_time = NOW() 
-- WHERE id = 'YOUR_RENT_ID_HERE';

-- 7. ОСВОБОДИТЬ КОНКРЕТНЫЙ ВЕЛОСИПЕД ПО ID
-- --------------------------------------------
-- UPDATE bikes 
-- SET status = 'available' 
-- WHERE id = 'YOUR_BIKE_ID_HERE';

-- ==========================================
-- KAFKA КОМАНДЫ (выполнять через docker exec)
-- ==========================================

-- Список топиков:
-- docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --list

-- Просмотр сообщений в топике:
-- docker exec kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic bike-rent-events --from-beginning --max-messages 10

-- Удалить топик:
-- docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic bike-rent-events

-- Создать топик заново:
-- docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic bike-rent-events --partitions 1 --replication-factor 1

-- Описание топика:
-- docker exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic bike-rent-events

-- ==========================================
-- REDIS КОМАНДЫ (выполнять через docker exec)
-- ==========================================

-- Просмотр всех ключей:
-- docker exec redis redis-cli KEYS "*"

-- Получить значение ключа:
-- docker exec redis redis-cli GET "stats:active:count"

-- Удалить конкретный ключ:
-- docker exec redis redis-cli DEL "stats:active:count"

-- Удалить все ключи в текущей БД (DB 0):
-- docker exec redis redis-cli FLUSHDB

-- Удалить все ключи во всех БД:
-- docker exec redis redis-cli FLUSHALL

-- Просмотр информации о Redis:
-- docker exec redis redis-cli INFO

