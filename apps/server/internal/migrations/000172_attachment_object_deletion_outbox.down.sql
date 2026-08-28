-- Migration 000172 is forward-only after any attachment-object deletion work
-- exists. Dropping the table would discard the only durable copy of object
-- routing metadata and could permanently leak retained blobs.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.attachment_object_deletion_outbox
    ) THEN
        RAISE EXCEPTION
            'cannot remove attachment object deletion outbox while any delivery row exists; drain or repair it and deploy a forward fix'
            USING ERRCODE = '55000';
    END IF;
END;
$$;

DROP TABLE public.attachment_object_deletion_outbox;
