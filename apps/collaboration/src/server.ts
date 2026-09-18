import { Server } from "@hocuspocus/server";
import { Redis } from "@hocuspocus/extension-redis";
import * as Y from "yjs";
import { isAllowedOrigin } from "./origins";
import { DocumentStore, MAX_DOCUMENT_BYTES, type Identity } from "./documents";

export function createCollaborationServer(
  store: DocumentStore,
  options: { port: number; origins: string[]; redis?: Redis },
) {
  const checks = new Map<string, ReturnType<typeof setInterval>>();
  return new Server<Identity>({
    port: options.port,
    stopOnSignals: false,
    quiet: true,
    maxPendingDocuments: 4,
    maxUnauthenticatedQueueSize: MAX_DOCUMENT_BYTES,
    maxUnauthenticatedQueueMessages: 30,
    websocketOptions: { maxPayload: MAX_DOCUMENT_BYTES },
    extensions: options.redis ? [options.redis] : [],
    async onRequest({ request, response }) {
      let healthy = false;
      if (request.url === "/healthz") {
        try {
          await store.ping();
          healthy = true;
        } catch {
          /* Report readiness without exposing database errors. */
        }
      }
      response.writeHead(
        request.url !== "/healthz" ? 404 : healthy ? 200 : 503,
        { "Content-Type": "text/plain", "Cache-Control": "no-store" },
      );
      response.end(healthy ? "ready" : "unavailable");
      // Hocuspocus uses an empty rejection to signal a handled HTTP request.
      throw null;
    },
    async onConnect({ requestHeaders }) {
      if (!isAllowedOrigin(requestHeaders.get("origin") ?? "", options.origins))
        throw new Error("Origin is not allowed");
    },
    async onAuthenticate({ documentName, token, connectionConfig }) {
      const identity = await store.authenticate(documentName, token);
      connectionConfig.readOnly = !identity.canEdit;
      return identity;
    },
    async onLoadDocument({ documentName, context, document }) {
      Y.applyUpdate(document, await store.load(documentName, context));
    },
    async beforeHandleMessage({ documentName, context, connection }) {
      const current = await store.authorize(documentName, context.tokenHash);
      connection.readOnly = !current.canEdit;
    },
    async beforeSync({
      type,
      payload,
      documentName,
      context,
      connection,
      document,
    }) {
      if (type === 0 || connection.readOnly) return;
      // Persist before Hocuspocus applies/broadcasts/acknowledges the update.
      // Merging the returned state also catches concurrent writers on other nodes.
      const state = await store.apply(documentName, context, payload);
      Y.applyUpdate(document, state);
    },
    async connected({ connection, documentName, context, socketId }) {
      let checking = false;
      const timer = setInterval(async () => {
        if (checking) return;
        checking = true;
        try {
          const current = await store.authorize(
            documentName,
            context.tokenHash,
          );
          if (current.canEdit !== !connection.readOnly)
            connection.close({ code: 4403, reason: "Document access changed" });
        } catch {
          connection.close({ code: 4403, reason: "Document access changed" });
        } finally {
          checking = false;
        }
      }, 5_000);
      timer.unref();
      checks.set(`${socketId}:${documentName}`, timer);
    },
    async onDisconnect({ socketId, documentName }) {
      const key = `${socketId}:${documentName}`;
      clearInterval(checks.get(key));
      checks.delete(key);
    },
    async onStateless({ connection }) {
      // Clients cannot relay arbitrary messages to other viewers.
      connection.close({ code: 4400, reason: "Unsupported message" });
    },
  });
}
