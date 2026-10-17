-- Non-destructive default alignment. Apply after 001_prd_completion.sql.
-- This changes future INSERT defaults only; existing account values are retained.
BEGIN;
ALTER TABLE sikapj.users ALTER COLUMN must_change_password SET DEFAULT TRUE;
COMMIT;
