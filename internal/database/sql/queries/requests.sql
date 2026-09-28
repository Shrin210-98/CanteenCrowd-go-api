-- name: CreateRequest :one
INSERT INTO requests (
    tenant_id,
    request_number,
    request_type,
    category,
    title,
    description,
    requester_id,
    assignee_id,
    routed_to_manager_id,
    support_department_id,
    support_category,
    priority,
    status,
    due_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'new', $13
) RETURNING *;

-- name: GetRequestByID :one
SELECT 
    r.*,
    req.first_name as requester_first_name,
    req.last_name as requester_last_name,
    assign.first_name as assignee_first_name,
    assign.last_name as assignee_last_name,
    mgr.first_name as manager_first_name,
    mgr.last_name as manager_last_name,
    dept.name as support_department_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
LEFT JOIN employees mgr ON mgr.id = r.routed_to_manager_id
LEFT JOIN departments dept ON dept.id = r.support_department_id
WHERE r.id = $1 AND r.tenant_id = $2;

-- name: GetRequestsByAssignee :many
SELECT 
    r.*,
    req.first_name as requester_first_name,
    req.last_name as requester_last_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
WHERE r.tenant_id = $1 
  AND r.assignee_id = $2
  AND ($3 = '' OR r.status = $3)
  AND ($4 = '' OR r.request_type = $4)
ORDER BY r.created_at DESC
LIMIT $5 OFFSET $6;

-- name: GetRequestsByRequester :many
SELECT 
    r.*,
    assign.first_name as assignee_first_name,
    assign.last_name as assignee_last_name
FROM requests r
LEFT JOIN employees assign ON assign.id = r.assignee_id
WHERE r.tenant_id = sqlc.arg(tenant_id)
  AND r.requester_id = sqlc.arg(requester_id)
  AND (sqlc.arg(status)::text = '' OR r.status = sqlc.arg(status)::text)
  AND (sqlc.arg(request_type)::text = '' OR r.request_type = sqlc.arg(request_type)::text)
ORDER BY r.created_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: GetRequestsForManager :many
SELECT 
    r.*,
    req.first_name as requester_first_name,
    req.last_name as requester_last_name,
    assign.first_name as assignee_first_name,
    assign.last_name as assignee_last_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
WHERE r.tenant_id = $1 
  AND r.routed_to_manager_id = $2
  AND ($3 = '' OR r.status = $3)
ORDER BY r.created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetDepartmentRequests :many
SELECT 
    r.*,
    req.first_name AS requester_first_name,
    req.last_name  AS requester_last_name,
    assign.first_name AS assignee_first_name,
    assign.last_name  AS assignee_last_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
