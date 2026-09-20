import { mergeAttributes, Node, type Extensions } from "@tiptap/core";
import Color from "@tiptap/extension-color";
import Image from "@tiptap/extension-image";
import Highlight from "@tiptap/extension-highlight";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { TextStyle } from "@tiptap/extension-text-style";
import Underline from "@tiptap/extension-underline";
import {
  DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT,
  DEFAULT_RICH_TEXT_IMAGE_WIDTH,
  normalizeRichTextImageWidth,
  normalizeRichTextImageAlignment,
  getRichTextImageStyleString,
} from "./image-utils";

export const DocumentImage = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      width: {
        default: DEFAULT_RICH_TEXT_IMAGE_WIDTH,
        parseHTML: (element) =>
          normalizeRichTextImageWidth(
            element.getAttribute("data-width") || element.style.width,
          ),
        renderHTML: (attributes) => ({
          "data-width": normalizeRichTextImageWidth(attributes.width),
          style: getRichTextImageStyleString(
            attributes.width,
            attributes.align,
          ),
        }),
      },
      align: {
        default: DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT,
        parseHTML: (element) =>
          normalizeRichTextImageAlignment(element.getAttribute("data-align")),
        renderHTML: (attributes) => ({
          "data-align": normalizeRichTextImageAlignment(attributes.align),
        }),
      },
      attachmentId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-attachment-id"),
        renderHTML: (attributes) =>
          typeof attributes.attachmentId === "string" &&
          attributes.attachmentId.trim()
            ? { "data-attachment-id": attributes.attachmentId }
            : {},
      },
      uploadId: {
        default: null,
        parseHTML: () => null,
        renderHTML: () => ({}),
      },
      isUploading: {
        default: false,
        parseHTML: () => false,
        renderHTML: () => ({}),
      },
    };
  },
});

export const DocumentVideo = Node.create({
  name: "documentVideo",
  group: "block",
  atom: true,
  draggable: true,
  selectable: true,
  addAttributes() {
    return {
      src: { default: null },
      attachmentId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-attachment-id"),
        renderHTML: (attributes) =>
          typeof attributes.attachmentId === "string" &&
          attributes.attachmentId.trim()
            ? { "data-attachment-id": attributes.attachmentId }
            : {},
      },
      uploadId: {
        default: null,
        parseHTML: () => null,
        renderHTML: () => ({}),
      },
      isUploading: {
        default: false,
        parseHTML: () => false,
        renderHTML: () => ({}),
      },
    };
  },
  parseHTML() {
    return [{ tag: "video[data-document-media-video]" }];
  },
  renderHTML({ HTMLAttributes }) {
    return [
      "video",
      mergeAttributes(HTMLAttributes, {
        class:
          "my-5 aspect-video w-full max-w-full rounded-xl border border-border bg-black object-contain",
        controls: "true",
        "data-document-media-video": "true",
        preload: "metadata",
      }),
    ];
  },
});

// This schema is shared by browser editing and server persistence. UI extensions
// may add node views, but must not change the persisted content structure.
export const createDocumentExtensions = ({
  collaborative = false,
  image = DocumentImage,
  video = DocumentVideo,
} = {}): Extensions => [
  StarterKit.configure({
    link: false,
    underline: false,
    undoRedo: collaborative ? false : undefined,
  }),
  Underline,
  TextStyle,
  Color,
  Highlight.configure({ multicolor: true }),
  TaskList,
  TaskItem.configure({ nested: true }),
  Link.configure({ autolink: true }),
  image.configure({
    allowBase64: false,
    HTMLAttributes: { class: "max-w-full rounded-xl border border-border" },
  }),
  video,
  Table.configure({ resizable: true }),
  TableRow,
  TableHeader,
  TableCell,
];
