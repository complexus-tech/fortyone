import type { NodeViewProps } from "@tiptap/react";
import { NodeViewWrapper, ReactNodeViewRenderer } from "@tiptap/react";
import { DocumentMention, DocumentReference } from "@fortyone/document-editor";
import Link from "next/link";
import { ObjectiveIcon, StoryIcon } from "icons";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { hexToRgba } from "@/utils/color";
import { useDocument } from "./hooks";

function PersonReference({ node }: NodeViewProps) {
  const { withWorkspace } = useWorkspacePath();
  return (
    <NodeViewWrapper as="span" className="inline" contentEditable={false}>
      <Link
        className="mention"
        data-type="mention"
        href={withWorkspace(
          `/profile/${encodeURIComponent(String(node.attrs.id))}`,
        )}
      >
        @{String(node.attrs.label)}
      </Link>
    </NodeViewWrapper>
  );
}

function WorkReference({ node, extension }: NodeViewProps) {
  const { withWorkspace } = useWorkspacePath();
  const { data: document } = useDocument(
    (extension.options as { documentId: string }).documentId,
  );
  const work = document?.relatedWork.find(
    (item) =>
      item.entityId === node.attrs.id &&
      item.entityType === node.attrs.entityType,
  );
  const story = node.attrs.entityType === "story";
  const displayId = work?.reference ?? node.attrs.displayId;
  const title = work?.title ?? node.attrs.label;
  const statusColor = story ? work?.statusColor : undefined;
  const style = statusColor
    ? { backgroundColor: hexToRgba(statusColor, 0.1) }
    : undefined;
  let path: string | null = null;
  if (story && displayId)
    path = `/work/${encodeURIComponent(String(displayId))}`;
  else if (work?.teamId)
    path = `/teams/${work.teamId}/objectives/${work.entityId}`;
  const content = (
    <>
      {story ? (
        <StoryIcon
          className="document-reference-icon"
          style={statusColor ? { color: statusColor } : undefined}
        />
      ) : (
        <ObjectiveIcon className="document-reference-icon" />
      )}
      {String(title)}
    </>
  );
  return (
    <NodeViewWrapper as="span" className="inline" contentEditable={false}>
      {path ? (
        <Link
          data-type="document-reference"
          href={withWorkspace(path)}
          style={style}
        >
          {content}
        </Link>
      ) : (
        <span data-type="document-reference" style={style}>
          {content}
        </span>
      )}
    </NodeViewWrapper>
  );
}

export const DocumentPersonReference = DocumentMention.extend({
  addNodeView: () => ReactNodeViewRenderer(PersonReference),
});
export const DocumentWorkReference = DocumentReference.extend<{
  documentId: string;
}>({
  addOptions: () => ({ documentId: "" }),
  addNodeView: () => ReactNodeViewRenderer(WorkReference),
});
