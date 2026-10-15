import type { HocuspocusProvider } from "@hocuspocus/provider";
import type * as Y from "yjs";

export type Connection = {
  document: Y.Doc;
  provider: HocuspocusProvider | null;
};
export type Peer = { id: string; name: string; color: string };
export type Status = "connecting" | "saved" | "saving" | "offline" | "blocked";
const COLORS = ["#2563eb", "#9333ea", "#db2777", "#0891b2", "#16a34a"];

export function collaborationColor(id: string) {
  return COLORS[
    [...id].reduce((hash, char) => hash + char.charCodeAt(0), 0) % COLORS.length
  ];
}

export function replaceCollaborativeTitle(text: Y.Text, value: string) {
  const previous = text.toJSON();
  let start = 0;
  while (
    start < previous.length &&
    start < value.length &&
    previous[start] === value[start]
  )
    start++;
  let end = 0;
  while (
    end < previous.length - start &&
    end < value.length - start &&
    previous[previous.length - end - 1] === value[value.length - end - 1]
  )
    end++;
  text.doc?.transact(() => {
    if (previous.length - start - end > 0)
      text.delete(start, previous.length - start - end);
    if (value.length - start - end > 0)
      text.insert(start, value.slice(start, value.length - end));
  });
}
