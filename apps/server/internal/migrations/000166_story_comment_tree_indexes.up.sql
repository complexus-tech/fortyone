-- Root comment pagination and reply hydration are separate access patterns.
-- Partial indexes keep each tree branch compact while matching the stable
-- ordering used by the SQLC comment-read queries.
CREATE INDEX idx_story_comments_roots_page
    ON public.story_comments (story_id, created_at DESC, comment_id DESC)
    WHERE parent_id IS NULL;

CREATE INDEX idx_story_comments_replies_page
    ON public.story_comments (story_id, parent_id, created_at, comment_id)
    WHERE parent_id IS NOT NULL;
