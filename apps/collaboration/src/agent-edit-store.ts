import type { Pool, PoolClient } from "pg";
import * as Y from "yjs";
import { AgentEditError, applyAgentOperations } from "./agent-operations";
import {
  DocumentStore,
  MAX_DOCUMENT_BYTES,
  renderDocument,
  restoreDocument,
} from "./documents";

type PendingEdit = {
  id: string;
  invocation_id: string;
  workspace_id: string;
  document_id: string;
  requester_id: string;
  expected_revision: string;
  expected_epoch: string;
  operations: unknown;
  status: string;
  expired: boolean;
  proposal_matches: boolean;
  source_matches: boolean;
  approved: boolean;
};
type StoredDocument = {
  collaboration_state: Buffer | null;
  title: string;
  content_html: string;
  revision: string;
  collaboration_epoch: string;
  workspace_id: string;
  visibility: string;
  created_by: string;
};
export type CommittedAgentEdit = {
  name: string;
  state: Uint8Array;
  revision: string;
};

export class AgentEditStore {
  private readonly transactions: DocumentStore;
  constructor(private readonly pool: Pool) {
    this.transactions = new DocumentStore(pool);
  }

  async pending(limit = 5): Promise<{ id: string; invocation_id: string }[]> {
    const result = await this.pool.query(
      "SELECT id, invocation_id FROM document_agent_edits WHERE status = 'pending' ORDER BY created_at, id LIMIT $1",
      [Math.min(20, Math.max(1, limit))],
    );
    return result.rows;
  }

  async apply(
    id: string,
    invocationId: string,
  ): Promise<CommittedAgentEdit | null> {
    return this.transactions.transact(async (client) => {
      // Match core's lock order. Cancellation cannot race a committed command.
      const invocation = await client.query<{
        status: string;
        started: boolean;
      }>(
        "SELECT status, action_started_at IS NOT NULL AS started FROM agent_invocations WHERE id = $1 FOR UPDATE SKIP LOCKED",
        [invocationId],
      );
      if (!invocation.rows[0]) return null;
      const queued = await client.query<PendingEdit>(
        `
        SELECT edit.*,
          edit.created_at < clock_timestamp() - interval '15 minutes' AS expired,
          (run.surface = 'document_comment' AND run.entity_id = edit.document_id
            AND run.workspace_id = edit.workspace_id AND run.requester_id = edit.requester_id
            AND run.proposal->'action'->>'version' = '1'
            AND run.proposal->'action'->>'documentId' = CAST(edit.document_id AS text)
            AND run.proposal->'action'->>'expectedRevision' = CAST(edit.expected_revision AS text)
            AND run.proposal->'action'->>'expectedEpoch' = CAST(edit.expected_epoch AS text)
            AND run.proposal->'action'->'operations' = edit.operations) AS proposal_matches,
          (COALESCE(run.proposal->>'requiresApproval', 'false') <> 'true' OR run.approved_at IS NOT NULL) AS approved,
          EXISTS (SELECT 1 FROM document_comments AS source
            JOIN document_comment_threads AS thread ON thread.thread_id = source.thread_id
            WHERE source.comment_id = run.source_id AND source.created_by = edit.requester_id
              AND thread.thread_id = run.conversation_id AND thread.document_id = edit.document_id
              AND thread.workspace_id = edit.workspace_id
              AND encode(sha256(convert_to(source.body, 'UTF8')), 'hex') = run.source_hash
              AND source.updated_at = CAST(run.source_revision AS timestamptz)) AS source_matches
        FROM document_agent_edits AS edit JOIN agent_invocations AS run ON run.id = edit.invocation_id
        WHERE edit.id = $1 AND edit.invocation_id = $2 FOR UPDATE OF edit`,
        [id, invocationId],
      );
      const edit = queued.rows[0];
      if (!edit || edit.status !== "pending") return null;
      await client.query("SAVEPOINT agent_edit");
      try {
        if (edit.expired) throw new AgentEditError("expired");
        if (
          !invocation.rows[0].started ||
          !["queued", "running"].includes(invocation.rows[0].status) ||
          !edit.proposal_matches ||
          !edit.source_matches ||
          !edit.approved
        ) {
          throw new AgentEditError("invocation_not_authorized", "forbidden");
        }
        const result = await client.query<StoredDocument>(
          `
          SELECT collaboration_state, title, content_html, revision, collaboration_epoch, workspace_id, visibility, created_by
          FROM documents WHERE document_id = $1 AND workspace_id = $2 AND archived_at IS NULL FOR UPDATE`,
          [edit.document_id, edit.workspace_id],
        );
        const stored = result.rows[0];
        if (!stored)
          throw new AgentEditError("document_unavailable", "forbidden");
        await this.authorize(client, edit, stored);
        if (!stored.collaboration_state)
          throw new AgentEditError("collaboration_unavailable", "forbidden");
        if (
          stored.revision !== edit.expected_revision ||
          stored.collaboration_epoch !== edit.expected_epoch
        )
          throw new AgentEditError("document_changed", "conflict");
        let document: Y.Doc;
        try {
          document = restoreDocument(
            stored.collaboration_state,
            stored.title,
            stored.content_html,
          );
        } catch {
          throw new AgentEditError("document_invalid");
        }
        try {
          const changedFields = applyAgentOperations(document, edit.operations);
          const state = Buffer.from(Y.encodeStateAsUpdate(document));
          if (state.byteLength > MAX_DOCUMENT_BYTES)
            throw new AgentEditError("document_too_large");
          let rendered: ReturnType<typeof renderDocument>;
          try {
            rendered = renderDocument(document);
          } catch {
            throw new AgentEditError("document_invalid");
          }
          let revision = stored.revision;
          if (changedFields.length) {
            const updated = await client.query<{ revision: string }>(
              `
              UPDATE documents SET collaboration_state = $2, title = $3, content_html = $4,
                content_text = $5, updated_by = $6 WHERE document_id = $1 RETURNING revision`,
              [
                edit.document_id,
                state,
                rendered.title,
                rendered.html,
                rendered.text,
                edit.requester_id,
              ],
            );
            revision = updated.rows[0].revision;
            await client.query(
              "SELECT sync_document_body_references($1, $2, $3, CAST($4 AS jsonb)), sync_document_body_mentions($1, $2, $3, CAST($5 AS jsonb))",
              [
                edit.workspace_id,
                edit.requester_id,
                edit.document_id,
                JSON.stringify(rendered.references),
                JSON.stringify(rendered.mentions),
              ],
            );
          }
          const receipt = {
            editId: edit.id,
            invocationId,
            documentId: edit.document_id,
            requesterId: edit.requester_id,
            beforeRevision: Number(stored.revision),
            revision: Number(revision),
            collaborationEpoch: Number(stored.collaboration_epoch),
            changedFields,
          };
          await client.query(
            `UPDATE document_agent_edits SET status = 'applied', resulting_revision = $2,
            receipt = CAST($3 AS jsonb), error_code = NULL, updated_at = clock_timestamp(), applied_at = clock_timestamp() WHERE id = $1`,
            [id, revision, JSON.stringify(receipt)],
          );
          await client.query("SELECT pg_notify('document_agent_edits', $1)", [
            `${edit.document_id}:${stored.collaboration_epoch}`,
          ]);
          return changedFields.length
            ? {
                name: `${edit.document_id}:${stored.collaboration_epoch}`,
                state,
                revision,
              }
            : null;
        } finally {
          document.destroy();
        }
      } catch (error) {
        const failure =
          error instanceof AgentEditError
            ? error
            : (error as { code?: string }).code === "42501"
              ? new AgentEditError("edit_permission_revoked", "forbidden")
              : null;
        if (!failure) throw error;
        await client.query("ROLLBACK TO SAVEPOINT agent_edit");
        await client.query(
          "UPDATE document_agent_edits SET status = $2, error_code = $3, updated_at = clock_timestamp() WHERE id = $1",
          [id, failure.status, failure.code],
        );
        return null;
      }
    });
  }

