-- Create sequence "request_number_seq"
CREATE SEQUENCE "public"."request_number_seq";
-- Create "update_updated_at_column" function
CREATE FUNCTION "public"."update_updated_at_column" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;
-- Create "tenants" table
CREATE TABLE "public"."tenants" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "plan" text NOT NULL DEFAULT 'free',
  "status" text NOT NULL DEFAULT 'active',
  "settings" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "tenants_slug_key" UNIQUE ("slug")
);
-- Create "departments" table
CREATE TABLE "public"."departments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tenant_id" uuid NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "department_type" text NOT NULL DEFAULT 'operational',
  "is_active" boolean NOT NULL DEFAULT true,
  "is_system" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "unique_tenant_department" UNIQUE ("tenant_id", "name"),
  CONSTRAINT "departments_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_departments_tenant" to table: "departments"
CREATE INDEX "idx_departments_tenant" ON "public"."departments" ("tenant_id");
-- Create trigger "update_departments_updated_at"
CREATE TRIGGER "update_departments_updated_at" BEFORE UPDATE ON "public"."departments" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "positions" table
CREATE TABLE "public"."positions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tenant_id" uuid NOT NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "level" integer NULL,
  "is_management" boolean NOT NULL DEFAULT false,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "unique_tenant_position" UNIQUE ("tenant_id", "title"),
  CONSTRAINT "positions_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_positions_tenant" to table: "positions"
CREATE INDEX "idx_positions_tenant" ON "public"."positions" ("tenant_id");
-- Create "roles" table
CREATE TABLE "public"."roles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tenant_id" uuid NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "permissions" jsonb NOT NULL DEFAULT '[]',
  "is_default" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "unique_role_per_tenant" UNIQUE ("tenant_id", "name"),
  CONSTRAINT "fk_roles_tenant" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_roles_tenant" to table: "roles"
CREATE INDEX "idx_roles_tenant" ON "public"."roles" ("tenant_id");
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tenant_id" uuid NOT NULL,
  "username" text NOT NULL,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "last_login" timestamptz NULL,
  "login_attempts" integer NOT NULL DEFAULT 0,
  "is_locked" boolean NOT NULL DEFAULT false,
  "locked_until" timestamptz NULL,
  "password_reset_token" text NULL,
  "token_expiry" timestamptz NULL,
  "must_change_password" boolean NOT NULL DEFAULT false,
  "email_verified" boolean NOT NULL DEFAULT false,
  "mfa_enabled" boolean NOT NULL DEFAULT false,
  "last_password_change" timestamptz NULL,
  "user_type" text NOT NULL,
  "role_id" uuid NULL,
  "permissions_override" jsonb NULL,
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_tenant" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "users_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "users_user_type_check" CHECK (user_type = ANY (ARRAY['tenant_owner'::text, 'staff'::text, 'guest'::text]))
);
-- Create index "idx_users_email" to table: "users"
CREATE INDEX "idx_users_email" ON "public"."users" ("email") WHERE (deleted_at IS NULL);
-- Create index "idx_users_role_id" to table: "users"
CREATE INDEX "idx_users_role_id" ON "public"."users" ("role_id") WHERE (deleted_at IS NULL);
-- Create index "idx_users_tenant_id" to table: "users"
CREATE INDEX "idx_users_tenant_id" ON "public"."users" ("tenant_id");
-- Create index "idx_users_tenant_type" to table: "users"
CREATE INDEX "idx_users_tenant_type" ON "public"."users" ("tenant_id", "user_type");
-- Create index "users_tenant_email_unique" to table: "users"
CREATE UNIQUE INDEX "users_tenant_email_unique" ON "public"."users" ("tenant_id", "email") WHERE (deleted_at IS NULL);
-- Create index "users_tenant_username_unique" to table: "users"
CREATE UNIQUE INDEX "users_tenant_username_unique" ON "public"."users" ("tenant_id", "username") WHERE (deleted_at IS NULL);
-- Create "employees" table
CREATE TABLE "public"."employees" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tenant_id" uuid NOT NULL,
  "user_id" uuid NULL,
  "first_name" text NOT NULL,
  "last_name" text NOT NULL,
  "email" text NOT NULL,
  "phone" text NULL,
  "address" text NULL,
  "hire_date" date NOT NULL,
  "termination_date" date NULL,
  "position_id" uuid NOT NULL,
  "department_id" uuid NOT NULL,
  "reports_to" uuid NULL,
  "salary" numeric(10,2) NULL,
  "emergency_contact_name" text NULL,
  "emergency_contact_phone" text NULL,
  "profile_description" text NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "employees_user_id_key" UNIQUE ("user_id"),
  CONSTRAINT "unique_tenant_employee_email" UNIQUE ("tenant_id", "email"),
  CONSTRAINT "employees_department_id_fkey" FOREIGN KEY ("department_id") REFERENCES "public"."departments" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "employees_position_id_fkey" FOREIGN KEY ("position_id") REFERENCES "public"."positions" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "employees_reports_to_fkey" FOREIGN KEY ("reports_to") REFERENCES "public"."employees" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "employees_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "employees_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_employees_active" to table: "employees"
