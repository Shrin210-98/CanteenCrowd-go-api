-- Modify "users" table
ALTER TABLE "public"."users" DROP CONSTRAINT "users_user_type_check", ADD CONSTRAINT "users_user_type_check" CHECK (user_type = ANY (ARRAY['tenant_owner'::text, 'staff'::text, 'external'::text]));
