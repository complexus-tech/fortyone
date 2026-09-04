DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.google_drive_document_import_operations) THEN
        RAISE EXCEPTION
            'refusing to discard durable Google Drive document import operations'
            USING ERRCODE = '55000';
    END IF;
END;
$$;

DROP TABLE IF EXISTS public.google_drive_document_import_operations;