  private async authorize(
    client: PoolClient,
    edit: PendingEdit,
    document: StoredDocument,
  ) {
    const membership = await client.query(
      `
      SELECT membership.role, workspace_can_use_maya(workspace.workspace_id) AS "mayaAllowed" FROM workspace_members AS membership
      JOIN users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE AND actor.is_system = FALSE
      JOIN workspaces AS workspace ON workspace.workspace_id = membership.workspace_id AND workspace.deleted_at IS NULL
      WHERE membership.workspace_id = $1 AND membership.user_id = $2 AND membership.role IN ('member', 'admin')
      FOR SHARE OF membership, actor, workspace`,
      [edit.workspace_id, edit.requester_id],
    );
    if (!membership.rows[0])
      throw new AgentEditError("edit_permission_revoked", "forbidden");
    if (!membership.rows[0].mayaAllowed)
      throw new AgentEditError("maya_access_revoked", "forbidden");
    if (
      document.visibility === "workspace" ||
      document.created_by === edit.requester_id
    )
      return;
    if (document.visibility === "restricted") {
      const access = await client.query(
        "SELECT role FROM document_members WHERE document_id = $1 AND user_id = $2 AND role = 'editor' FOR SHARE",
        [edit.document_id, edit.requester_id],
      );
      if (access.rows[0]) return;
    }
    throw new AgentEditError("edit_permission_revoked", "forbidden");
  }

  async revisions(names: string[]) {
    if (!names.length) return [];
    const result = await this.pool.query<{ name: string; revision: string }>(
      `
      SELECT CAST(document.document_id AS text) || ':' || CAST(document.collaboration_epoch AS text) AS name, document.revision
      FROM documents AS document WHERE CAST(document.document_id AS text) || ':' || CAST(document.collaboration_epoch AS text) = ANY(CAST($1 AS text[]))
        AND document.archived_at IS NULL AND EXISTS (SELECT 1 FROM document_agent_edits AS edit
          WHERE edit.document_id = document.document_id AND edit.expected_epoch = document.collaboration_epoch AND edit.status = 'applied')`,
      [names],
    );
    return result.rows;
  }

  async committedState(name: string) {
    const result = await this.pool.query<{
      collaboration_state: Buffer;
      revision: string;
    }>(
      `
      SELECT collaboration_state, revision FROM documents WHERE CAST(document_id AS text) || ':' || CAST(collaboration_epoch AS text) = $1
        AND collaboration_state IS NOT NULL AND archived_at IS NULL`,
      [name],
    );
    return result.rows[0];
  }

  async heartbeat(instanceId: string) {
    await this.pool.query(
      "DELETE FROM document_agent_consumer_health WHERE expires_at < clock_timestamp() - interval '1 hour'",
    );
    await this.pool.query(
      "INSERT INTO document_agent_consumer_health(instance_id, expires_at) VALUES ($1, clock_timestamp() + interval '45 seconds') ON CONFLICT (instance_id) DO UPDATE SET expires_at = EXCLUDED.expires_at",
      [instanceId],
    );
  }
  async removeHeartbeat(instanceId: string) {
    await this.pool.query(
      "DELETE FROM document_agent_consumer_health WHERE instance_id = $1 OR expires_at < clock_timestamp() - interval '1 hour'",
      [instanceId],
    );
  }
}
