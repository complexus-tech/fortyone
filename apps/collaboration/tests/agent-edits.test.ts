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
import { AgentEditStore } from "../src/agent-edit-store";
import { AgentEditConsumer } from "../src/agent-edit-consumer";
import { DocumentStore, renderDocument } from "../src/documents";
import { createCollaborationServer } from "../src/server";

async function until(check: () => boolean | Promise<boolean>) {
  const deadline = Date.now() + 10_000;
  while (!(await check())) {
    if (Date.now() > deadline)
      throw new Error("Timed out waiting for agent collaboration");
    await new Promise((resolve) => setTimeout(resolve, 25));
  }
}

test(
  "authoritative queued edits are atomic, fenced, idempotent and broadcast to real collaborators",
  { timeout: 60_000 },
  async () => {
    const controlURL = process.env.TEST_DATABASE_URL;
    assert.ok(
      controlURL,
      "TEST_DATABASE_URL must point to disposable local PostgreSQL",
    );
    const parsed = new URL(controlURL);
    assert.ok(
      ["127.0.0.1", "localhost"].includes(parsed.hostname),
      "Only local PostgreSQL is allowed",
    );
    const control = new Pool({ connectionString: controlURL });
    const database = `fortyone_test_agent_${randomBytes(8).toString("hex")}`;
    await control.query(`CREATE DATABASE "${database}"`);
    parsed.pathname = `/${database}`;
    const pool = new Pool({ connectionString: parsed.href });
    const server = createCollaborationServer(new DocumentStore(pool), {
      port: 0,
      origins: ["http://localhost:3000"],
    });
    const store = new AgentEditStore(pool);
    const consumer = new AgentEditConsumer(
      store,
      () => server.hocuspocus.documents,
    );
    const clients: HocuspocusProvider[] = [];
    try {
      const root = fileURLToPath(new URL("../../server/", import.meta.url));
      execFileSync(
        `${root}.tools/bin/migrate`,
        ["-path", `${root}internal/migrations`, "-database", parsed.href, "up"],
        { stdio: "pipe" },
      );
      const workspace = randomUUID();
      const owner = randomUUID();
      const member = randomUUID();
      const documentId = randomUUID();
      const threadId = randomUUID();
      await pool.query(
        "INSERT INTO workspaces(workspace_id,name,slug,trial_ends_on) VALUES($1,'Agent test',$2,clock_timestamp()+interval '1 day')",
        [workspace, database],
      );
      for (const user of [owner, member]) {
        await pool.query(
          "INSERT INTO users(user_id,username,email,full_name,is_active,is_system) VALUES($1,$2,$3,$2,TRUE,FALSE)",
          [user, user, `${user}@example.com`],
        );
        await pool.query(
          "INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')",
          [workspace, user],
        );
      }
      await pool.query(
        "INSERT INTO documents(document_id,workspace_id,title,content_html,content_text,visibility,created_by,updated_by) VALUES($1,$2,'Plan','<p>Original paragraph.</p><p>Keep this.</p>','Original paragraph. Keep this.','workspace',$3,$3)",
        [documentId, workspace, owner],
      );
      await pool.query(
        "INSERT INTO document_comment_threads(thread_id,document_id,workspace_id,quote,anchor_start,anchor_end,created_by) VALUES($1,$2,$3,'Original paragraph.',1,20,$4)",
        [threadId, documentId, workspace, owner],
      );
      const current = async () =>
        (
          await pool.query(
            "SELECT revision, collaboration_epoch, content_html, updated_by FROM documents WHERE document_id=$1",
            [documentId],
          )
        ).rows[0];
      const enqueue = async (
        options: {
          requester?: string;
          operations?: unknown;
          revision?: string;
          epoch?: string;
          expired?: boolean;
          tampered?: boolean;
        } = {},
      ) => {
        const stored = await current();
        const invocation = randomUUID();
        const source = randomUUID();
        const edit = randomUUID();
        const requester = options.requester ?? owner;
        const operations = options.operations ?? {
          version: 1,
          appendParagraphs: ["Agent addition."],
        };
        const revision = options.revision ?? stored.revision;
        const epoch = options.epoch ?? stored.collaboration_epoch;
        const sourceRow = await pool.query(
          "INSERT INTO document_comments(comment_id,thread_id,body,created_by) VALUES($1,$2,'Please update this document',$3) RETURNING CAST(updated_at AS text) AS revision",
          [source, threadId, requester],
        );
        const proposal = {
          summary: "Update document",
          reply: "Updated.",
          requiresApproval: false,
          action: {
            version: 1,
            documentId,
            expectedRevision: Number(revision),
            expectedEpoch: Number(epoch),
            operations,
          },
        };
        await pool.query(
          `INSERT INTO agent_invocations(id,workspace_id,requester_id,team_id,surface,source_id,conversation_id,entity_id,source_revision,source_hash,idempotency_key,request_hash,proposal,action_started_at)
        VALUES($1,$2,$3,NULL,'document_comment',$4,$5,$6,$10,$11,$7,$8,CAST($9 AS jsonb),clock_timestamp())`,
          [
            invocation,
            workspace,
            requester,
            source,
            threadId,
            documentId,
            invocation,
            Buffer.alloc(32),
            JSON.stringify(proposal),
            sourceRow.rows[0].revision,
            createHash("sha256")
              .update("Please update this document")
              .digest("hex"),
          ],
        );
        await pool.query(
          `INSERT INTO document_agent_edits(id,invocation_id,workspace_id,document_id,requester_id,expected_revision,expected_epoch,operations,created_at)
        VALUES($1,$2,$3,$4,$5,$6,$7,CAST($8 AS jsonb),clock_timestamp() - CAST($9 AS interval))`,
          [
            edit,
            invocation,
            workspace,
            documentId,
            requester,
            revision,
            epoch,
            JSON.stringify(
              options.tampered ? { version: 1, title: "Tampered" } : operations,
            ),
            options.expired ? "16 minutes" : "0 seconds",
          ],
        );
        return { id: edit, invocation_id: invocation };
      };
      const result = async (id: string) =>
        (
          await pool.query(
            "SELECT status,error_code,receipt FROM document_agent_edits WHERE id=$1",
            [id],
          )
        ).rows[0];
      const legacy = await enqueue();
      await store.apply(legacy.id, legacy.invocation_id);
      assert.equal(
        (await result(legacy.id)).error_code,
        "collaboration_unavailable",
      );
      const legacyState = await pool.query(
        "SELECT collaboration_state FROM documents WHERE document_id=$1",
        [documentId],
      );
      assert.equal(
        legacyState.rows[0].collaboration_state,
        null,
        "agent edits must never activate a legacy document",
      );
      await server.listen();
      const url = `ws://127.0.0.1:${server.address.port}`;
      class LocalSocket extends WebSocket {
        constructor(address: string) {
          super(address, { origin: "http://localhost:3000" });
        }
      }
      for (const user of [owner, member]) {
        const raw = randomBytes(32).toString("hex");
        await pool.query(
          "INSERT INTO document_collaboration_sessions(token_hash,document_id,user_id,session_version,epoch,expires_at) SELECT $1,$2,user_id,auth_session_version,1,NOW()+interval '1 hour' FROM users WHERE user_id=$3",
          [createHash("sha256").update(raw).digest("hex"), documentId, user],
        );
        const provider = new HocuspocusProvider({
          url,
          name: `${documentId}:1`,
          token: raw,
          document: new Y.Doc(),
          websocketProvider: new HocuspocusProviderWebsocket({
            url,
            WebSocketPolyfill: LocalSocket,
          }),
        });
        provider.attach();
        clients.push(provider);
      }
      await until(() => clients.every((client) => client.isSynced));
      const before = await current();
      const applied = await enqueue({
        operations: {
          version: 1,
          replacements: [{ before: "Original", after: "Updated" }],
          appendParagraphs: ["Agent addition."],
        },
      });
      await consumer.tick();
      await until(() =>
        clients.every((client) =>
          renderDocument(client.document).text.includes("Agent addition."),
        ),
      );
      const receipt = await result(applied.id);
      assert.equal(receipt.status, "applied");
      assert.equal(receipt.receipt.requesterId, owner);
      assert.equal(receipt.receipt.beforeRevision, Number(before.revision));
      const after = await current();
      assert.equal(after.updated_by, owner);
      assert.match(after.content_html, /Updated paragraph/);
      assert.equal(after.collaboration_epoch, before.collaboration_epoch);
      const audit = await pool.query(
        "SELECT edited_by,content_html FROM document_revisions WHERE document_id=$1 ORDER BY revision DESC LIMIT 1",
        [documentId],
      );
      assert.equal(audit.rows[0].edited_by, owner);
      assert.match(audit.rows[0].content_html, /Agent addition/);
      await store.apply(applied.id, applied.invocation_id);
      assert.equal(
        (await current()).revision,
        after.revision,
        "receipt replay must not write twice",
      );

      // A crash after commit must not strand connected editors on old state.
      const missedBroadcast = await enqueue({
        operations: { version: 1, appendParagraphs: ["Recovered broadcast."] },
      });
      await store.apply(missedBroadcast.id, missedBroadcast.invocation_id);
      await consumer.tick();
      await until(() =>
        clients.every((client) =>
          renderDocument(client.document).text.includes("Recovered broadcast."),
        ),
      );
      const stable = await current();
      const noOp = await enqueue({
        operations: {
          version: 1,
          replacements: [{ before: "Keep this.", after: "Keep this." }],
        },
      });
      await store.apply(noOp.id, noOp.invocation_id);
      assert.equal((await result(noOp.id)).status, "applied");
      assert.deepEqual((await result(noOp.id)).receipt.changedFields, []);
      assert.equal((await current()).revision, stable.revision);
      const partial = await enqueue({
        operations: {
          version: 1,
          replacements: [
            { before: "Keep this.", after: "Should roll back." },
            { before: "Missing text", after: "Other" },
          ],
        },
      });
      await store.apply(partial.id, partial.invocation_id);
      assert.equal((await result(partial.id)).status, "conflict");
      assert.equal((await current()).content_html, stable.content_html);
      const concurrentA = await enqueue({
        operations: { version: 1, appendParagraphs: ["Concurrent accepted."] },
      });
      const concurrentB = await enqueue({
        operations: { version: 1, appendParagraphs: ["Concurrent accepted."] },
      });
      await Promise.all([
        store.apply(concurrentA.id, concurrentA.invocation_id),
        store.apply(concurrentB.id, concurrentB.invocation_id),
      ]);
      assert.deepEqual(
        [
          (await result(concurrentA.id)).status,
          (await result(concurrentB.id)).status,
        ].sort(),
        ["applied", "conflict"],
      );
      await consumer.tick();
      await until(() =>
        clients.every((client) =>
          renderDocument(client.document).text.includes("Concurrent accepted."),
        ),
      );

      const stale = await enqueue();
      const text = (
        clients[0].document.getXmlFragment("default").get(1) as Y.XmlElement
      ).get(0) as Y.XmlText;
      text.insert(0, "Human: ");
      await until(() => !clients[0].hasUnsyncedChanges);
      await store.apply(stale.id, stale.invocation_id);
      assert.equal((await result(stale.id)).status, "conflict");
      assert.match((await current()).content_html, /Human: Keep this/);

      // Human body atoms persist through Yjs, index related work, and notify once.
      const teamId = randomUUID();
      const storyId = randomUUID();
      const objectiveId = randomUUID();
      await pool.query(
        "INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'Team','TST','blue')",
        [teamId, workspace],
      );
      await pool.query(
        "INSERT INTO team_members(team_id,user_id) VALUES($1,$2)",
        [teamId, owner],
      );
      await pool.query(
        "INSERT INTO stories(id,workspace_id,team_id,sequence_id,title,reporter_id) VALUES($1,$2,$3,1,'Linked story',$4)",
        [storyId, workspace, teamId, owner],
      );
      await pool.query(
        "INSERT INTO objectives(objective_id,workspace_id,team_id,sequence_id,name,created_by) VALUES($1,$2,$3,1,'Linked objective',$4)",
        [objectiveId, workspace, teamId, owner],
      );
      const atom = (name: string, attrs: Record<string, string>) => {
        const node = new Y.XmlElement(name);
        for (const [key, value] of Object.entries(attrs))
          node.setAttribute(key, value);
        return node;
      };
      const references = new Y.XmlElement("paragraph");
      references.insert(0, [
        atom("mention", { id: member, label: "Colleague" }),
        atom("documentReference", {
          id: storyId,
          entityType: "story",
          label: "Linked story",
          displayId: "TST-1",
        }),
        atom("documentReference", {
          id: storyId,
          entityType: "story",
          label: "Linked story again",
          displayId: "TST-1",
        }),
        atom("documentReference", {
          id: objectiveId,
          entityType: "objective",
          label: "Linked objective",
        }),
      ]);
      const body = clients[0].document.getXmlFragment("default");
      body.insert(body.length, [references]);
      await until(() => !clients[0].hasUnsyncedChanges);
      await until(() =>
        renderDocument(clients[1].document).mentions.includes(member),
      );
      const relationships = async () =>
        (
          await pool.query(
            "SELECT entity_type,entity_id,source FROM document_relationships WHERE document_id=$1 ORDER BY entity_type",
            [documentId],
          )
        ).rows;
      assert.deepEqual(await relationships(), [
        { entity_type: "objective", entity_id: objectiveId, source: "body" },
        { entity_type: "story", entity_id: storyId, source: "body" },
      ]);
      const mentionCount = async () =>
        (
          await pool.query(
            "SELECT count(*) FROM notifications WHERE entity_id=$1 AND recipient_id=$2 AND type='mention'",
            [documentId, member],
          )
        ).rows[0].count;
      assert.equal(await mentionCount(), "1");
      const delivery = await pool.query(
        "SELECT count(*) FROM document_notification_deliveries AS delivery JOIN notifications AS notification ON notification.notification_id=delivery.notification_id WHERE notification.entity_id=$1 AND delivery.delivered_at IS NULL",
        [documentId],
      );
      assert.equal(
        delivery.rows[0].count,
        "1",
        "body mentions must enqueue durable notification delivery",
      );
      text.insert(0, "More context. ");
      await until(() => !clients[0].hasUnsyncedChanges);
      assert.equal(
        await mentionCount(),
        "1",
        "ordinary prose edits must not notify again",
      );
      references.delete(1, 1);
      await until(() => !clients[0].hasUnsyncedChanges);
      assert.equal(
        (await relationships()).length,
        2,
        "the last inline occurrence owns removal",
      );
      await pool.query(
        "UPDATE document_relationships SET source='manual' WHERE document_id=$1 AND entity_id=$2",
        [documentId, objectiveId],
      );
      body.delete(body.length - 1, 1);
      await until(() => !clients[0].hasUnsyncedChanges);
      assert.deepEqual(await relationships(), [
        { entity_type: "objective", entity_id: objectiveId, source: "manual" },
      ]);
      assert.equal(
        (
          await pool.query(
            "SELECT count(*) FROM document_body_mentions WHERE document_id=$1",
            [documentId],
          )
        ).rows[0].count,
        "0",
      );
      assert.equal(await mentionCount(), "1");

      const changedSource = await enqueue();
      await pool.query(
        "UPDATE document_comments SET body='Changed request' WHERE comment_id=(SELECT source_id FROM agent_invocations WHERE id=$1)",
        [changedSource.invocation_id],
      );
      await store.apply(changedSource.id, changedSource.invocation_id);
      assert.equal((await result(changedSource.id)).status, "forbidden");
      const unapproved = await enqueue();
      await pool.query(
        "UPDATE agent_invocations SET proposal=jsonb_set(proposal,'{requiresApproval}','true') WHERE id=$1",
        [unapproved.invocation_id],
      );
      await store.apply(unapproved.id, unapproved.invocation_id);
      assert.equal((await result(unapproved.id)).status, "forbidden");

      const privateAccess = await enqueue({ requester: member });
      await pool.query(
        "INSERT INTO document_members(document_id,user_id,role) VALUES($1,$2,'editor')",
        [documentId, member],
      );
      await pool.query(
        "UPDATE documents SET visibility='private' WHERE document_id=$1",
        [documentId],
      );
      await store.apply(privateAccess.id, privateAccess.invocation_id);
      assert.equal(
        (await result(privateAccess.id)).error_code,
        "edit_permission_revoked",
        "a stale editor grant cannot bypass private visibility",
      );
      await pool.query(
        "UPDATE documents SET visibility='workspace' WHERE document_id=$1",
        [documentId],
      );
      const revoked = await enqueue({ requester: member });
      await pool.query(
        "DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2",
        [workspace, member],
      );
      await store.apply(revoked.id, revoked.invocation_id);
      assert.equal((await result(revoked.id)).status, "forbidden");
      const tampered = await enqueue({ tampered: true });
      await store.apply(tampered.id, tampered.invocation_id);
      assert.equal((await result(tampered.id)).status, "forbidden");
      const expired = await enqueue({ expired: true });
      await store.apply(expired.id, expired.invocation_id);
      assert.equal((await result(expired.id)).error_code, "expired");
      const staleEpoch = await enqueue({ epoch: "2" });
      await store.apply(staleEpoch.id, staleEpoch.invocation_id);
      assert.equal((await result(staleEpoch.id)).status, "conflict");
      const trialRevoked = await enqueue();
      await pool.query(
        "UPDATE workspaces SET trial_ends_on=clock_timestamp()-interval '1 day' WHERE workspace_id=$1",
        [workspace],
      );
      await store.apply(trialRevoked.id, trialRevoked.invocation_id);
      assert.equal(
        (await result(trialRevoked.id)).error_code,
        "maya_access_revoked",
      );
      await pool.query(
        "INSERT INTO workspace_subscriptions(workspace_id,stripe_customer_id,stripe_subscription_id,subscription_status,subscription_tier) VALUES($1,'customer-test','subscription-test','active','pro')",
        [workspace],
      );
      const paid = await enqueue({
        operations: { version: 1, title: "Paid workspace plan" },
      });
      await store.apply(paid.id, paid.invocation_id);
      assert.equal((await result(paid.id)).status, "applied");
      const paidRevoked = await enqueue();
      await pool.query(
        "UPDATE workspace_subscriptions SET subscription_status='canceled' WHERE workspace_id=$1",
        [workspace],
      );
      await store.apply(paidRevoked.id, paidRevoked.invocation_id);
      assert.equal(
        (await result(paidRevoked.id)).error_code,
        "maya_access_revoked",
      );
      const heartbeat = await pool.query(
        "SELECT count(*) FROM document_agent_consumer_health WHERE expires_at > clock_timestamp()",
      );
      assert.equal(heartbeat.rows[0].count, "1");
    } finally {
      await consumer.stop().catch(() => undefined);
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
