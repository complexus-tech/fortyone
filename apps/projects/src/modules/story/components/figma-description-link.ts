import type { Editor } from "@tiptap/core";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";
import { isFigmaURL } from "@/modules/settings/workspace/integrations/figma/url";

const FIGMA_LINK_CLASS = "figma-smart-link";

export const getStandaloneFigmaURL = (value: string) => {
  const trimmedValue = value.trim();
  if (!isFigmaURL(trimmedValue)) return null;

  const protocol = new URL(trimmedValue).protocol;
  return protocol === "https:" || protocol === "http:" ? trimmedValue : null;
};

export const getFigmaLinkLabel = (artifact: FigmaArtifact) =>
  artifact.nodeName?.trim() || artifact.fileName.trim() || "Figma design";

const withFigmaLinkClass = (value: unknown) => {
  const classes = typeof value === "string" ? value.split(/\s+/) : [];
  return [...new Set([...classes.filter(Boolean), FIGMA_LINK_CLASS])].join(" ");
};

export const replacePastedFigmaURLLabel = ({
  approximatePosition,
  artifact,
  editor,
  rawURL,
}: {
  approximatePosition: number;
  artifact: FigmaArtifact;
  editor: Editor;
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
        (mark.attrs.href === rawURL ||
          mark.attrs.href === artifact.canonicalUrl),
    );
    const from = position;
    const to = position + node.nodeSize;
    const distance = Math.min(
      Math.abs(approximatePosition - from),
      Math.abs(approximatePosition - to),
    );
    const linkMark = linkType.create({
      ...existingLink?.attrs,
      class: withFigmaLinkClass(existingLink?.attrs.class),
      href: artifact.canonicalUrl,
      title: getFigmaLinkLabel(artifact),
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

  const replacement = editor.state.schema.text(
    getFigmaLinkLabel(artifact),
    candidate.marks,
  );
  editor.view.dispatch(
    editor.state.tr.replaceWith(candidate.from, candidate.to, replacement),
  );
  return true;
};
