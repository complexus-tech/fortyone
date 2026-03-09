import type { Slice } from "@tiptap/pm/model";
import type { EditorView } from "@tiptap/pm/view";
import { parseFigmaUrl } from "@/lib/utils/figma";

const getFigmaLinkTitle = (url: string) => {
  const parsed = parseFigmaUrl(url);
  if (!parsed) {
    return null;
  }

  if (parsed.name) {
    return parsed.name;
  }

  return parsed.resourceType === "node" ? "Figma frame" : "Figma file";
};

const isSingleURL = (value: string) => {
  if (!value) {
    return false;
  }

  return !/\s/.test(value);
};

const fetchFigmaTitle = async (url: string): Promise<string | null> => {
  try {
    const response = await fetch(
      `/api/metadata?url=${encodeURIComponent(url)}`,
    );
    if (!response.ok) {
      return null;
    }

    const payload = (await response.json()) as { title?: string };
    const title = payload.title?.trim();

    return title || null;
  } catch {
    return null;
  }
};

const replaceFirstMatchingLinkText = ({
  view,
  href,
  currentTitle,
  nextTitle,
}: {
  view: EditorView;
  href: string;
  currentTitle: string;
  nextTitle: string;
}) => {
  const linkMark = view.state.schema.marks.link;
  if (!linkMark) {
    return;
  }

  let from = -1;
  let to = -1;
  view.state.doc.descendants((node, pos) => {
    if (!node.isText || node.text !== currentTitle) {
      return true;
    }

    const hasMatchingHref = node.marks.some(
      (mark) => mark.type === linkMark && mark.attrs.href === href,
    );

    if (!hasMatchingHref) {
      return true;
    }

    from = pos;
    to = pos + node.nodeSize;

    return false;
  });

  if (from === -1 || to === -1) {
    return;
  }

  const transaction = view.state.tr.replaceWith(
    from,
    to,
    view.state.schema.text(nextTitle, [linkMark.create({ href })]),
  );
  view.dispatch(transaction);
};

export const handleFigmaLinkPaste = (
  view: EditorView,
  event: ClipboardEvent,
  _slice: Slice,
) => {
  const text = event.clipboardData?.getData("text/plain")?.trim();
  if (!text || !isSingleURL(text)) {
    return false;
  }

  const parsed = parseFigmaUrl(text);
  if (!parsed) {
    return false;
  }

  const linkMark = view.state.schema.marks.link;
  if (!linkMark) {
    return false;
  }

  const title = getFigmaLinkTitle(text);
  if (!title) {
    return false;
  }

  event.preventDefault();

  const textNode = view.state.schema.text(title, [
    linkMark.create({
      href: parsed.canonicalUrl,
    }),
  ]);

  const transaction = view.state.tr.replaceSelectionWith(textNode, false);
  view.dispatch(transaction.scrollIntoView());

  void fetchFigmaTitle(parsed.canonicalUrl).then((resolvedTitle) => {
    if (!resolvedTitle || resolvedTitle === title) {
      return;
    }

    replaceFirstMatchingLinkText({
      view,
      href: parsed.canonicalUrl,
      currentTitle: title,
      nextTitle: resolvedTitle,
    });
  });

  return true;
};
