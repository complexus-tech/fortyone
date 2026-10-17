-- name: LockTranscriptionAudio :one
SELECT ca.attachment_id FROM public.chat_attachments ca
JOIN public.chat_conversation_messages m ON m.id=ca.message_id
JOIN public.attachments a ON a.attachment_id=ca.attachment_id
WHERE ca.attachment_id=sqlc.arg(attachment_id) AND ca.workspace_id=sqlc.arg(workspace_id)
AND ca.conversation_id=sqlc.arg(conversation_id) AND m.deleted_at IS NULL
AND a.mime_type IN ('audio/webm','audio/mp4')
AND a.size BETWEEN 1 AND 20971520 AND a.scan_status IN ('unscanned','clean')
FOR UPDATE OF ca;

-- name: GetAudioTranscription :one
SELECT COALESCE(t.status,'not_requested') AS status, COALESCE(t.transcript,'') AS transcript,
COALESCE(t.error_code,'') AS error_code, t.updated_at
FROM public.chat_attachments ca
JOIN public.chat_conversation_messages m ON m.id=ca.message_id AND m.deleted_at IS NULL
JOIN public.attachments a ON a.attachment_id=ca.attachment_id AND a.scan_status IN ('unscanned','clean')
LEFT JOIN public.chat_audio_transcriptions t ON t.attachment_id=ca.attachment_id
WHERE ca.attachment_id=sqlc.arg(attachment_id) AND ca.workspace_id=sqlc.arg(workspace_id)
AND ca.conversation_id=sqlc.arg(conversation_id)
AND a.mime_type IN ('audio/webm','audio/mp4') AND a.size BETWEEN 1 AND 20971520;

-- name: RequestAudioTranscription :exec
INSERT INTO public.chat_audio_transcriptions(attachment_id,workspace_id,conversation_id,requested_by)
VALUES(sqlc.arg(attachment_id),sqlc.arg(workspace_id),sqlc.arg(conversation_id),sqlc.arg(user_id))
ON CONFLICT (attachment_id) DO UPDATE SET
status='pending', transcript=NULL, error_code=NULL, model=NULL, lease_token=NULL, leased_until=NULL,
requested_by=EXCLUDED.requested_by, updated_at=now()
WHERE chat_audio_transcriptions.status='failed' AND CAST(sqlc.arg(retry) AS boolean);

-- name: ExpireAudioTranscriptions :exec
UPDATE public.chat_audio_transcriptions SET status='failed', error_code='interrupted',
lease_token=NULL, leased_until=NULL, updated_at=now()
WHERE status='processing' AND leased_until < now();

-- name: ClaimAudioTranscriptions :many
WITH pending AS (
  SELECT attachment_id FROM public.chat_audio_transcriptions WHERE status='pending'
  AND EXISTS (SELECT 1 FROM public.workspaces w LEFT JOIN public.workspace_settings s ON s.workspace_id=w.workspace_id WHERE w.workspace_id=public.chat_audio_transcriptions.workspace_id AND w.deleted_at IS NULL AND COALESCE(s.chat_enabled,TRUE))
  ORDER BY created_at,attachment_id LIMIT sqlc.arg(page_limit) FOR UPDATE SKIP LOCKED
)
UPDATE public.chat_audio_transcriptions t SET status='processing', attempts=t.attempts+1,
lease_token=gen_random_uuid(),leased_until=now()+interval '5 minutes',updated_at=now()
FROM pending p WHERE t.attachment_id=p.attachment_id
RETURNING t.attachment_id,t.workspace_id,t.conversation_id,t.requested_by,t.lease_token;

-- name: ReadTranscriptionAudio :one
SELECT a.blob_name,a.filename,a.mime_type,a.size FROM public.chat_audio_transcriptions t
JOIN public.chat_attachments ca ON ca.attachment_id=t.attachment_id AND ca.workspace_id=t.workspace_id AND ca.conversation_id=t.conversation_id
JOIN public.chat_conversation_messages m ON m.id=ca.message_id AND m.deleted_at IS NULL
JOIN public.attachments a ON a.attachment_id=ca.attachment_id
WHERE t.attachment_id=sqlc.arg(attachment_id) AND t.lease_token=sqlc.arg(lease_token) AND t.status='processing' AND t.leased_until>now()
AND a.mime_type IN ('audio/webm','audio/mp4') AND a.size BETWEEN 1 AND 20971520 AND a.scan_status IN ('unscanned','clean')
AND public.can_read_chat_conversation(t.workspace_id,t.requested_by,t.conversation_id);

-- name: CompleteAudioTranscription :execrows
UPDATE public.chat_audio_transcriptions t SET status='ready',transcript=sqlc.arg(transcript),model=sqlc.arg(model),error_code=NULL,
lease_token=NULL,leased_until=NULL,updated_at=now()
WHERE t.attachment_id=sqlc.arg(attachment_id) AND t.lease_token=sqlc.arg(lease_token) AND t.status='processing' AND t.leased_until>now()
AND EXISTS(SELECT 1 FROM public.chat_attachments ca JOIN public.chat_conversation_messages m ON m.id=ca.message_id
JOIN public.attachments a ON a.attachment_id=ca.attachment_id AND a.scan_status IN ('unscanned','clean')
WHERE ca.attachment_id=t.attachment_id AND m.deleted_at IS NULL);

-- name: FailAudioTranscription :execrows
UPDATE public.chat_audio_transcriptions SET status='failed',error_code=sqlc.arg(error_code),transcript=NULL,
lease_token=NULL,leased_until=NULL,updated_at=now()
WHERE attachment_id=sqlc.arg(attachment_id) AND lease_token=sqlc.arg(lease_token) AND status='processing';

-- name: RequeueAudioTranscription :execrows
UPDATE public.chat_audio_transcriptions SET status='pending', lease_token=NULL, leased_until=NULL, updated_at=now()
WHERE attachment_id=sqlc.arg(attachment_id) AND lease_token=sqlc.arg(lease_token) AND status='processing';
