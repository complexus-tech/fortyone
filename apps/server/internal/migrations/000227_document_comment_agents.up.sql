ALTER TYPE public.entity_type ADD VALUE IF NOT EXISTS 'document';

ALTER TABLE public.agent_invocations ALTER COLUMN team_id DROP NOT NULL;
ALTER TABLE public.agent_invocations ADD COLUMN action_started_at timestamptz;
ALTER TABLE public.agent_invocations ADD CONSTRAINT agent_invocations_surface_team
    CHECK ((surface = 'document_comment' AND team_id IS NULL) OR (surface <> 'document_comment' AND team_id IS NOT NULL));

ALTER TABLE public.document_comments
    ADD COLUMN creation_id uuid,
    ADD COLUMN creation_hash bytea,
    ADD COLUMN reply_to_comment_id uuid REFERENCES public.document_comments(comment_id) ON DELETE SET NULL;
CREATE UNIQUE INDEX document_comments_creation ON public.document_comments (created_by, creation_id) WHERE creation_id IS NOT NULL;

CREATE TABLE public.document_comment_notification_receipts (
    comment_id uuid NOT NULL REFERENCES public.document_comments(comment_id) ON DELETE CASCADE,
    recipient_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    notification_id uuid REFERENCES public.notifications(notification_id) ON DELETE SET NULL,
    PRIMARY KEY (comment_id, recipient_id)
);

-- This predicate deliberately has no admin bypass: private and restricted
-- documents retain their explicit document ACL in every delivery channel.
CREATE FUNCTION public.can_read_document(workspace uuid, actor uuid, target uuid)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS (
        SELECT 1 FROM public.documents AS document
        JOIN public.workspaces AS tenant ON tenant.workspace_id = document.workspace_id AND tenant.deleted_at IS NULL
        JOIN public.workspace_members AS membership ON membership.workspace_id = document.workspace_id AND membership.user_id = actor
        JOIN public.users AS account ON account.user_id = membership.user_id AND account.is_active = TRUE AND account.is_system = FALSE
        WHERE document.document_id = target AND document.workspace_id = workspace AND document.archived_at IS NULL
          AND (document.visibility = 'workspace' OR document.created_by = actor OR (document.visibility = 'restricted' AND EXISTS (
              SELECT 1 FROM public.document_members AS reader WHERE reader.document_id = document.document_id AND reader.user_id = actor
          )))
    );
$$;