CREATE INDEX "idx_employees_active" ON "public"."employees" ("tenant_id", "is_active");
-- Create index "idx_employees_department" to table: "employees"
CREATE INDEX "idx_employees_department" ON "public"."employees" ("department_id");
-- Create index "idx_employees_position" to table: "employees"
CREATE INDEX "idx_employees_position" ON "public"."employees" ("position_id");
-- Create index "idx_employees_tenant" to table: "employees"
CREATE INDEX "idx_employees_tenant" ON "public"."employees" ("tenant_id");
-- Create trigger "update_employees_updated_at"
CREATE TRIGGER "update_employees_updated_at" BEFORE UPDATE ON "public"."employees" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "menu_items" table
CREATE TABLE "public"."menu_items" (
  "id" serial NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "price" numeric(10,2) NOT NULL,
  "cost" numeric(10,2) NULL,
  "preparation_time" integer NULL,
  "is_vegetarian" boolean NULL DEFAULT false,
  "is_vegan" boolean NULL DEFAULT false,
  "is_gluten_free" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "image_url" text NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create "menus" table
CREATE TABLE "public"."menus" (
  "id" serial NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "categories" jsonb NOT NULL DEFAULT '[]',
  "valid_from" timestamp NULL,
  "valid_to" timestamp NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "menus_name_key" UNIQUE ("name")
);
-- Create trigger "update_positions_updated_at"
CREATE TRIGGER "update_positions_updated_at" BEFORE UPDATE ON "public"."positions" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "requests" table
CREATE TABLE "public"."requests" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tenant_id" uuid NOT NULL,
  "request_number" text NOT NULL,
  "request_type" text NOT NULL,
  "category" text NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "requester_id" uuid NOT NULL,
  "assignee_id" uuid NULL,
  "routed_to_manager_id" uuid NULL,
  "support_department_id" uuid NULL,
  "support_category" text NULL,
  "priority" text NOT NULL DEFAULT 'medium',
  "status" text NOT NULL DEFAULT 'new',
  "due_date" date NULL,
  "completed_at" timestamptz NULL,
  "parent_request_id" uuid NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "unique_request_number_per_tenant" UNIQUE ("tenant_id", "request_number"),
  CONSTRAINT "requests_assignee_id_fkey" FOREIGN KEY ("assignee_id") REFERENCES "public"."employees" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "requests_parent_request_id_fkey" FOREIGN KEY ("parent_request_id") REFERENCES "public"."requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "requests_requester_id_fkey" FOREIGN KEY ("requester_id") REFERENCES "public"."employees" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "requests_routed_to_manager_id_fkey" FOREIGN KEY ("routed_to_manager_id") REFERENCES "public"."employees" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "requests_support_department_id_fkey" FOREIGN KEY ("support_department_id") REFERENCES "public"."departments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "requests_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "requests_category_check" CHECK ((category = ANY (ARRAY['it'::text, 'hr'::text, 'finance'::text, 'general'::text])) OR (category IS NULL)),
  CONSTRAINT "requests_priority_check" CHECK (priority = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'urgent'::text])),
  CONSTRAINT "requests_request_type_check" CHECK (request_type = ANY (ARRAY['task'::text, 'support'::text, 'approval'::text, 'query'::text])),
  CONSTRAINT "requests_status_check" CHECK (status = ANY (ARRAY['new'::text, 'acknowledged'::text, 'in_progress'::text, 'pending'::text, 'completed'::text, 'declined'::text]))
);
-- Create index "idx_requests_assignee" to table: "requests"
CREATE INDEX "idx_requests_assignee" ON "public"."requests" ("assignee_id", "status");
-- Create index "idx_requests_manager" to table: "requests"
CREATE INDEX "idx_requests_manager" ON "public"."requests" ("routed_to_manager_id", "status");
-- Create index "idx_requests_parent" to table: "requests"
CREATE INDEX "idx_requests_parent" ON "public"."requests" ("parent_request_id");
-- Create index "idx_requests_requester" to table: "requests"
CREATE INDEX "idx_requests_requester" ON "public"."requests" ("requester_id", "created_at" DESC);
-- Create index "idx_requests_status" to table: "requests"
CREATE INDEX "idx_requests_status" ON "public"."requests" ("tenant_id", "status");
-- Create index "idx_requests_support_category" to table: "requests"
CREATE INDEX "idx_requests_support_category" ON "public"."requests" ("support_category", "status");
-- Create index "idx_requests_support_dept" to table: "requests"
CREATE INDEX "idx_requests_support_dept" ON "public"."requests" ("support_department_id", "status");
-- Create index "idx_requests_tenant" to table: "requests"
CREATE INDEX "idx_requests_tenant" ON "public"."requests" ("tenant_id", "created_at" DESC);
-- Create index "idx_requests_type" to table: "requests"
CREATE INDEX "idx_requests_type" ON "public"."requests" ("tenant_id", "request_type");
-- Create "log_chat_message" function
CREATE FUNCTION "public"."log_chat_message" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    -- Update request's updated_at when chat message is added
    UPDATE requests 
    SET updated_at = NOW()
    WHERE id = NEW.request_id;
    
    RETURN NEW;
