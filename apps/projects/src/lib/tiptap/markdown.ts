import { Extension, generateHTML } from "@tiptap/core";
import Link from "@tiptap/extension-link";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { Markdown, MarkdownManager } from "@tiptap/markdown";
import { Plugin } from "@tiptap/pm/state";
import { createRichTextStarterKit } from "./starter-kit";

const MARKDOWN_BLOCK_PATTERN =
  /(?:^|\n)\s{0,3}(?:#{1,6}\s+\S|[-+*]\s+(?:\[[ xX]\]\s+)?\S|\d+[.)]\s+\S|>\s+\S|```|~~~)/u;
const MARKDOWN_INLINE_PATTERN =
  /(?:\*\*[^*\n]+\*\*|__[^_\n]+__|~~[^~\n]+~~|\[[^\]\n]+\]\((?:https?:\/\/|mailto:)[^)\s]+\))/u;

export const looksLikeMarkdown = (value: string) =>
  MARKDOWN_BLOCK_PATTERN.test(value) || MARKDOWN_INLINE_PATTERN.test(value);

export const getRichTextContentType = (
  description: string,
  descriptionHTML?: string | null,
) =>
  !descriptionHTML && looksLikeMarkdown(description)
    ? ("markdown" as const)
    : ("html" as const);

export const RichTextMarkdown = Markdown.configure({
  markedOptions: { gfm: true },
});

const markdownDocumentExtensions = [
  createRichTextStarterKit(),
  TaskList,
  TaskItem.configure({ nested: true }),
  Link.configure({ autolink: true }),
  Table,
  TableRow,
  TableHeader,
  TableCell,
  RichTextMarkdown,
];
const markdownManager = new MarkdownManager({
  extensions: markdownDocumentExtensions,
  markedOptions: { gfm: true },
});

export const markdownToRichTextHTML = (markdown: string) =>
  generateHTML(markdownManager.parse(markdown), markdownDocumentExtensions);

export const RichTextMarkdownPaste = Extension.create({
  name: "richTextMarkdownPaste",

  addProseMirrorPlugins() {
    const editor = this.editor;

    return [
      new Plugin({
        props: {
          handlePaste(_view, event) {
            const markdown = event.clipboardData?.getData("text/plain");
            if (!markdown || !looksLikeMarkdown(markdown)) return false;

            const inserted = editor.commands.insertContent(markdown, {
              contentType: "markdown",
            });
            if (!inserted) return false;

            event.preventDefault();
            return true;
          },
        },
      }),
    ];
  },
});
