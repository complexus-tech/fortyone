DROP INDEX public.stories_active_status_count_idx;
ALTER TABLE public.statuses DROP CONSTRAINT statuses_wip_limit_range, DROP COLUMN wip_limit;
