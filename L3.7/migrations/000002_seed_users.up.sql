INSERT INTO roles (name, description)
VALUES
    ('admin', 'Полный доступ к системе'),
    ('manager', 'Просмотр и редактирование товаров'),
    ('viewer', 'Только просмотр товаров')
ON CONFLICT (name) DO NOTHING;

INSERT INTO users (username, role_id)
SELECT 'admin', id
FROM roles
WHERE name = 'admin'
ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, role_id)
SELECT 'manager', id
FROM roles
WHERE name = 'manager'
ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, role_id)
SELECT 'viewer', id
FROM roles
WHERE name = 'viewer'
ON CONFLICT (username) DO NOTHING;