import { useAccountNative } from "../hooks/use-account-native";
import { useState } from "react";
import type { WorkspaceData } from "../types";
import { errorMessage } from "../native";
import { ErrorNotice, Spinner } from "./primitives";

export function AttachmentList({
  data,
  resourceId,
  kind,
}: {
  data: WorkspaceData;
  resourceId: string;
  kind: "story" | "document";
}) {
  const client = useAccountNative();
  const [busy, setBusy] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const uploaded = (data.attachments ?? []).filter(
    (file) => file.resourceId === resourceId && file.targetKind === kind,
  );
  const pending = data.pending.filter(
    (operation) =>
      operation.kind === "attachment.upload" &&
      operation.resourceId === resourceId &&
      operation.payload.targetKind === kind,
  );
  if (!uploaded.length && !pending.length) return null;
  const download = async (attachmentId: string) => {
    setBusy(attachmentId);
    setError(null);
    setMessage(null);
    try {
      const saved = await client.exportAttachment({
        workspaceId: data.workspaceId,
        resourceId,
        kind,
        attachmentId,
      });
      if (saved) setMessage("Attachment saved.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(null);
    }
  };
  return (
    <section className="attachment-list" aria-label="Attachments">
      <h2>Attachments</h2>
      <ErrorNotice message={error} />
      {message ? <p role="status">{message}</p> : null}
      {uploaded.map((file) => (
        <div className="attachment-row" key={file.id}>
          <span>
            <strong>{file.filename}</strong>
            <small>{formatBytes(file.size)}</small>
          </span>
          <button
            disabled={busy !== null}
            onClick={() => {
              void download(file.id);
            }}
          >
            {busy === file.id ? <Spinner /> : null} Save a copy
          </button>
        </div>
      ))}
      {pending.map((operation) => (
        <div className="attachment-row" key={operation.id}>
          <span>
            <strong>
              {String(operation.payload.fileName ?? "Attachment")}
            </strong>
            <small>{formatBytes(Number(operation.payload.size ?? 0))}</small>
          </span>
          <span className="save-status">
            {operation.status === "pending"
              ? "Saved on this device · waiting to upload"
              : operation.error ?? "Upload needs attention"}
          </span>
        </div>
      ))}
    </section>
  );
}

function formatBytes(bytes: number) {
  return bytes < 1024 * 1024
    ? `${Math.max(1, Math.round(bytes / 1024))} KB`
    : `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