WHERE r.tenant_id = sqlc.arg(tenant_id)
  AND r.support_department_id = sqlc.arg(support_department_id)
  AND (
      sqlc.arg(filter)::text = ''
      OR (sqlc.arg(filter) = 'unassigned' AND r.assignee_id IS NULL)
      OR (sqlc.arg(filter) = 'mine'       AND r.assignee_id = sqlc.arg(employee_id))
      OR (sqlc.arg(filter) = 'team'       AND r.assignee_id IS NOT NULL AND r.assignee_id != sqlc.arg(employee_id))
      OR (sqlc.arg(filter) = 'all'        AND TRUE)
      OR (sqlc.arg(filter) = 'resolved'   AND r.status IN ('completed', 'declined'))
  )
  AND (sqlc.arg(status)::text = '' OR r.status = sqlc.arg(status)::text)
  AND (sqlc.arg(priority)::text = '' OR r.priority = sqlc.arg(priority)::text)
  AND (
      sqlc.arg(search)::text = ''
      OR r.title ILIKE '%' || sqlc.arg(search) || '%'
      OR r.description ILIKE '%' || sqlc.arg(search) || '%'
      OR r.request_number ILIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY 
    CASE r.priority 
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    r.created_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: GetDepartmentRequestCounts :one
SELECT 
    COUNT(*) FILTER (WHERE assignee_id IS NULL) AS unassigned_count,
    COUNT(*) FILTER (WHERE assignee_id = $3) AS mine_count,
    COUNT(*) FILTER (
        WHERE assignee_id IS NOT NULL 
          AND assignee_id != $3
          AND status NOT IN ('completed', 'declined')
    ) AS team_count,
    COUNT(*) FILTER (
        WHERE status NOT IN ('completed', 'declined')
    ) AS all_count
    FROM requests
    WHERE tenant_id = $1 
    AND support_department_id = $2;

-- name: UpdateRequestStatus :one
UPDATE requests 
SET 
    status = $3,
    updated_at = NOW(),
    completed_at = CASE WHEN $3 = 'completed' THEN NOW() ELSE completed_at END
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: UpdateRequestAssignee :one
UPDATE requests 
SET 
    assignee_id = $3,
    status = 'acknowledged',
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: AddRequestComment :one
INSERT INTO request_comments (
    request_id,
    user_id,
    comment
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetRequestComments :many
SELECT 
    rc.*,
    e.first_name,
    e.last_name
FROM request_comments rc
JOIN users u ON u.id = rc.user_id
LEFT JOIN employees e ON e.user_id = u.id
WHERE rc.request_id = $1
ORDER BY rc.created_at;

-- Chat Messages
-- name: CreateChatMessage :one
INSERT INTO request_chat_messages (
    request_id,
    sender_id,
    message
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetChatMessages :many
SELECT 
    cm.*,
    u.username,
    e.first_name,
    e.last_name
FROM request_chat_messages cm
JOIN users u ON u.id = cm.sender_id
LEFT JOIN employees e ON e.user_id = u.id
WHERE cm.request_id = $1
ORDER BY cm.created_at;

-- name: MarkChatMessagesRead :exec
UPDATE request_chat_messages
SET is_read = TRUE,
    read_at = NOW()
WHERE request_id = $1
  AND sender_id != $2
  AND is_read = FALSE;

-- name: GetUnreadChatCount :one
SELECT COUNT(*)
FROM request_chat_messages
WHERE request_id = $1
  AND sender_id != $2
  AND is_read = FALSE;

-- name: GetTeamRequests :many
SELECT 
    r.*,
    req.first_name as requester_first_name,
    req.last_name as requester_last_name,
    assign.first_name as assignee_first_name,
    assign.last_name as assignee_last_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
WHERE r.tenant_id = $1 
  AND req.reports_to = $2
  AND ($3 = '' OR r.status = $3)
ORDER BY r.created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetNextRequestNumber :one
SELECT COALESCE(MAX(CAST(SUBSTRING(request_number FROM '(\d+)$') AS INT)), 0) + 1
FROM requests
WHERE tenant_id = $1;

-- name: GetDepartmentMembers :many
SELECT 
    e.id,
    e.first_name,
    e.last_name,
    e.email,
    e.position_id,
    p.title AS position_title
FROM employees e
LEFT JOIN positions p ON p.id = e.position_id
WHERE e.tenant_id = sqlc.arg(tenant_id)
  AND e.department_id = sqlc.arg(department_id)
  AND e.is_active = TRUE
ORDER BY e.first_name, e.last_name;

-- name: GetAllTenantSupportRequests :many
SELECT 
    r.*,
    req.first_name AS requester_first_name,
    req.last_name  AS requester_last_name,
    assign.first_name AS assignee_first_name,
    assign.last_name  AS assignee_last_name,
    dept.name AS support_department_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
LEFT JOIN departments dept ON dept.id = r.support_department_id
WHERE r.tenant_id = sqlc.arg(tenant_id)
  AND r.request_type = 'support'
  AND (
      sqlc.arg(filter)::text = ''
      OR (sqlc.arg(filter) = 'unassigned' AND r.assignee_id IS NULL)
      OR (sqlc.arg(filter) = 'mine' 
          AND r.assignee_id IS NOT NULL
          AND r.assignee_id IS NOT DISTINCT FROM sqlc.narg(employee_id))
      OR (sqlc.arg(filter) = 'team' 
          AND r.assignee_id IS NOT NULL
          AND r.assignee_id IS DISTINCT FROM sqlc.narg(employee_id)
          AND r.status NOT IN ('completed', 'declined'))
      OR (sqlc.arg(filter) = 'all' 
          AND TRUE)
      OR (sqlc.arg(filter) = 'resolved' 
          AND r.status IN ('completed', 'declined'))
  )
  AND (sqlc.arg(status)::text = '' OR r.status = sqlc.arg(status)::text)
  AND (sqlc.arg(priority)::text = '' OR r.priority = sqlc.arg(priority)::text)
  AND (
      sqlc.arg(search)::text = ''
      OR r.title ILIKE '%' || sqlc.arg(search) || '%'
      OR r.description ILIKE '%' || sqlc.arg(search) || '%'
      OR r.request_number ILIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY 
    CASE r.priority 
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    r.created_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: GetTenantSupportRequestCounts :one
SELECT 
    COUNT(*) FILTER (WHERE assignee_id IS NULL) AS unassigned_count,
    COUNT(*) FILTER (
        WHERE assignee_id IS NOT NULL 
          AND assignee_id IS NOT DISTINCT FROM sqlc.narg(employee_id)
    ) AS mine_count,
    COUNT(*) FILTER (
        WHERE assignee_id IS NOT NULL 
          AND assignee_id IS DISTINCT FROM sqlc.narg(employee_id)
          AND status NOT IN ('completed', 'declined')
    ) AS team_count,
    COUNT(*) AS all_count
FROM requests
WHERE tenant_id = sqlc.arg(tenant_id)
  AND request_type = 'support';