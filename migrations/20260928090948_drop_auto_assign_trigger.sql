-- Drop "auto_assign_support_request" trigger
DROP TRIGGER  IF EXISTS "auto_assign_support_request" ON "public"."requests";
-- Drop "auto_assign_support_request" function
DROP FUNCTION  IF EXISTS "public"."auto_assign_support_request";
