DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.story_comment_seen)
       OR EXISTS (SELECT 1 FROM public.maya_reply_notification_deliveries) THEN
        RAISE EXCEPTION 'Maya reply receipts remain; preserve them and repair forward';
    END IF;
END $$;

DROP TABLE public.maya_reply_notification_deliveries;
DROP TABLE public.story_comment_seen;
