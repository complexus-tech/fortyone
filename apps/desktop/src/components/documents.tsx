import { lazy, Suspense, useDeferredValue, useState } from "react";
import { DocsIcon } from "icons/src/docs";
import { PlusIcon } from "icons/src/plus-icon";
import { SearchIcon } from "icons/src/search";
import type {
  Document,
  JsonRecord,
  OperationInput,
  WorkspaceData,
} from "../types";
import { useDraft } from "../hooks/use-draft";
import { flushDrafts } from "../drafts";
import { errorMessage } from "../native";
import { formatTime, resourceBaseVersion } from "../model";
import { AttachmentList } from "./attachment-list";
import { Dialog, Empty, ErrorNotice, Spinner } from "./primitives";

const RichText = lazy(() =>
  import("./rich-text").then((module) => ({ default: module.RichText })),
);
export function Documents({
  data,
  onOpen,
  onCreate,
}: {
  data: WorkspaceData;
  onOpen: (document: Document) => void;
  onCreate: () => void;
}) {
  const [query, setQuery] = useState("");
  const deferred = useDeferredValue(query).trim().toLocaleLowerCase();
  const documents = data.documents.filter((document) =>
    `${document.title} ${document.contentText}`
      .toLocaleLowerCase()
      .includes(deferred),
  );
  return (
    <section className="collection">
      <div className="page-heading">
        <div>
          <span className="eyebrow">Ideas that stay with you</span>
          <h1>Documents</h1>
          <p>Your team’s knowledge, available on this device.</p>
        </div>
        <button className="primary-button" onClick={onCreate}>
          <PlusIcon />
          New document
        </button>
      </div>
      <div className="collection-toolbar">
        <label className="search-field">
          <SearchIcon />
          <input
            id="desktop-search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search documents…"
            aria-label="Search documents"
          />
        </label>
        <span className="item-count">{documents.length} documents</span>
      </div>
      {!documents.length ? (
        <Empty
          title={query ? "No matching documents" : "Make room for your ideas"}
        >
          {query
            ? "Try another title or phrase."
            : "Create a document or sync to download your workspace’s knowledge."}
        </Empty>
      ) : (
        <div className="documents-list">
          {documents.map((document) => (
            <button
              className="document-row"
              key={document.id}
              onClick={() => onOpen(document)}
            >
              <DocsIcon />
              <div>
                <strong>{document.title || "Untitled"}</strong>
                <span>
                  {document.contentText?.slice(0, 120) || "Open document"}
                </span>
              </div>
              <time>{formatTime(document.updatedAt)}</time>
              {!document.canEdit ? <small>Read only</small> : null}
            </button>
          ))}
        </div>
      )}
    </section>
  );
}

const CollaborativeDocumentEditor = lazy(() =>
  import("./collaborative-document").then((module) => ({
    default: module.CollaborativeDocumentEditor,
  })),
);
export function DocumentEditor(props: {
  document: Document | null;
  data: WorkspaceData;
  save: (input: OperationInput) => Promise<void>;
  onClose: () => void;
}) {
  return props.document?.collaborative === true ? (
    <Suspense fallback={<p role="status">Opening collaborative document…</p>}>
      <CollaborativeDocumentEditor
        document={props.document}
        data={props.data}
        onClose={props.onClose}
      />
    </Suspense>
  ) : (
    <StandardDocumentEditor {...props} />
  );
}

type DocumentDraft = JsonRecord & {
  resourceId: string;
  title: string;
  contentHtml: string;
  contentText: string;
  baseVersion: string | null;
  collaborationEpoch: number;
};
function StandardDocumentEditor({
  document,
  data,
  save,
  onClose,
}: {
  document: Document | null;
  data: WorkspaceData;
  save: (input: OperationInput) => Promise<void>;
  onClose: () => void;
}) {
  const draftKey = document?.id ?? "new-document";
  const draft = useDraft<DocumentDraft>(
    data.workspaceId,
    draftKey,
    "document",
    {
      resourceId: document?.id ?? crypto.randomUUID(),
      title: document?.title ?? "",
      contentHtml: document?.contentHtml ?? "",
      contentText: document?.contentText ?? "",
      baseVersion: document
        ? resourceBaseVersion(
            "document",
            document.id,
            document.revision,
            data.pending,
          )
        : null,
      collaborationEpoch: document?.collaborationEpoch ?? 0,
    },
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const readOnly = document?.canEdit === false;
  const close = async () => {
    try {
      await flushDrafts();
      onClose();
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };
  const submit = async () => {
    setBusy(true);
    setError(null);
    try {
      const { resourceId, baseVersion, ...payload } = draft.value;
      await save({
        workspaceId: data.workspaceId,
        resourceId,
        kind: document ? "document.update" : "document.create",
        consumedDraft: { resourceId: draftKey, kind: "document" },
        baseVersion,
        payload: document
          ? payload
          : {
              title: payload.title,
              contentHtml: payload.contentHtml,
              contentText: payload.contentText,
              visibility: "workspace",
            },
      });
      onClose();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };
  return (
    <Dialog
      title={document ? "Document" : "New document"}
      onClose={() => {
        void close();
      }}
      wide
    >
      <div className="editor-body document-editor">
        <ErrorNotice message={error ?? draft.error} />
        {!draft.ready ? (
          <p role="status">
            <Spinner />
            Recovering your saved document…
          </p>
        ) : (
          <>
            <label className="sr-only" htmlFor="document-title">
              Document title
            </label>
            <input
              id="document-title"
              className="title-input"
              placeholder="Untitled document"
              value={draft.value.title}
              readOnly={readOnly || busy}
              onChange={(event) =>
                draft.update({ ...draft.value, title: event.target.value })
              }
              autoFocus
            />
            <Suspense fallback={<p>Loading editor…</p>}>
              <RichText
                mediaScope={{
                  workspaceId: data.workspaceId,
                  resourceId: draft.value.resourceId,
                  kind: "document",
                }}
                html={draft.value.contentHtml}
                readOnly={readOnly || busy}
                label="Document content"
                onChange={(contentHtml, contentText) =>
                  draft.update({ ...draft.value, contentHtml, contentText })
                }
              />
            </Suspense>
            {document ? (
              <AttachmentList
                data={data}
                resourceId={document.id}
                kind="document"
              />
            ) : null}
          </>
        )}
      </div>
      <footer className="dialog-footer">
        <span className="save-status" role="status">
          {readOnly ? "You have view access" : draft.status}
        </span>
        {draft.error ? (
          <button
            onClick={() =>
              draft.ready ? draft.update(draft.value) : draft.retryRead()
            }
          >
            {draft.ready ? "Retry local save" : "Retry draft recovery"}
          </button>
        ) : null}
        {!readOnly ? (
          <button
            className="primary-button"
            disabled={!draft.ready || busy || !draft.value.title.trim()}
            onClick={() => {
              void submit();
            }}
          >
            {busy ? <Spinner /> : null}
            {document ? "Save changes" : "Create document"}
          </button>
        ) : null}
      </footer>
    </Dialog>
  );
}
