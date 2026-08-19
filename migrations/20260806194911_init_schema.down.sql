-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;
DROP EXTENSION IF EXISTS "uuid-ossp";