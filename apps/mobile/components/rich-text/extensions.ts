import { mergeAttributes, Node, type Editor } from "@tiptap/core";
import { StarterKit } from "@tiptap/starter-kit";
import { Image } from "@tiptap/extension-image";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import {
  Table,
  TableCell,
  TableHeader,
  TableRow,
} from "@tiptap/extension-table";
import { Placeholder } from "@tiptap/extension-placeholder";
import { isSafeLink, type RichTextValue } from "./content";
import { sanitizeRichText } from "./sanitize";

const attachmentAttribute = {
  default: null,
  parseHTML: (element: HTMLElement) =>
    element.getAttribute("data-attachment-id"),
  renderHTML: (attributes: Record<string, unknown>) =>
    attributes.attachmentId
      ? { "data-attachment-id": attributes.attachmentId }
      : {},
};

const normalizeWidth = (value: unknown) =>
  typeof value === "string" && /^\d+(?:\.\d+)?(?:px|%)$/.test(value)
    ? value
    : "100%";
const normalizeAlignment = (value: unknown) =>
  value === "left" || value === "right" ? value : "center";

const RichTextImage = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      attachmentId: attachmentAttribute,
      width: {
        default: "100%",
        parseHTML: (element) =>
          normalizeWidth(
            element.getAttribute("data-width") || element.style.width,
          ),
        renderHTML: (attributes) => ({
          "data-width": normalizeWidth(attributes.width),
          style: `display:block;max-width:100%;width:${normalizeWidth(attributes.width)};margin-left:${attributes.align === "left" ? "0" : "auto"};margin-right:${attributes.align === "right" ? "0" : "auto"}`,
        }),
      },
      align: {
        default: "center",
        parseHTML: (element) =>
          normalizeAlignment(element.getAttribute("data-align")),
        renderHTML: (attributes) => ({
          "data-align": normalizeAlignment(attributes.align),
        }),
      },
    };
  },
}).configure({ allowBase64: false });

const RichTextVideo = Node.create({
  name: "documentVideo",
  group: "block",
  atom: true,
  draggable: true,
  addAttributes() {
    return { src: { default: null }, attachmentId: attachmentAttribute };
  },
  parseHTML() {
    return [{ tag: "video[data-document-media-video]" }];
  },
  renderHTML({ HTMLAttributes }) {
    return [
      "video",
      mergeAttributes(HTMLAttributes, {
        "data-document-media-video": "true",
        controls: "",
        preload: "none",
      }),
    ];
  },
});

// The web comment editor emits profile anchors. Also accept standard TipTap mentions.
const Mention = Node.create({
  name: "mention",
  group: "inline",
  inline: true,
  atom: true,
  priority: 1000,
  addAttributes() {
    return {
      id: {
        default: null,
        parseHTML: (element) =>
          element.getAttribute("data-id") ??
          element.getAttribute("href")?.split("/profile/")[1],
      },
      label: {
        default: "",
        parseHTML: (element) =>
          element.getAttribute("data-label") ??
          element.textContent?.replace(/^@/, "") ??
          "",
      },
    };
  },
  parseHTML() {
    return [
      { tag: 'span[data-type="mention"]' },
      { tag: 'a[href^="/profile/"]' },
    ];
  },
  renderHTML({ node }) {
    return [
      "a",
      {
        "data-type": "mention",
        "data-id": node.attrs.id,
        "data-label": node.attrs.label,
        href: `/profile/${node.attrs.id}`,
        class: "mention",
      },
      `@${node.attrs.label}`,
    ];
  },
  renderText({ node }) {
    return `@${node.attrs.label}`;
  },
});

export const createMobileRichTextExtensions = (
  placeholder = "Write something…",
) => [
  StarterKit.configure({
    link: {
      openOnClick: false,
      autolink: true,
      defaultProtocol: "https",
      isAllowedUri: (url) => isSafeLink(url),
    },
  }),
  TaskList,
  TaskItem.configure({ nested: true }),
  Table.configure({ resizable: false }),
  TableRow,
  TableHeader,
  TableCell,
  RichTextImage,
  RichTextVideo,
  Mention,
  Placeholder.configure({ placeholder }),
];

export const getRichTextValue = (editor: Editor): RichTextValue => {
  const mentions = new Set<string>();
  editor.state.doc.descendants((node) => {
    if (node.type.name === "mention" && typeof node.attrs.id === "string")
      mentions.add(node.attrs.id);
  });
  return {
    html: editor.isEmpty ? "" : sanitizeRichText(editor.getHTML()),
    text: editor.isEmpty ? "" : editor.getText(),
    mentions: [...mentions],
  };
};
