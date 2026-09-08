-- Do not restore the obsolete uniqueness rule: separate events may already
-- exist for the same recipient/resource, and restoring it would lose history.
SELECT 1;
