import { strict as assert } from "node:assert";
import { execFileSync } from "node:child_process";
import { createHash, randomBytes, randomUUID } from "node:crypto";
import { fileURLToPath } from "node:url";
import { test } from "node:test";
import { Pool } from "pg";
import WebSocket from "ws";
import {
  HocuspocusProvider,
  HocuspocusProviderWebsocket,
} from "@hocuspocus/provider";
import * as Y from "yjs";
import { DocumentStore, renderDocument } from "../src/documents";
import { createCollaborationServer } from "../src/server";

async function until(check: () => boolean | Promise<boolean>) {
  const deadline = Date.now() + 10_000;
  while (!(await check())) {
    if (Date.now() > deadline)
      throw new Error("Timed out waiting for collaboration");
    await new Promise((resolve) => setTimeout(resolve, 25));
  }
}

test(
  "real clients merge durably and revoked or restored sessions cannot write",
  { timeout: 60_000 },
  async () => {
    const controlURL = process.env.TEST_DATABASE_URL;
    assert.ok(
      controlURL,
      "TEST_DATABASE_URL must point to a disposable local PostgreSQL server",
    );
    const parsed = new URL(controlURL);
    assert.ok(
      ["127.0.0.1", "localhost"].includes(parsed.hostname),
      "Integration harness only accepts local PostgreSQL",
    );
    const control = new Pool({ connectionString: controlURL });
    const database = `fortyone_test_collab_${randomBytes(8).toString("hex")}`;
    await control.query(`CREATE DATABASE "${database}"`);
    parsed.pathname = `/${database}`;
    const pool = new Pool({ connectionString: parsed.href });
    const clients: HocuspocusProvider[] = [];
    const store = new DocumentStore(pool);
    const server = createCollaborationServer(store, {
      port: 0,
      origins: ["http://localhost:3000"],
    });
    try {
      const root = fileURLToPath(new URL("../../server/", import.meta.url));
      execFileSync(
        `${root}.tools/bin/migrate`,
        ["-path", `${root}internal/migrations`, "-database", parsed.href, "up"],
        { stdio: "pipe" },
      );
      const workspace = randomUUID();
      const owner = randomUUID();
      const editor = randomUUID();
      const viewer = randomUUID();
      const docId = randomUUID();
      await pool.query(
        "INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Collaboration test',$2)",
        [workspace, database],
      );
      for (const [id, role] of [
        [owner, "admin"],
        [editor, "member"],
        [viewer, "guest"],
      ]) {
        await pool.query(
          "INSERT INTO users(user_id,username,email,full_name,is_active,is_system) VALUES($1,$2,$3,$2,TRUE,FALSE)",
          [id, id, `${id}@example.com`],
        );
        await pool.query(
          "INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,CAST($3 AS public.user_role))",
          [workspace, id, role],
        );
      }
      await pool.query(
        "INSERT INTO documents(document_id,workspace_id,title,content_html,content_text,visibility,created_by,updated_by) VALUES($1,$2,'Plan','<p>First</p><p>Last</p>','First Last','workspace',$3,$3)",
        [docId, workspace, owner],
      );
      const name = `${docId}:1`;
      async function token(user: string) {
        const raw = randomBytes(32).toString("hex");
        await pool.query(
          "INSERT INTO document_collaboration_sessions(token_hash,document_id,user_id,session_version,epoch,expires_at) SELECT $1,$2,user_id,auth_session_version,1,NOW()+INTERVAL '1 hour' FROM users WHERE user_id=$3",
          [createHash("sha256").update(raw).digest("hex"), docId, user],
        );
        return raw;
      }
      const ownerToken = await token(owner);
      const editorToken = await token(editor);
      const viewerToken = await token(viewer);
      const identity = await store.authenticate(name, ownerToken);
      // Two first connections import the HTML exactly once.
      const seeds = await Promise.all([
        store.load(name, identity),
        store.load(name, identity),
      ]);
      assert.deepEqual(seeds[0], seeds[1]);
      await server.listen();
      const url = `ws://127.0.0.1:${server.address.port}`;
      class LocalSocket extends WebSocket {
        constructor(address: string) {
          super(address, { origin: "http://localhost:3000" });
        }
      }
      function connect(raw: string) {
        const provider = new HocuspocusProvider({
          url,
          name,
          token: raw,
          document: new Y.Doc(),
          websocketProvider: new HocuspocusProviderWebsocket({
            url,
            WebSocketPolyfill: LocalSocket,
          }),
        });
        provider.attach();
        clients.push(provider);
        return provider;
      }
      const alice = connect(ownerToken);
      const bob = connect(editorToken);
      const reader = connect(viewerToken);
      await until(() => alice.isSynced && bob.isSynced && reader.isSynced);
      assert.match(renderDocument(alice.document).text, /First/);
      assert.equal(reader.authorizedScope, "readonly");
      assert.equal(alice.authorizedScope, "read-write");
      const prepend = (
        provider: HocuspocusProvider,
        index: number,
        text: string,
      ) => {
        const paragraph = provider.document
          .getXmlFragment("default")
          .get(index) as Y.XmlElement;
        (paragraph.get(0) as Y.XmlText).insert(0, text);
      };
      prepend(alice, 0, "Alice: ");
      prepend(bob, 1, "Bob: ");
      await until(
        () =>
          !alice.hasUnsyncedChanges &&
          !bob.hasUnsyncedChanges &&
          renderDocument(reader.document).text.includes("Bob: "),
      );
      const stored = await pool.query(
        "SELECT content_html,revision FROM documents WHERE document_id=$1",
        [docId],
      );
      assert.match(stored.rows[0].content_html, /Alice: First/);
      assert.match(stored.rows[0].content_html, /Bob: Last/);
      assert.equal(
        renderDocument(alice.document).html,
        renderDocument(bob.document).html,
      );
      // A read-only client must fail even when it bypasses the editor UI.
      const readerIdentity = await store.authenticate(name, viewerToken);
      await assert.rejects(
        () =>
          store.apply(
            name,
            readerIdentity,
            Y.encodeStateAsUpdate(alice.document),
          ),
        /read-only/,
      );
      prepend(reader, 0, "UNAUTHORIZED ");
      await new Promise((resolve) => setTimeout(resolve, 150));
      assert.doesNotMatch(
        (
          await pool.query(
            "SELECT content_html FROM documents WHERE document_id=$1",
            [docId],
          )
        ).rows[0].content_html,
        /UNAUTHORIZED/,
      );
      await pool.query(
        "DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2",
        [workspace, editor],
      );
      await assert.rejects(
        () => store.authenticate(name, editorToken),
        /revoked/,
      );
      // Restore invalidates the entire old CRDT, including a late/offline update.
      const stale = Y.encodeStateAsUpdate(alice.document);
      await pool.query(
        "UPDATE documents SET collaboration_state=NULL,collaboration_epoch=collaboration_epoch+1,title='Restored',content_html='<p>Original</p>',content_text='Original' WHERE document_id=$1",
        [docId],
      );
      await assert.rejects(() => store.apply(name, identity, stale), /changed/);
      await assert.rejects(
        () => store.authenticate(name, ownerToken),
        /revoked/,
      );
      const revisions = await pool.query(
        "SELECT title FROM document_revisions WHERE document_id=$1 ORDER BY revision DESC",
        [docId],
      );
      assert.equal(revisions.rows[0].title, "Restored");
      assert.ok(revisions.rowCount! >= 3);
    } finally {
      for (const client of clients) {
        client.destroy();
        client.configuration.websocketProvider.destroy();
        client.document.destroy();
      }
      await server.destroy();
      await pool.end();
      await control.query(`DROP DATABASE "${database}" WITH (FORCE)`);
      await control.end();
    }
  },
);
