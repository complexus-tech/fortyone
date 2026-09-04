import type { Editor } from "@tiptap/core";
import type { GoogleDriveFileReference } from "@/shared/google-drive/types";

const GOOGLE_DRIVE_LINK_CLASS = "google-drive-smart-link";

const withGoogleDriveLinkClass = (value: unknown) => {
  const classes = typeof value === "string" ? value.split(/\s+/) : [];
  return [
    ...new Set([...classes.filter(Boolean), GOOGLE_DRIVE_LINK_CLASS]),
  ].join(" ");
};

export const replacePastedGoogleDriveURLLabel = ({
  approximatePosition,
  editor,
  file,
  rawURL,
}: {
  approximatePosition: number;
  editor: Editor;
  file: GoogleDriveFileReference;
  rawURL: string;
}) => {
  if (editor.isDestroyed) return false;

  const linkType = editor.state.schema.marks.link;
  const candidates: {
    distance: number;
    from: number;
    marks: Parameters<typeof editor.state.schema.text>[1];
    to: number;
  }[] = [];

  editor.state.doc.descendants((node, position) => {
    if (!node.isText || node.text !== rawURL) return;

    const existingLink = node.marks.find(
      (mark) =>
        mark.type === linkType &&
        (mark.attrs.href === rawURL || mark.attrs.href === file.webViewLink),
    );
    const from = position;
    const to = position + node.nodeSize;
    const distance = Math.min(
      Math.abs(approximatePosition - from),
      Math.abs(approximatePosition - to),
    );
    const linkMark = linkType.create({
      ...existingLink?.attrs,
      class: withGoogleDriveLinkClass(existingLink?.attrs.class),
      href: file.webViewLink,
      title: file.name,
    });

    candidates.push({
      distance,
      from,
      marks: [...node.marks.filter((mark) => mark.type !== linkType), linkMark],
      to,
    });
  });

  if (candidates.length === 0) return false;
  const candidate = candidates.toSorted(
    (left, right) => left.distance - right.distance,
  )[0];
  const replacement = editor.state.schema.text(file.name, candidate.marks);
  editor.view.dispatch(
    editor.state.tr.replaceWith(candidate.from, candidate.to, replacement),
  );
  return true;
};
