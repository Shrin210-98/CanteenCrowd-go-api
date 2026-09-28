-- Modify "log_request_assignee_change" function
CREATE OR REPLACE FUNCTION "public"."log_request_assignee_change" () RETURNS trigger LANGUAGE plpgsql AS $$
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
$$;
-- Modify "log_request_status_change" function
CREATE OR REPLACE FUNCTION "public"."log_request_status_change" () RETURNS trigger LANGUAGE plpgsql AS $$
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
$$;
