import type { JSONContent } from "@tiptap/core";

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
export type DocumentReferenceTarget = {
  entityType: "story" | "objective";
  id: string;
};

export function extractDocumentLinks(document: JSONContent) {
  const references = new Map<string, DocumentReferenceTarget>();
  const mentions = new Set<string>();
  const visit = (node: JSONContent) => {
    if (node.type === "mention" || node.type === "documentReference") {
      const id: unknown = node.attrs?.id;
      if (typeof id !== "string" || !UUID.test(id))
        throw new Error("Invalid document reference");
      if (node.type === "mention") mentions.add(id);
      else {
        const entityType: unknown = node.attrs?.entityType;
        if (entityType !== "story" && entityType !== "objective")
          throw new Error("Invalid document reference");
        references.set(`${entityType}:${id}`, { entityType, id });
      }
    }
    node.content?.forEach(visit);
  };
  visit(document);
  return { references: [...references.values()], mentions: [...mentions] };
}
