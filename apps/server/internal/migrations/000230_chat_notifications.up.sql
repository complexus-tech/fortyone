-- Chat notifications use their own preference category and destination. Inbox,
-- push and digest readers recheck current conversation membership before use.
ALTER TYPE public.notification_type ADD VALUE IF NOT EXISTS 'chat_message';
ALTER TYPE public.entity_type ADD VALUE IF NOT EXISTS 'chat_message';
