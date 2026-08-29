ALTER TABLE public.attachments
    ADD COLUMN scan_status character varying(32) NOT NULL DEFAULT 'unscanned',
    ADD COLUMN scan_completed_at timestamp with time zone,
    ADD COLUMN scan_failure_reason character varying(512),
    ADD COLUMN optimization_status character varying(32) NOT NULL DEFAULT 'not_requested',
    ADD COLUMN optimization_attempts integer NOT NULL DEFAULT 0,
    ADD COLUMN optimization_started_at timestamp with time zone,
    ADD COLUMN optimization_completed_at timestamp with time zone,
    ADD COLUMN optimization_lease_expires_at timestamp with time zone,
    ADD COLUMN optimization_last_error character varying(512),
    ADD COLUMN updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD CONSTRAINT attachments_scan_status_check CHECK (
        scan_status IN ('unscanned', 'pending', 'clean', 'infected', 'failed')
    ),
    ADD CONSTRAINT attachments_optimization_status_check CHECK (
        optimization_status IN ('not_requested', 'queued', 'processing', 'succeeded', 'skipped', 'failed')
    ),
    ADD CONSTRAINT attachments_optimization_attempts_check CHECK (optimization_attempts >= 0);

CREATE INDEX idx_attachments_optimization_queue
    ON public.attachments (optimization_status, optimization_lease_expires_at, created_at, attachment_id)
    WHERE optimization_status IN ('queued', 'processing', 'failed');
