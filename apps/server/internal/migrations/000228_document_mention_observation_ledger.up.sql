-- Record what is already in the document independently from who may receive
-- a notification now. Sharing or reactivating an account must not make an
-- unchanged old mention look newly inserted on an unrelated edit.
CREATE OR REPLACE FUNCTION public.sync_document_body_mentions(workspace uuid, actor uuid, target uuid, mentions jsonb)
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
 IF NOT public.can_edit_document(workspace,actor,target) THEN RAISE EXCEPTION 'Document access changed' USING ERRCODE='42501'; END IF;
 IF jsonb_typeof(mentions)<>'array' OR jsonb_array_length(mentions)>1000 THEN RAISE EXCEPTION 'Invalid document mentions' USING ERRCODE='22023'; END IF;
 INSERT INTO public.notifications(recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message,dedupe_key,in_app_enabled,coalescing_key)
 SELECT candidate.user_id,workspace,CAST('mention' AS public.notification_type),CAST('document' AS public.entity_type),target,actor,'Mentioned you in a document',
  jsonb_build_object('template','{actor} mentioned you in {document_title}','variables',jsonb_build_object(
   'actor',jsonb_build_object('value',COALESCE(NULLIF(author.full_name,''),author.username),'type','text'),
   'document_title',jsonb_build_object('value',document.title,'type','text'),'document_id',jsonb_build_object('value',CAST(target AS text),'type','text')),
   'identityReferences',jsonb_build_object('actor',actor)),
  'document-body:'||CAST(target AS text)||':'||CAST(document.revision AS text)||':'||CAST(candidate.user_id AS text),
  CASE WHEN jsonb_typeof(preference.preferences->'mention'->'in_app')='boolean' THEN CAST(preference.preferences->'mention'->>'in_app' AS boolean) ELSE TRUE END,'document_body'
 FROM (SELECT DISTINCT CAST(value AS uuid) AS user_id FROM jsonb_array_elements_text(mentions)) AS candidate
 JOIN public.documents AS document ON document.document_id=target AND document.workspace_id=workspace
 JOIN public.users AS author ON author.user_id=actor
 LEFT JOIN public.notification_preferences AS preference ON preference.user_id=candidate.user_id AND preference.workspace_id=workspace
 WHERE candidate.user_id<>actor AND public.can_read_document(workspace,candidate.user_id,target)
  AND NOT EXISTS(SELECT 1 FROM public.document_body_mentions AS existing WHERE existing.document_id=target AND existing.user_id=candidate.user_id)
 ON CONFLICT(dedupe_key) DO NOTHING;
 DELETE FROM public.document_body_mentions AS existing WHERE existing.document_id=target AND NOT EXISTS(
  SELECT 1 FROM jsonb_array_elements_text(mentions) AS item WHERE CAST(item AS uuid)=existing.user_id
 );
 INSERT INTO public.document_body_mentions(document_id,user_id)
 SELECT target,CAST(value AS uuid) FROM jsonb_array_elements_text(mentions)
 WHERE EXISTS (SELECT 1 FROM public.users AS account WHERE account.user_id=CAST(value AS uuid) AND account.is_system=FALSE)
 ON CONFLICT DO NOTHING;
END;
$$;
