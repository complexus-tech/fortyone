import { useState } from "react";
import type { Document, WorkspaceData } from "../types";
import { useCollaboration } from "../hooks/use-collaboration";
import { flushDrafts } from "../drafts";
import { errorMessage } from "../native";
import { Dialog, ErrorNotice, Spinner } from "./primitives";
import { AttachmentList } from "./attachment-list";
import { RichText } from "./rich-text";

export function CollaborativeDocumentEditor({
  document,
  data,
  onClose,
}: {
  document: Document;
  data: WorkspaceData;
  onClose: () => void;
}) {
  const workspaceId = data.workspaceId;
  const collaboration = useCollaboration(document, workspaceId);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const run = async (task: () => Promise<void>) => {
    setBusy(true);
    setError(null);
    try {
      await task();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };
  const close = async () => {
    await flushDrafts();
    onClose();
  };
  return (
    <Dialog
      title="Collaborative document"
      wide
      onClose={() => {
        void run(close);
      }}
    >
      <div className="editor-body document-editor">
        <ErrorNotice message={error ?? collaboration.error} />
        {collaboration.blocked ? (
          <div className="info-notice">
            Your local document is preserved. Export a copy to keep these edits
            before joining a restored version.
          </div>
        ) : null}
        {!collaboration.ready || !collaboration.connection ? (
          <p role="status">
            <Spinner />
            {collaboration.status}
          </p>
        ) : (
          <>
            <label className="sr-only" htmlFor="document-title">
              Document title
            </label>
            <input
              id="document-title"
              className="title-input"
              value={collaboration.title}
              readOnly={!collaboration.writable}
              onChange={(event) => collaboration.setTitle(event.target.value)}
            />
            <RichText
              key={collaboration.connection.name}
              mediaScope={{
                workspaceId,
                resourceId: document.id,
                kind: "document",
              }}
              html=""
              collaborationDocument={collaboration.connection.document}
              readOnly={!collaboration.writable}
              label="Document content"
              onChange={collaboration.updateContent}
            />
            <AttachmentList
              data={data}
              resourceId={document.id}
              kind="document"
            />
          </>
        )}
      </div>
      <footer className="dialog-footer">
        <span className="save-status" role="status">
          {collaboration.status}
        </span>
        {collaboration.error ? (
          <button disabled={busy} onClick={collaboration.retryConnection}>
            Retry connection
          </button>
        ) : null}
        {collaboration.error ? (
          <button disabled={busy} onClick={collaboration.retryLocalSave}>
            Retry local save
          </button>
        ) : null}
        {collaboration.ready ? (
          <button
            disabled={busy}
            onClick={() => {
              void run(collaboration.export);
            }}
          >
            Export document
          </button>
        ) : null}
        {collaboration.blocked && collaboration.exported ? (
          <button
            disabled={busy}
            onClick={() => {
              void run(async () => {
                await collaboration.discardAfterExport();
                onClose();
              });
            }}
          >
            Remove exported local draft
          </button>
        ) : (
          <button
            className="primary-button"
            disabled={busy}
            onClick={() => {
              void run(close);
            }}
          >
            Done
          </button>
        )}
      </footer>
    </Dialog>
  );
}
