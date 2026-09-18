import { createHash } from "node:crypto";
import type { Pool, PoolClient } from "pg";
import { getSchema, type JSONContent } from "@tiptap/core";
import { generateHTML, generateJSON } from "@tiptap/html/server";
import { prosemirrorJSONToYDoc, yDocToProsemirrorJSON } from "@tiptap/y-tiptap";
import * as Y from "yjs";
import { createDocumentExtensions } from "@fortyone/document-editor";

const extensions = createDocumentExtensions({ collaborative: true });
const schema = getSchema(extensions);
export const MAX_DOCUMENT_BYTES = 4 * 1024 * 1024;
export type Identity = {
  tokenHash: string;
  userId: string;
  canEdit: boolean;
  name: string;
};
type StoredDocument = {
  collaboration_state: Buffer | null;
  title: string;
  content_html: string;
  revision: string;
};

export function parseDocumentName(name: string) {
  const match =
    /^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}):([1-9][0-9]{0,14})$/.exec(
      name,
    );
  if (!match) throw new Error("Invalid document");
  return { id: match[1], epoch: match[2] };
}

export function cleanDocumentJSON(node: JSONContent): JSONContent | null {
  if (
    (node.type === "image" || node.type === "documentVideo") &&
    node.attrs?.isUploading
  )
    return null;
  return {
    ...node,
    content: node.content
      ?.map(cleanDocumentJSON)
      .filter((child): child is JSONContent => child !== null),
  };
}

export function renderDocument(document: Y.Doc) {
  const raw = yDocToProsemirrorJSON(document, "default") as JSONContent;
  const json = cleanDocumentJSON(raw) ?? {
    type: "doc",
    content: [{ type: "paragraph" }],
  };
  if (!json.content?.length) json.content = [{ type: "paragraph" }];
  const title =
    document.getText("title").toString().trim() || "Untitled document";
  if ([...title].length > 255) throw new Error("Document title is too long");
  const html = generateHTML(json, extensions);
  const text = schema
    .nodeFromJSON(json)
    .textBetween(0, schema.nodeFromJSON(json).content.size, "\n");
  if (Buffer.byteLength(html) > MAX_DOCUMENT_BYTES)
    throw new Error("Document is too large");
  return { title, html, text };
}

// A single Postgres transaction validates access, merges CRDT updates and stores
// the rendered content. A WebSocket acknowledgement therefore means durable.
export class DocumentStore {
  constructor(private readonly pool: Pool) {}

  async ping() {
    await this.pool.query("SELECT 1");
  }

  async authorize(
    name: string,
    tokenHash: string,
    client: Pool | PoolClient = this.pool,
  ): Promise<Identity> {
    const { id, epoch } = parseDocumentName(name);
    const result = await client.query<Identity>(
      `
      SELECT actor.user_id AS "userId", COALESCE(NULLIF(actor.full_name, ''), actor.username) AS name,
        COALESCE((membership.role <> 'guest' AND (document.visibility = 'workspace' OR document.created_by = actor.user_id OR member.role = 'editor')), FALSE) AS "canEdit"
      FROM document_collaboration_sessions AS session
      JOIN documents AS document ON document.document_id = session.document_id
      JOIN users AS actor ON actor.user_id = session.user_id AND actor.is_active = TRUE
      JOIN workspace_members AS membership ON membership.workspace_id = document.workspace_id AND membership.user_id = actor.user_id
      JOIN workspaces AS workspace ON workspace.workspace_id = document.workspace_id AND workspace.deleted_at IS NULL
      LEFT JOIN document_members AS member ON member.document_id = document.document_id AND member.user_id = actor.user_id
      WHERE session.token_hash = $1 AND document.document_id = $2
        AND session.epoch = $3 AND document.collaboration_epoch = $3
        AND session.session_version = actor.auth_session_version
        AND session.expires_at > CURRENT_TIMESTAMP AND document.archived_at IS NULL
        AND (document.visibility = 'workspace' OR document.created_by = actor.user_id OR (document.visibility = 'restricted' AND member.user_id IS NOT NULL))
    `,
      [tokenHash, id, epoch],
    );
    if (!result.rows[0])
      throw new Error("Document access expired or was revoked");
    return { ...result.rows[0], tokenHash };
  }

  async authenticate(name: string, token: string) {
    if (!/^[0-9a-f]{64}$/.test(token))
      throw new Error("Invalid document session");
    return this.authorize(
      name,
      createHash("sha256").update(token).digest("hex"),
    );
  }

  async transact<T>(operation: (client: PoolClient) => Promise<T>): Promise<T> {
    for (let attempt = 0; ; attempt++) {
      const client = await this.pool.connect();
      try {
        await client.query("BEGIN ISOLATION LEVEL SERIALIZABLE");
        const result = await operation(client);
        await client.query("COMMIT");
        return result;
      } catch (error) {
        await client.query("ROLLBACK");
        if (attempt < 3 && (error as { code?: string }).code === "40001")
          continue;
        throw error;
      } finally {
        client.release();
      }
    }
  }

  async load(name: string, identity: Identity) {
    return this.transact(async (client) => {
      const { id, epoch } = parseDocumentName(name);
      const result = await client.query<StoredDocument>(
        "SELECT collaboration_state, title, content_html, revision FROM documents WHERE document_id = $1 AND collaboration_epoch = $2 AND archived_at IS NULL FOR UPDATE",
        [id, epoch],
      );
      if (!result.rows[0]) throw new Error("Document changed");
      await this.authorize(name, identity.tokenHash, client);
      const row = result.rows[0];
      if (row.collaboration_state) return row.collaboration_state;
      // Initialization is locked and persisted once, before any client receives
      // it. Concurrent first connections cannot seed duplicate paragraphs.
      const document = prosemirrorJSONToYDoc(
        schema,
        generateJSON(row.content_html || "<p></p>", extensions),
        "default",
      );
      document.getText("title").insert(0, row.title);
      const state = Buffer.from(Y.encodeStateAsUpdate(document));
      document.destroy();
      await client.query(
        "UPDATE documents SET collaboration_state = $2 WHERE document_id = $1",
        [id, state],
      );
      return state;
    });
  }

  async apply(name: string, identity: Identity, update: Uint8Array) {
    if (update.byteLength > MAX_DOCUMENT_BYTES)
      throw new Error("Update is too large");
    return this.transact(async (client) => {
      const { id, epoch } = parseDocumentName(name);
      const result = await client.query<StoredDocument>(
        "SELECT collaboration_state, title, content_html, revision FROM documents WHERE document_id = $1 AND collaboration_epoch = $2 AND archived_at IS NULL FOR UPDATE",
        [id, epoch],
      );
      if (!result.rows[0]?.collaboration_state)
        throw new Error("Document changed");
      const current = await this.authorize(name, identity.tokenHash, client);
      if (!current.canEdit) throw new Error("Document is read-only");
      const document = new Y.Doc();
      try {
        Y.applyUpdate(document, result.rows[0].collaboration_state);
        Y.applyUpdate(document, update);
        const state = Buffer.from(Y.encodeStateAsUpdate(document));
        if (state.byteLength > MAX_DOCUMENT_BYTES)
          throw new Error("Document is too large");
        if (!state.equals(result.rows[0].collaboration_state)) {
          const { title, html, text } = renderDocument(document);
          await client.query(
            "UPDATE documents SET collaboration_state = $2, title = $3, content_html = $4, content_text = $5, updated_by = $6 WHERE document_id = $1",
            [id, state, title, html, text, current.userId],
          );
        }
        return state;
      } finally {
        document.destroy();
      }
    });
  }
}
