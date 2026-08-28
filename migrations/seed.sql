-- Seed users for development
-- 1. Admin: rogab@admin.com / r0g4b@2026!
-- 2. Motorista (Driver): ricberns@gmail.com / 1254101254@Abc

INSERT INTO users (name, lastname, phone, type, email, document, password_hash)
VALUES 
    (
        'Rogab',
        'Admin',
        '11999999999',
        'admin',
        'rogab@admin.com',
        '00000000000',
        crypt('r0g4b@2026!', gen_salt('bf', 10))
    ),
    (
        'Ricardo',
        'Berns',
        '11988888888',
        'driver',
        'ricberns@gmail.com',
        '11111111111',
        crypt('1254101254@Abc', gen_salt('bf', 10))
    )
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    lastname = EXCLUDED.lastname,
    phone = EXCLUDED.phone,
    type = EXCLUDED.type,
    document = EXCLUDED.document,
    password_hash = EXCLUDED.password_hash,
    updated_at = now();

