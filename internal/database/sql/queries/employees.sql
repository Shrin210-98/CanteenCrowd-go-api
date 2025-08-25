-- name: ListDepartments :many
SELECT * FROM departments;

-- name: CreateDepartment :one
INSERT INTO departments (name, description) VALUES ($1, $2) RETURNING *;

-- name: UpdateDepartment :exec
UPDATE departments set name = $2, description = $3 WHERE id = $1;

-- name: DeleteDepartment :exec
DELETE FROM departments WHERE id = $1;

-- name: CountDepartments :one
SELECT COUNT(*) FROM departments;




-- name: ListPositions :many
SELECT * FROM positions;

-- name: CreatePosition :one
INSERT INTO positions ( title, description, level, is_management) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdatePosition :exec
UPDATE positions set  title = $1, description = $2, level = $3, is_management = $4 WHERE id = $1;

-- name: DeletePosition :exec
DELETE FROM positions WHERE id = $1;

-- name: CountPositions :one
SELECT COUNT(*) FROM positions;




-- name: ListEmployees :many
SELECT * FROM employees ORDER BY id LIMIT $1 OFFSET $2;

-- name: CreateEmployee :one
INSERT INTO employees (
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET 
    first_name = $2,
    last_name = $3,
    email = $4,
    phone = $5,
    address = $6,
    hire_date = $7,
    termination_date = $8,
    position_id = $9,
    department_id = $10,
    salary = $11,
    emergency_contact_name = $12,
    emergency_contact_phone = $13,
    profile_description = $14,
    is_active = $15,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees WHERE id = $1;
  
-- name: CountEmployees :one
SELECT COUNT(*) FROM employees;

-- name: GetEmployeeById :one
SELECT * FROM employees WHERE id = $1 LIMIT 1;

-- name: SearchEmployees :many
SELECT * FROM employees
WHERE 
  (first_name ILIKE '%' || $1 || '%' OR last_name ILIKE '%' || $1 || '%' OR $1 = '')
AND
  (department_id = $2 OR $2 IS NULL)
AND
  (position_id = $3 OR $3 IS NULL)
AND
  (is_active = $4 OR $4 IS NULL)
AND
  (hire_date BETWEEN $5 AND $6 OR ($5 IS NULL AND $6 IS NULL))
ORDER BY id
LIMIT $7 OFFSET $8;
 
-- name: CountSearchEmployees :one
SELECT COUNT(*) FROM employees
WHERE 
  (first_name ILIKE '%' || $1 || '%' OR last_name ILIKE '%' || $1 || '%' OR $1 = '')
AND
  (department_id = $2 OR $2 IS NULL)
AND
  (position_id = $3 OR $3 IS NULL)
AND
  (is_active = $4 OR $4 IS NULL)
AND
  (hire_date BETWEEN $5 AND $6 OR ($5 IS NULL AND $6 IS NULL));
