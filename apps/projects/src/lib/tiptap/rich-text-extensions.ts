import type { Editor } from "@tiptap/core";
import Link from "@tiptap/extension-link";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import Placeholder from "@tiptap/extension-placeholder";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import Underline from "@tiptap/extension-underline";
import { createRichTextStarterKit } from "./starter-kit";
import { SlashCommand } from "./slash-command";
import { RichTextMarkdown, RichTextMarkdownPaste } from "./markdown";
import {
  RichTextImage,
  RichTextMediaDrop,
  RichTextVideo,
} from "./rich-text-media";

type CreateRichTextExtensionsOptions = {
  onMediaFiles: (editor: Editor, files: File[], position?: number) => void;
  onMediaRequest: (editor: Editor) => void;
  placeholder: string;
};

export const createRichTextExtensions = ({
  onMediaFiles,
  onMediaRequest,
  placeholder,
}: CreateRichTextExtensionsOptions) => [
  createRichTextStarterKit(),
  Underline,
  TaskList,
  TaskItem.configure({ nested: true }),
  RichTextMarkdown,
  RichTextMarkdownPaste,
  Link.configure({ autolink: true }),
  RichTextImage.configure({
    allowBase64: false,
    HTMLAttributes: {
      class: "max-w-full rounded-xl border border-border",
    },
  }),
  RichTextVideo,
  RichTextMediaDrop.configure({ onFiles: onMediaFiles }),
  Table.configure({ resizable: true }),
  TableRow,
  TableHeader,
  TableCell,
  Placeholder.configure({ placeholder }),
  SlashCommand.configure({ onMediaRequest }),
];
