DELETE FROM users
WHERE username IN ('admin', 'manager', 'viewer');

DELETE FROM roles
WHERE name IN ('admin', 'manager', 'viewer');