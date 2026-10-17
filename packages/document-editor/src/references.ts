import { Node } from "@tiptap/core";

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const validID = (value: string | null) => Boolean(value && UUID.test(value));

/** Persisted schema only; browser callers may extend it with suggestions/views. */
export const DocumentMention = Node.create({
  name: "mention",
  group: "inline",
  inline: true,
  atom: true,
  selectable: false,
  addAttributes() {
    return {
      id: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-id"),
      },
      label: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-label") ?? "",
      },
    };
  },
  parseHTML() {
    return [
      {
        tag: 'span[data-type="mention"][data-mention-kind="user"]',
        getAttrs: (element) =>
          validID(element.getAttribute("data-id")) ? {} : false,
      },
    ];
  },
  renderHTML({ node }) {
    return [
      "span",
      {
        "data-type": "mention",
        "data-mention-kind": "user",
        "data-id": node.attrs.id,
        "data-label": node.attrs.label,
        class: "mention",
      },
      `@${node.attrs.label}`,
    ];
  },
  renderText({ node }) {
    return `@${node.attrs.label}`;
  },
});

export const DocumentReference = Node.create({
  name: "documentReference",
  group: "inline",
  inline: true,
  atom: true,
  selectable: false,
  addAttributes() {
    return {
      id: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-id"),
      },
      entityType: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-entity-type"),
      },
      label: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-label") ?? "",
      },
      displayId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-display-id"),
      },
    };
  },
  parseHTML() {
    return [
      {
        tag: 'span[data-type="document-reference"]',
        getAttrs: (element) =>
          validID(element.getAttribute("data-id")) &&
          ["story", "objective"].includes(
            element.getAttribute("data-entity-type") ?? "",
          )
            ? {}
            : false,
      },
    ];
  },
  renderHTML({ node }) {
    return [
      "span",
      {
        "data-type": "document-reference",
        "data-entity-type": node.attrs.entityType,
        "data-id": node.attrs.id,
        "data-label": node.attrs.label,
        ...(node.attrs.displayId
          ? { "data-display-id": node.attrs.displayId }
          : {}),
        class: "document-reference",
      },
      node.attrs.label,
    ];
  },
  renderText({ node }) {
    return node.attrs.label;
  },
});
