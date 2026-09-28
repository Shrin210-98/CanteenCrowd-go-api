-- 1. CREATE REQUESTS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    request_number TEXT NOT NULL,
    request_type TEXT NOT NULL CHECK (
        request_type IN ('task', 'support', 'approval', 'query')
    ),
    category TEXT CHECK (
        category IN ('it', 'hr', 'finance', 'general') OR category IS NULL
    ),
    title TEXT NOT NULL,
    description TEXT,
    
    -- Hierarchy-Aware Requester & Assignee
    requester_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    assignee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    -- Auto-routed to requester's reports_to
    routed_to_manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    
    -- Support Department Routing
    support_department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    support_category TEXT,
    
    -- Priority & Status
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (
        priority IN ('low', 'medium', 'high', 'urgent')
    ),
    status TEXT NOT NULL DEFAULT 'new' CHECK (
        status IN ('new', 'acknowledged', 'in_progress', 'pending', 'completed', 'declined')
    ),
    
    -- Workflow
    due_date DATE,
    completed_at TIMESTAMPTZ,
    
    -- Parent-Child for related requests
    parent_request_id UUID REFERENCES requests(id) ON DELETE SET NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_request_number_per_tenant UNIQUE (tenant_id, request_number)
);

-- Indexes for requests
CREATE INDEX IF NOT EXISTS idx_requests_requester ON requests(requester_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_assignee ON requests(assignee_id, status);
CREATE INDEX IF NOT EXISTS idx_requests_manager ON requests(routed_to_manager_id, status);
CREATE INDEX IF NOT EXISTS idx_requests_tenant ON requests(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_requests_type ON requests(tenant_id, request_type);
CREATE INDEX IF NOT EXISTS idx_requests_parent ON requests(parent_request_id);
CREATE INDEX IF NOT EXISTS idx_requests_support_dept ON requests(support_department_id, status);
CREATE INDEX IF NOT EXISTS idx_requests_support_category ON requests(support_category, status);

-- 2. CREATE REQUEST COMMENTS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS request_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_comments ON request_comments(request_id, created_at);
CREATE INDEX IF NOT EXISTS idx_request_comments_user ON request_comments(user_id);

-- 2.5 CREATE REQUEST CHAT MESSAGES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS request_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for chat messages
CREATE INDEX IF NOT EXISTS idx_chat_messages_request ON request_chat_messages(request_id, created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_sender ON request_chat_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_unread ON request_chat_messages(request_id, is_read) 
WHERE is_read = FALSE;

-- 3. CREATE REQUEST HISTORY TABLE (Audit Trail)
-- ============================================

CREATE TABLE IF NOT EXISTS request_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    changed_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    field_changed TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_history ON request_history(request_id, changed_at);
CREATE INDEX IF NOT EXISTS idx_request_history_user ON request_history(changed_by);

-- 4. CREATE TRIGGERS
-- ============================================

CREATE TRIGGER update_requests_updated_at
BEFORE UPDATE ON requests
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- 5. HELPER FUNCTIONS
-- ============================================

-- Function: Auto-route request to requester's manager
CREATE OR REPLACE FUNCTION auto_route_request_to_manager()
RETURNS TRIGGER AS $$
BEGIN
    -- Set routed_to_manager_id to requester's reports_to if not specified
    IF NEW.routed_to_manager_id IS NULL AND NEW.request_type != 'support' THEN
        SELECT reports_to INTO NEW.routed_to_manager_id
        FROM employees
        WHERE id = NEW.requester_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Auto-route on insert
CREATE TRIGGER auto_route_request
BEFORE INSERT ON requests
FOR EACH ROW
EXECUTE FUNCTION auto_route_request_to_manager();

-- -- Function: Auto-assign support request to department
-- CREATE OR REPLACE FUNCTION auto_assign_support_request()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     -- For support requests, assign to available agent in department
--     IF NEW.request_type = 'support' AND NEW.support_department_id IS NOT NULL THEN
--         -- Find least loaded employee in support department
--         SELECT e.id INTO NEW.assignee_id
--         FROM employees e
--         LEFT JOIN requests r ON r.assignee_id = e.id 
--             AND r.status IN ('new', 'acknowledged', 'in_progress')
--         WHERE e.department_id = NEW.support_department_id
--             AND e.is_active = TRUE
--         GROUP BY e.id
--         ORDER BY COUNT(r.id) ASC
--         LIMIT 1;
--     END IF;
    
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;

---- Trigger for auto-assign support
-- CREATE TRIGGER auto_assign_support_request
-- BEFORE INSERT ON requests
-- FOR EACH ROW
-- EXECUTE FUNCTION auto_assign_support_request();

-- Function: Log chat message
CREATE OR REPLACE FUNCTION log_chat_message()
RETURNS TRIGGER AS $$
BEGIN
    -- Update request's updated_at when chat message is added
    UPDATE requests 
    SET updated_at = NOW()
    WHERE id = NEW.request_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for chat messages
CREATE TRIGGER log_chat_message_trigger
AFTER INSERT ON request_chat_messages
FOR EACH ROW
EXECUTE FUNCTION log_chat_message();

-- Function: Log request status changes
CREATE OR REPLACE FUNCTION log_request_status_change()
RETURNS TRIGGER AS $$
DECLARE
    actor_user_id UUID;
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        SELECT user_id INTO actor_user_id
        FROM employees
        WHERE id = NEW.assignee_id;

        IF actor_user_id IS NOT NULL THEN
            INSERT INTO request_history (
                request_id,
                changed_by,       
                field_changed,
                old_value,
                new_value
            ) VALUES (
                NEW.id,
                actor_user_id,    
                'status',
                OLD.status,
                NEW.status
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Log status changes
CREATE TRIGGER log_request_status_change
AFTER UPDATE OF status ON requests
FOR EACH ROW
EXECUTE FUNCTION log_request_status_change();

-- Function: Log request assignee changes
CREATE OR REPLACE FUNCTION log_request_assignee_change()
RETURNS TRIGGER AS $$
DECLARE
    actor_user_id UUID;
BEGIN
    IF OLD.assignee_id IS DISTINCT FROM NEW.assignee_id THEN
        SELECT user_id INTO actor_user_id
        FROM employees
        WHERE id = NEW.assignee_id;

        IF actor_user_id IS NULL AND OLD.assignee_id IS NOT NULL THEN
            SELECT user_id INTO actor_user_id
            FROM employees
            WHERE id = OLD.assignee_id;
        END IF;

        IF actor_user_id IS NOT NULL THEN
            INSERT INTO request_history (
                request_id,
                changed_by,       
                field_changed,
                old_value,
                new_value
            ) VALUES (
                NEW.id,
                actor_user_id,    
                'assignee',
                OLD.assignee_id::TEXT,
                NEW.assignee_id::TEXT
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Log assignee changes
CREATE TRIGGER log_request_assignee_change
AFTER UPDATE OF assignee_id ON requests
FOR EACH ROW
EXECUTE FUNCTION log_request_assignee_change();

-- 6. VIEWS FOR COMMON QUERIES
-- ============================================

-- View: Requests with employee names
CREATE OR REPLACE VIEW request_details AS
SELECT 
    r.*,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    assign.first_name AS assignee_first_name,
    assign.last_name AS assignee_last_name,
    mgr.first_name AS manager_first_name,
    mgr.last_name AS manager_last_name,
    dept.name AS support_department_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
LEFT JOIN employees mgr ON mgr.id = r.routed_to_manager_id
LEFT JOIN departments dept ON dept.id = r.support_department_id;

-- View: Pending requests for manager
CREATE OR REPLACE VIEW manager_pending_requests AS
SELECT 
    r.*,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name
FROM requests r
JOIN employees req ON req.id = r.requester_id
WHERE r.status IN ('new', 'pending')
  AND (r.routed_to_manager_id IS NOT NULL OR r.support_department_id IS NOT NULL)
ORDER BY r.created_at DESC;

-- View: Support department queue
CREATE OR REPLACE VIEW support_department_queue AS
SELECT 
    r.*,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    assign.first_name AS assignee_first_name,
    assign.last_name AS assignee_last_name,
    dept.name AS department_name,
    COUNT(cm.id) AS chat_message_count
FROM requests r
JOIN employees req ON req.id = r.requester_id
LEFT JOIN employees assign ON assign.id = r.assignee_id
LEFT JOIN departments dept ON dept.id = r.support_department_id
LEFT JOIN request_chat_messages cm ON cm.request_id = r.id
WHERE r.request_type = 'support'
GROUP BY r.id, req.first_name, req.last_name, assign.first_name, assign.last_name, dept.name
ORDER BY 
    CASE r.priority 
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    r.created_at;

-- 7. SEQUENCE FOR REQUEST NUMBERS
-- ============================================

CREATE SEQUENCE IF NOT EXISTS request_number_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Function: Generate next request number
CREATE OR REPLACE FUNCTION generate_request_number(tenant_id UUID)
RETURNS TEXT AS $$
DECLARE
    next_num INT;
    tenant_prefix TEXT;
BEGIN
    -- Get next number for this tenant
    SELECT COALESCE(MAX(CAST(SUBSTRING(request_number FROM '(\d+)$') AS INT)), 0) + 1
    INTO next_num
    FROM requests
    WHERE tenant_id = $1;
    
    -- Format: REQ-0001
    RETURN 'REQ-' || LPAD(next_num::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;