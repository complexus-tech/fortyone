-- Delete rows while the retirement trigger is present so stored derivatives
-- retain durable cleanup work even if this feature is rolled back.
DELETE FROM public.attachment_thumbnails;
DROP TABLE public.attachment_thumbnails;
DROP FUNCTION public.retire_attachment_thumbnail_object();