END;
$$;
-- Create "request_chat_messages" table
CREATE TABLE "public"."request_chat_messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "request_id" uuid NOT NULL,
  "sender_id" uuid NOT NULL,
  "message" text NOT NULL,
  "is_read" boolean NOT NULL DEFAULT false,
  "read_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "request_chat_messages_request_id_fkey" FOREIGN KEY ("request_id") REFERENCES "public"."requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "request_chat_messages_sender_id_fkey" FOREIGN KEY ("sender_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_chat_messages_request" to table: "request_chat_messages"
CREATE INDEX "idx_chat_messages_request" ON "public"."request_chat_messages" ("request_id", "created_at");
-- Create index "idx_chat_messages_sender" to table: "request_chat_messages"
CREATE INDEX "idx_chat_messages_sender" ON "public"."request_chat_messages" ("sender_id");
-- Create index "idx_chat_messages_unread" to table: "request_chat_messages"
CREATE INDEX "idx_chat_messages_unread" ON "public"."request_chat_messages" ("request_id", "is_read") WHERE (is_read = false);
-- Create trigger "log_chat_message_trigger"
CREATE TRIGGER "log_chat_message_trigger" AFTER INSERT ON "public"."request_chat_messages" FOR EACH ROW EXECUTE FUNCTION "public"."log_chat_message"();
-- Create "auto_assign_support_request" function
CREATE FUNCTION "public"."auto_assign_support_request" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    -- For support requests, assign to available agent in department
    IF NEW.request_type = 'support' AND NEW.support_department_id IS NOT NULL THEN
        -- Find least loaded employee in support department
        SELECT e.id INTO NEW.assignee_id
        FROM employees e
        LEFT JOIN requests r ON r.assignee_id = e.id 
            AND r.status IN ('new', 'acknowledged', 'in_progress')
        WHERE e.department_id = NEW.support_department_id
            AND e.is_active = TRUE
        GROUP BY e.id
        ORDER BY COUNT(r.id) ASC
        LIMIT 1;
    END IF;
    
    RETURN NEW;
END;
$$;
-- Create trigger "auto_assign_support_request"
CREATE TRIGGER "auto_assign_support_request" BEFORE INSERT ON "public"."requests" FOR EACH ROW EXECUTE FUNCTION "public"."auto_assign_support_request"();
-- Create "auto_route_request_to_manager" function
CREATE FUNCTION "public"."auto_route_request_to_manager" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    -- Set routed_to_manager_id to requester's reports_to if not specified
    IF NEW.routed_to_manager_id IS NULL AND NEW.request_type != 'support' THEN
        SELECT reports_to INTO NEW.routed_to_manager_id
        FROM employees
        WHERE id = NEW.requester_id;
    END IF;
    
    RETURN NEW;
END;
$$;
-- Create trigger "auto_route_request"
CREATE TRIGGER "auto_route_request" BEFORE INSERT ON "public"."requests" FOR EACH ROW EXECUTE FUNCTION "public"."auto_route_request_to_manager"();
-- Create "request_history" table
CREATE TABLE "public"."request_history" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "request_id" uuid NOT NULL,
  "changed_by" uuid NOT NULL,
  "field_changed" text NOT NULL,
  "old_value" text NULL,
  "new_value" text NULL,
  "changed_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "request_history_changed_by_fkey" FOREIGN KEY ("changed_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "request_history_request_id_fkey" FOREIGN KEY ("request_id") REFERENCES "public"."requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_request_history" to table: "request_history"
CREATE INDEX "idx_request_history" ON "public"."request_history" ("request_id", "changed_at");
-- Create index "idx_request_history_user" to table: "request_history"
CREATE INDEX "idx_request_history_user" ON "public"."request_history" ("changed_by");
-- Create "log_request_assignee_change" function
CREATE FUNCTION "public"."log_request_assignee_change" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.assignee_id IS DISTINCT FROM NEW.assignee_id THEN
        INSERT INTO request_history (
            request_id,
            changed_by,
            field_changed,
            old_value,
            new_value
        ) VALUES (
            NEW.id,
            NEW.assignee_id,
            'assignee',
            OLD.assignee_id::TEXT,
            NEW.assignee_id::TEXT
        );
    END IF;
    
    RETURN NEW;
END;
$$;
-- Create trigger "log_request_assignee_change"
CREATE TRIGGER "log_request_assignee_change" AFTER UPDATE OF "assignee_id" ON "public"."requests" FOR EACH ROW EXECUTE FUNCTION "public"."log_request_assignee_change"();
-- Create "log_request_status_change" function
CREATE FUNCTION "public"."log_request_status_change" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO request_history (
            request_id,
            changed_by,
            field_changed,
            old_value,
            new_value
        ) VALUES (
            NEW.id,
            NEW.assignee_id,
            'status',
            OLD.status,
            NEW.status
        );
    END IF;
    
    RETURN NEW;
END;
$$;
-- Create trigger "log_request_status_change"
CREATE TRIGGER "log_request_status_change" AFTER UPDATE OF "status" ON "public"."requests" FOR EACH ROW EXECUTE FUNCTION "public"."log_request_status_change"();
-- Create trigger "update_requests_updated_at"
CREATE TRIGGER "update_requests_updated_at" BEFORE UPDATE ON "public"."requests" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_roles_updated_at"
CREATE TRIGGER "update_roles_updated_at" BEFORE UPDATE ON "public"."roles" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_tenants_updated_at"
CREATE TRIGGER "update_tenants_updated_at" BEFORE UPDATE ON "public"."tenants" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "user_profiles" table
CREATE TABLE "public"."user_profiles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "tenant_id" uuid NOT NULL,
  "full_name" text NULL,
  "phone" text NULL,
  "avatar_url" text NULL,
  "timezone" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_profiles_user_id_key" UNIQUE ("user_id"),
  CONSTRAINT "fk_user_profiles_tenant" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_profiles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_user_profiles_tenant" to table: "user_profiles"
CREATE INDEX "idx_user_profiles_tenant" ON "public"."user_profiles" ("tenant_id");
-- Create trigger "update_user_profiles_updated_at"
CREATE TRIGGER "update_user_profiles_updated_at" BEFORE UPDATE ON "public"."user_profiles" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_users_updated_at"
CREATE TRIGGER "update_users_updated_at" BEFORE UPDATE ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "generate_request_number" function
CREATE FUNCTION "public"."generate_request_number" ("tenant_id" uuid) RETURNS text LANGUAGE plpgsql AS $$
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
$$;
-- Create "refresh_tokens" table
CREATE TABLE "public"."refresh_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "tenant_id" uuid NOT NULL,
  "token" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "refresh_tokens_token_key" UNIQUE ("token"),
  CONSTRAINT "fk_refresh_tokens_tenant" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "refresh_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_refresh_tokens_tenant" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_tenant" ON "public"."refresh_tokens" ("tenant_id");
-- Create "request_comments" table
CREATE TABLE "public"."request_comments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "request_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "comment" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "request_comments_request_id_fkey" FOREIGN KEY ("request_id") REFERENCES "public"."requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "request_comments_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_request_comments" to table: "request_comments"
CREATE INDEX "idx_request_comments" ON "public"."request_comments" ("request_id", "created_at");
-- Create index "idx_request_comments_user" to table: "request_comments"
CREATE INDEX "idx_request_comments_user" ON "public"."request_comments" ("user_id");
-- Create "user_tokens" table
CREATE TABLE "public"."user_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "tenant_id" uuid NOT NULL,
  "token" text NOT NULL,
  "type" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "used_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_tokens_tenant" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_user_tokens_tenant" to table: "user_tokens"
CREATE INDEX "idx_user_tokens_tenant" ON "public"."user_tokens" ("tenant_id");
-- Create index "idx_user_tokens_token" to table: "user_tokens"
CREATE INDEX "idx_user_tokens_token" ON "public"."user_tokens" ("token") WHERE (used_at IS NULL);
-- Create "manager_pending_requests" view
CREATE VIEW "public"."manager_pending_requests" (
  "id",
  "tenant_id",
  "request_number",
  "request_type",
  "category",
  "title",
  "description",
  "requester_id",
  "assignee_id",
  "routed_to_manager_id",
  "support_department_id",
  "support_category",
  "priority",
  "status",
  "due_date",
  "completed_at",
  "parent_request_id",
  "created_at",
  "updated_at",
  "requester_first_name",
  "requester_last_name"
) AS SELECT r.id,
    r.tenant_id,
    r.request_number,
    r.request_type,
    r.category,
    r.title,
    r.description,
    r.requester_id,
    r.assignee_id,
    r.routed_to_manager_id,
    r.support_department_id,
    r.support_category,
    r.priority,
    r.status,
    r.due_date,
    r.completed_at,
    r.parent_request_id,
    r.created_at,
    r.updated_at,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name
   FROM public.requests r
     JOIN public.employees req ON req.id = r.requester_id
  WHERE (r.status = ANY (ARRAY['new'::text, 'pending'::text])) AND (r.routed_to_manager_id IS NOT NULL OR r.support_department_id IS NOT NULL)
  ORDER BY r.created_at DESC;
-- Create "request_details" view
CREATE VIEW "public"."request_details" (
  "id",
  "tenant_id",
  "request_number",
  "request_type",
  "category",
  "title",
  "description",
  "requester_id",
  "assignee_id",
  "routed_to_manager_id",
  "support_department_id",
  "support_category",
  "priority",
  "status",
  "due_date",
  "completed_at",
  "parent_request_id",
  "created_at",
  "updated_at",
  "requester_first_name",
  "requester_last_name",
  "assignee_first_name",
  "assignee_last_name",
  "manager_first_name",
  "manager_last_name",
  "support_department_name"
) AS SELECT r.id,
    r.tenant_id,
    r.request_number,
    r.request_type,
    r.category,
    r.title,
    r.description,
    r.requester_id,
    r.assignee_id,
    r.routed_to_manager_id,
    r.support_department_id,
    r.support_category,
    r.priority,
    r.status,
    r.due_date,
    r.completed_at,
    r.parent_request_id,
    r.created_at,
    r.updated_at,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    assign.first_name AS assignee_first_name,
    assign.last_name AS assignee_last_name,
    mgr.first_name AS manager_first_name,
    mgr.last_name AS manager_last_name,
    dept.name AS support_department_name
   FROM public.requests r
     JOIN public.employees req ON req.id = r.requester_id
     LEFT JOIN public.employees assign ON assign.id = r.assignee_id
     LEFT JOIN public.employees mgr ON mgr.id = r.routed_to_manager_id
     LEFT JOIN public.departments dept ON dept.id = r.support_department_id;
-- Create "support_department_queue" view
CREATE VIEW "public"."support_department_queue" (
  "id",
  "tenant_id",
  "request_number",
  "request_type",
  "category",
  "title",
  "description",
  "requester_id",
  "assignee_id",
  "routed_to_manager_id",
  "support_department_id",
  "support_category",
  "priority",
  "status",
  "due_date",
  "completed_at",
  "parent_request_id",
  "created_at",
  "updated_at",
  "requester_first_name",
  "requester_last_name",
  "assignee_first_name",
  "assignee_last_name",
  "department_name",
  "chat_message_count"
) AS SELECT r.id,
    r.tenant_id,
    r.request_number,
    r.request_type,
    r.category,
    r.title,
    r.description,
    r.requester_id,
    r.assignee_id,
    r.routed_to_manager_id,
    r.support_department_id,
    r.support_category,
    r.priority,
    r.status,
    r.due_date,
    r.completed_at,
    r.parent_request_id,
    r.created_at,
    r.updated_at,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    assign.first_name AS assignee_first_name,
    assign.last_name AS assignee_last_name,
    dept.name AS department_name,
    count(cm.id) AS chat_message_count
   FROM public.requests r
     JOIN public.employees req ON req.id = r.requester_id
     LEFT JOIN public.employees assign ON assign.id = r.assignee_id
     LEFT JOIN public.departments dept ON dept.id = r.support_department_id
     LEFT JOIN public.request_chat_messages cm ON cm.request_id = r.id
  WHERE r.request_type = 'support'::text
  GROUP BY r.id, req.first_name, req.last_name, assign.first_name, assign.last_name, dept.name
  ORDER BY (
        CASE r.priority
            WHEN 'urgent'::text THEN 1
            WHEN 'high'::text THEN 2
            WHEN 'medium'::text THEN 3
            WHEN 'low'::text THEN 4
            ELSE NULL::integer
        END), r.created_at;
