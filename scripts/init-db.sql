CREATE TABLE IF NOT EXISTS bikes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    location VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    bike_id UUID REFERENCES bikes(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
);

-- Insert sample bikes
INSERT INTO bikes (name, status, location) VALUES
    ('Bike 1', 'available', 'Location A'),
    ('Bike 2', 'available', 'Location A'),
    ('Bike 3', 'available', 'Location B'),
    ('Bike 4', 'available', 'Location B'),
    ('Bike 5', 'available', 'Location C')
ON CONFLICT DO NOTHING;

