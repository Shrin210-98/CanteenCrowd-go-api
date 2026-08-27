-- ============ DEPARTMENT QUERIES ============

-- name: ListDepartments :many
SELECT * FROM departments 
WHERE tenant_id = $1 AND is_active = TRUE
ORDER BY name ASC;

-- name: CreateDepartment :one
INSERT INTO departments (tenant_id, name, description) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: UpdateDepartment :one
UPDATE departments 
SET 
    name = $2, 
    description = $3,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $4
RETURNING *;

-- name: DeleteDepartment :exec
DELETE FROM departments 
WHERE id = $1 AND tenant_id = $2;

-- name: SoftDeleteDepartment :exec
UPDATE departments 
SET 
    is_active = FALSE,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: CountDepartments :one
SELECT COUNT(*) FROM departments 
WHERE tenant_id = $1 AND is_active = TRUE;

-- name: GetDepartmentByID :one
SELECT * FROM departments 
WHERE id = $1 AND tenant_id = $2;

-- ============ POSITION QUERIES ============

-- name: ListPositions :many
SELECT * FROM positions 
WHERE tenant_id = $1 AND is_active = TRUE
ORDER BY title ASC;

-- name: CreatePosition :one
INSERT INTO positions (tenant_id, title, description, level, is_management) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;

-- name: UpdatePosition :one
UPDATE positions 
SET 
    title = $2, 
    description = $3, 
    level = $4, 
    is_management = $5,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $6
RETURNING *;

-- name: DeletePosition :exec
DELETE FROM positions 
WHERE id = $1 AND tenant_id = $2;

-- name: SoftDeletePosition :exec
UPDATE positions 
SET 
    is_active = FALSE,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: CountPositions :one
SELECT COUNT(*) FROM positions 
WHERE tenant_id = $1 AND is_active = TRUE;

-- name: GetPosition :one
SELECT * FROM positions 
WHERE id = $1 AND tenant_id = $2;

-- ============ EMPLOYEE QUERIES ============

-- name: ListEmployees :many
SELECT * FROM employees 
WHERE tenant_id = $1 AND is_active = TRUE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateEmployee :one
INSERT INTO employees (
    tenant_id,
    user_id,
    first_name, 
    last_name, 
    email, 
    phone, 
    address, 
    hire_date, 
    termination_date, 
    position_id, 
    department_id, 
    salary, 
    emergency_contact_name, 
    emergency_contact_phone, 
    profile_description, 
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET 
    user_id = $2,
    first_name = $3,
    last_name = $4,
    email = $5,
    phone = $6,
    address = $7,
    hire_date = $8,
    termination_date = $9,
    position_id = $10,
    department_id = $11,
    salary = $12,
    emergency_contact_name = $13,
    emergency_contact_phone = $14,
    profile_description = $15,
    is_active = $16,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $17
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees 
WHERE id = $1 AND tenant_id = $2;

-- name: SoftDeleteEmployee :one
UPDATE employees 
SET 
    is_active = FALSE,
    termination_date = $2,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees 
WHERE tenant_id = $1 AND is_active = TRUE;

-- name: GetEmployeeById :one
SELECT * FROM employees 
WHERE id = $1 AND tenant_id = $2 
LIMIT 1;

-- name: SearchEmployees :many
SELECT 
    e.*,
    d.name as department_name,
    p.title as position_title
FROM employees e
JOIN departments d ON e.department_id = d.id
JOIN positions p ON e.position_id = p.id
WHERE 
    e.tenant_id = @tenant_id
AND
    (e.first_name ILIKE '%' || @search::text || '%' 
     OR e.last_name ILIKE '%' || @search::text || '%' 
     OR e.email ILIKE '%' || @search::text || '%'
     OR @search::text = '')
AND
    (e.department_id = @department_id OR @department_id IS NULL)
AND
    (e.position_id = @position_id OR @position_id IS NULL)
AND
    (e.is_active = @is_active OR @is_active IS NULL)
AND
    (e.hire_date >= @hire_date_start OR @hire_date_start IS NULL)
AND
    (e.hire_date <= @hire_date_end OR @hire_date_end IS NULL)
ORDER BY e.created_at DESC
LIMIT @limit_count OFFSET @offset_count;

-- name: CountSearchEmployees :one
SELECT COUNT(*) FROM employees e
WHERE 
    e.tenant_id = @tenant_id
AND
    (e.first_name ILIKE '%' || @search::text || '%' 
     OR e.last_name ILIKE '%' || @search::text || '%' 
     OR e.email ILIKE '%' || @search::text || '%'
     OR @search::text = '')
AND
    (e.department_id = @department_id OR @department_id IS NULL)
AND
    (e.position_id = @position_id OR @position_id IS NULL)
AND
    (e.is_active = @is_active OR @is_active IS NULL)
AND
    (e.hire_date >= @hire_date_start OR @hire_date_start IS NULL)
AND
    (e.hire_date <= @hire_date_end OR @hire_date_end IS NULL);