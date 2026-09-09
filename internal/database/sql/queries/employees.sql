-- ============ DEPARTMENT QUERIES ============

-- name: ListDepartments :many
SELECT * FROM departments 
WHERE tenant_id = $1 AND is_active = TRUE
ORDER BY name ASC;

-- name: CreateDepartment :one
INSERT INTO departments (
    tenant_id,
    name,
    description,
    department_type,
    is_system,
    is_active
) VALUES ($1, $2, $3, $4, $5, TRUE)
RETURNING *;

-- name: UpdateDepartment :one
UPDATE departments 
SET 
    name = $2, 
    description = $3,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $4
RETURNING *;

-- name: DeleteDepartment :one
UPDATE departments 
SET is_active = FALSE,
    updated_at = NOW()
WHERE id = $1 
  AND tenant_id = $2 
  AND is_system = FALSE 
RETURNING *;

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

-- name: CountEmployeesInDepartment :one
SELECT COUNT(*) FROM employees 
WHERE department_id = $1 AND tenant_id = $2 AND is_active = TRUE;

-- ============ EMPLOYEE-USER LINK QUERIES ============

-- name: ListEmployeesEligibleForUserCreation :many
SELECT e.* 
FROM employees e
LEFT JOIN users u ON e.user_id = u.id
WHERE e.tenant_id = $1 
  AND e.user_id IS NULL 
  AND e.is_active = TRUE
ORDER BY e.first_name, e.last_name;

-- name: GetEmployeeByUserID :one
SELECT * FROM employees 
WHERE user_id = $1 AND tenant_id = $2 AND is_active = TRUE;

-- name: LinkUserToEmployee :one
UPDATE employees 
SET user_id = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3 AND user_id IS NULL
RETURNING *;

-- name: UnlinkUserFromEmployee :exec
UPDATE employees 
SET user_id = NULL, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: CheckEmployeeHasUser :one
SELECT EXISTS(
    SELECT 1 FROM employees 
    WHERE id = $1 AND tenant_id = $2 AND user_id IS NOT NULL
) AS has_user;

-- name: ListEmployeesWithUserStatus :many
SELECT 
    e.*,
    CASE WHEN e.user_id IS NOT NULL THEN TRUE ELSE FALSE END as has_user_account,
    u.username as user_username,
    u.email as user_email,
    u.user_type as user_type,
    u.is_locked as user_locked
FROM employees e
LEFT JOIN users u ON e.user_id = u.id AND u.deleted_at IS NULL
WHERE e.tenant_id = $1 AND e.is_active = TRUE
ORDER BY e.first_name, e.last_name;

-- ============ EMPLOYEE Reporting Tree QUERIES ============

-- name: GetEmployeeHierarchy :many
WITH RECURSIVE hierarchy AS (
    -- Base: Get all employees with no manager (top level)
    SELECT 
        e.id,
        e.reports_to,
        e.first_name,
        e.last_name,
        p.title as position_title,
        0 as hierarchy_level,
        ARRAY[e.id] as path
    FROM employees e
    LEFT JOIN positions p ON p.id = e.position_id
    WHERE e.tenant_id = $1 
      AND e.reports_to IS NULL
      AND e.is_active = TRUE
    
    UNION ALL
    
    -- Recursive: Get subordinates
    SELECT 
        e.id,
        e.reports_to,
        e.first_name,
        e.last_name,
        p.title as position_title,
        h.hierarchy_level + 1,
        h.path || e.id
    FROM employees e
    JOIN hierarchy h ON e.reports_to = h.id
    LEFT JOIN positions p ON p.id = e.position_id
    WHERE e.tenant_id = $1
      AND e.is_active = TRUE
      AND NOT e.id = ANY(h.path)
)
SELECT * FROM hierarchy ORDER BY path;

-- name: GetDirectReports :many
SELECT 
    e.*,
    p.title as position_title
FROM employees e
LEFT JOIN positions p ON p.id = e.position_id
WHERE e.tenant_id = $1 
  AND e.reports_to = $2
  AND e.is_active = TRUE
ORDER BY e.first_name, e.last_name;

-- name: GetEmployeeWithManager :one
SELECT 
    e.*,
    m.first_name as manager_first_name,
    m.last_name as manager_last_name,
    p.title as position_title
FROM employees e
LEFT JOIN employees m ON m.id = e.reports_to
LEFT JOIN positions p ON p.id = e.position_id
WHERE e.id = $1 AND e.tenant_id = $2;

-- name: UpdateEmployeeReportsTo :one
UPDATE employees 
SET 
    reports_to = $3,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: GetUnassignedEmployees :many
SELECT 
    e.*,
    p.title as position_title
FROM employees e
LEFT JOIN positions p ON p.id = e.position_id
WHERE e.tenant_id = $1 
  AND e.reports_to IS NULL
  AND e.is_active = TRUE
  AND e.id != $2
ORDER BY e.first_name, e.last_name;

-- -- name: IsDescendant :one
-- SELECT EXISTS (
--     SELECT 1
--     FROM employees e
--     WHERE e.id = $1
--       AND e.id IN (
--           WITH RECURSIVE ancestors AS (
--               SELECT id, reports_to
--               FROM employees
--               WHERE id = $2
              
--               UNION ALL
              
--               SELECT e.id, e.reports_to
--               FROM employees e
--               JOIN ancestors ON e.id = ancestors.reports_to
--           )
--           SELECT id FROM ancestors
--       )
-- ) AS is_ancestor;

-- ============ Support Chat QUERIES ============

-- name: GetSupportDepartments :many
SELECT * FROM departments 
WHERE tenant_id = $1 
  AND department_type = 'support'
  AND is_active = TRUE
ORDER BY name;

-- name: GetSupportTeamMembers :many
SELECT e.*
FROM employees e
WHERE e.tenant_id = $1 
  AND e.department_id = $2
  AND e.is_active = TRUE
ORDER BY e.first_name, e.last_name;

-- name: GetAvailableSupportAgent :one
SELECT e.*
FROM employees e
LEFT JOIN requests r ON r.assignee_id = e.id 
  AND r.status IN ('new', 'acknowledged', 'in_progress')
WHERE e.department_id = $1
  AND e.tenant_id = $2
  AND e.is_active = TRUE
GROUP BY e.id
ORDER BY COUNT(r.id) ASC
LIMIT 1;

-- name: GetSupportDepartmentByCategory :one
SELECT * FROM departments 
WHERE tenant_id = $1 
  AND department_type = 'support'
  AND name ILIKE '%' || sqlc.arg(category) || '%'
  AND is_active = TRUE
LIMIT 1;