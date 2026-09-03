ALTER TABLE departments DROP CONSTRAINT IF EXISTS fk_departments_manager;

DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
DROP TRIGGER IF EXISTS update_positions_updated_at ON positions;
DROP TRIGGER IF EXISTS update_departments_updated_at ON departments;

DROP INDEX IF EXISTS idx_employees_active;
DROP INDEX IF EXISTS idx_employees_position;
DROP INDEX IF EXISTS idx_employees_department;
DROP INDEX IF EXISTS idx_employees_tenant;
DROP INDEX IF EXISTS idx_positions_tenant;
DROP INDEX IF EXISTS idx_departments_tenant;

DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS departments;

