import { Extension, generateHTML, generateJSON } from "@tiptap/core";
import Link from "@tiptap/extension-link";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { Markdown, MarkdownManager } from "@tiptap/markdown";
import { Plugin } from "@tiptap/pm/state";
import { marked } from "marked";
import { createRichTextStarterKit } from "./starter-kit";

const MARKDOWN_BLOCK_PATTERN =
  /(?:^|\n)\s{0,3}(?:#{1,6}\s+\S|[-+*]\s+(?:\[[ xX]\]\s+)?\S|\d+[.)]\s+\S|>\s+\S|```|~~~)/u;
const MARKDOWN_INLINE_PATTERN =
  /(?:\*\*[^*\n]+\*\*|__[^_\n]+__|~~[^~\n]+~~|\[[^\]\n]+\]\((?:https?:\/\/|mailto:)[^)\s]+\))/u;
const MARKDOWN_TABLE_PATTERN =
  /(?:^|\n)\s*\|?[^\n|]+\|[^\n]+\n\s*\|?\s*:?-{3,}:?\s*\|/u;

export const looksLikeMarkdown = (value: string) =>
  MARKDOWN_BLOCK_PATTERN.test(value) ||
  MARKDOWN_INLINE_PATTERN.test(value) ||
  MARKDOWN_TABLE_PATTERN.test(value);

export const getRichTextContentType = (
  description: string,
  descriptionHTML?: string | null,
) =>
  !descriptionHTML && looksLikeMarkdown(description)
    ? ("markdown" as const)
    : ("html" as const);

const createMarkdownDocumentExtensions = () => [
  createRichTextStarterKit(),
  TaskList.configure({}),
  TaskItem.configure({ nested: true }),
  Link.configure({ autolink: true }),
  Table.configure({}),
  TableRow.configure({}),
  TableHeader.configure({}),
  TableCell.configure({}),
  Markdown.configure({ markedOptions: { gfm: true } }),
];

export const RichTextMarkdown = Markdown.configure({
  markedOptions: { gfm: true },
});

export const markdownDocumentExtensions = createMarkdownDocumentExtensions();
export const markdownManager = new MarkdownManager({
  extensions: markdownDocumentExtensions,
  markedOptions: { gfm: true },
});

export const markdownToRichTextHTML = (markdown: string) =>
  generateHTML(markdownManager.parse(markdown), markdownDocumentExtensions);

export const richTextHTMLToMarkdown = (html: string) =>
  markdownManager.serialize(generateJSON(html, markdownDocumentExtensions));

const markdownPasteHTML = (markdown: string) => {
  const template = document.createElement("template");
  template.innerHTML = marked.parse(markdown, { async: false, gfm: true });

  template.content.querySelectorAll("li").forEach((item) => {
    const checkbox = item.querySelector(":scope > input[type='checkbox']");
    if (!checkbox) return;

    item.dataset.type = "taskItem";
    item.dataset.checked = String(
      checkbox.hasAttribute("checked") ||
        (checkbox as HTMLInputElement).checked,
    );
    item.parentElement?.setAttribute("data-type", "taskList");
    checkbox.remove();
  });

  return template.innerHTML;
};

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

            // Parse the clipboard independently of the live editor's Markdown
            // manager. Its extension state is shared with the collaborative
            // editor and can lose nested table content after earlier blocks.
            const inserted = editor.commands.insertContent(
              markdownPasteHTML(markdown),
            );
            if (!inserted) return false;

            event.preventDefault();
            return true;
          },
        },
      }),
    ];
  },
});