CREATE TABLE public.document_agent_edits (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    invocation_id uuid NOT NULL UNIQUE REFERENCES public.agent_invocations(id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
    requester_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    expected_revision bigint NOT NULL CHECK (expected_revision > 0),
    expected_epoch bigint NOT NULL CHECK (expected_epoch > 0),
    operations jsonb NOT NULL CHECK (jsonb_typeof(operations) = 'object'),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','applied','conflict','forbidden','failed')),
    resulting_revision bigint,
    receipt jsonb,
    error_code text,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    applied_at timestamptz
);
CREATE INDEX document_agent_edits_pending ON public.document_agent_edits(created_at, id) WHERE status = 'pending';

CREATE TABLE public.document_comment_mentions (
 comment_id uuid NOT NULL REFERENCES public.document_comments(comment_id) ON DELETE CASCADE,
 user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
 PRIMARY KEY(comment_id,user_id)
);

CREATE TABLE public.document_agent_consumer_health (
 instance_id uuid PRIMARY KEY,
 expires_at timestamptz NOT NULL
);

ALTER TABLE public.document_relationships ADD COLUMN source text NOT NULL DEFAULT 'manual' CHECK (source IN ('manual','body'));
CREATE TABLE public.document_body_references (
 document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
 entity_type varchar(20) NOT NULL CHECK (entity_type IN ('story','objective')),
 entity_id uuid NOT NULL,
 PRIMARY KEY(document_id,entity_type,entity_id)
);
CREATE TABLE public.document_body_mentions (
 document_id uuid NOT NULL REFERENCES public.documents(document_id) ON DELETE CASCADE,
 user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
 PRIMARY KEY(document_id,user_id)
);

CREATE FUNCTION public.can_edit_document(workspace uuid, actor uuid, target uuid)
RETURNS boolean LANGUAGE sql STABLE AS $$
 SELECT EXISTS (
  SELECT 1 FROM public.documents AS document
  JOIN public.workspaces AS tenant ON tenant.workspace_id=document.workspace_id AND tenant.deleted_at IS NULL
  JOIN public.workspace_members AS membership ON membership.workspace_id=document.workspace_id AND membership.user_id=actor AND membership.role<>CAST('guest' AS public.user_role)
  JOIN public.users AS account ON account.user_id=actor AND account.is_active=TRUE AND account.is_system=FALSE
  WHERE document.document_id=target AND document.workspace_id=workspace AND document.archived_at IS NULL
   AND (document.visibility='workspace' OR document.created_by=actor OR (document.visibility='restricted' AND EXISTS (
    SELECT 1 FROM public.document_members AS editor WHERE editor.document_id=target AND editor.user_id=actor AND editor.role='editor'
   )))
 );
$$;

CREATE FUNCTION public.sync_document_body_references(workspace uuid, actor uuid, target uuid, refs jsonb)
RETURNS void LANGUAGE plpgsql AS $$
DECLARE reference jsonb; reference_id uuid; reference_type text;
BEGIN
 IF NOT public.can_edit_document(workspace,actor,target) THEN RAISE EXCEPTION 'Document access changed' USING ERRCODE='42501'; END IF;
 IF jsonb_typeof(refs)<>'array' OR jsonb_array_length(refs)>1000 THEN RAISE EXCEPTION 'Invalid document references' USING ERRCODE='22023'; END IF;
 FOR reference IN SELECT value FROM jsonb_array_elements(refs) LOOP
  reference_id:=CAST(reference->>'id' AS uuid);reference_type:=reference->>'entityType';
  IF reference_id IS NULL OR reference_type IS NULL OR reference_type NOT IN ('story','objective') THEN RAISE EXCEPTION 'Invalid document reference' USING ERRCODE='22023'; END IF;
  -- Existing chips may outlive target access. Unrelated prose edits must remain possible.
  IF NOT EXISTS(SELECT 1 FROM public.document_body_references AS existing WHERE existing.document_id=target AND existing.entity_type=reference_type AND existing.entity_id=reference_id) THEN
   IF reference_type='story' THEN
    PERFORM 1 FROM public.stories AS story
     JOIN public.teams AS team ON team.team_id=story.team_id AND team.workspace_id=workspace
     JOIN public.team_members AS member ON member.team_id=team.team_id AND member.user_id=actor
     WHERE story.id=reference_id AND story.workspace_id=workspace AND story.deleted_at IS NULL FOR KEY SHARE OF story;
   ELSE
    PERFORM 1 FROM public.objectives AS objective
     JOIN public.teams AS team ON team.team_id=objective.team_id AND team.workspace_id=workspace
     JOIN public.team_members AS member ON member.team_id=team.team_id AND member.user_id=actor
     WHERE objective.objective_id=reference_id AND objective.workspace_id=workspace FOR KEY SHARE OF objective;
   END IF;
   IF NOT FOUND THEN RAISE EXCEPTION 'Reference is not accessible' USING ERRCODE='42501'; END IF;
   INSERT INTO public.document_body_references(document_id,entity_type,entity_id) VALUES(target,reference_type,reference_id) ON CONFLICT DO NOTHING;
   INSERT INTO public.document_relationships(document_id,workspace_id,entity_type,entity_id,created_by,source)
    VALUES(target,workspace,reference_type,reference_id,actor,'body') ON CONFLICT(document_id,entity_type,entity_id) DO NOTHING;
  END IF;
 END LOOP;
 DELETE FROM public.document_relationships AS relationship
 WHERE relationship.document_id=target AND relationship.source='body' AND NOT EXISTS (
  SELECT 1 FROM jsonb_array_elements(refs) AS item WHERE item->>'entityType'=relationship.entity_type AND CAST(item->>'id' AS uuid)=relationship.entity_id
 );
 DELETE FROM public.document_body_references AS existing WHERE existing.document_id=target AND NOT EXISTS (
  SELECT 1 FROM jsonb_array_elements(refs) AS item WHERE item->>'entityType'=existing.entity_type AND CAST(item->>'id' AS uuid)=existing.entity_id
 );
END;
$$;

CREATE FUNCTION public.sync_document_body_mentions(workspace uuid, actor uuid, target uuid, mentions jsonb)
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
 WHERE public.can_read_document(workspace,CAST(value AS uuid),target)
 ON CONFLICT DO NOTHING;
END;
$$;

CREATE FUNCTION public.workspace_can_use_maya(workspace uuid)
RETURNS boolean LANGUAGE sql STABLE AS $$
 SELECT EXISTS (SELECT 1 FROM public.workspaces AS tenant
 WHERE tenant.workspace_id=workspace AND tenant.deleted_at IS NULL
 AND (tenant.trial_ends_on>CURRENT_TIMESTAMP OR EXISTS (
  SELECT 1 FROM public.workspace_subscriptions AS subscription
  WHERE subscription.workspace_id=tenant.workspace_id AND subscription.subscription_tier<>'free'
   AND subscription.subscription_status IN ('active','trialing','past_due')
 )));
$$;

-- Document writes originate in both Go and the collaboration process. A
-- transaction-local trigger gives both paths the same durable delivery wakeup.
CREATE TABLE public.document_notification_deliveries (
 notification_id uuid PRIMARY KEY REFERENCES public.notifications(notification_id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 delivered_at timestamptz
);
CREATE INDEX document_notification_deliveries_pending ON public.document_notification_deliveries(created_at,notification_id) WHERE delivered_at IS NULL;
CREATE FUNCTION public.enqueue_document_notification_delivery() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF CAST(NEW.entity_type AS text)='document' THEN
  INSERT INTO public.document_notification_deliveries(notification_id) VALUES(NEW.notification_id) ON CONFLICT DO NOTHING;
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER document_notification_delivery AFTER INSERT ON public.notifications
 FOR EACH ROW EXECUTE FUNCTION public.enqueue_document_notification_delivery();
