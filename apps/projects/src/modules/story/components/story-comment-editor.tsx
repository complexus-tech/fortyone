import { mergeAttributes, ReactRenderer } from "@tiptap/react";
import Placeholder from "@tiptap/extension-placeholder";
import Link from "@tiptap/extension-link";
import Mention from "@tiptap/extension-mention";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import Underline from "@tiptap/extension-underline";
import StarterKit from "@tiptap/starter-kit";
import tippy from "tippy.js";
import type { MentionItem, MentionListRef } from "./mentions/list";
import { MentionList } from "./mentions/list";

type RichTextMark = {
  attrs?: Record<string, unknown>;
  type: string;
};

type RichTextNode = {
  attrs?: Record<string, unknown>;
  content?: RichTextNode[];
  marks?: RichTextMark[];
  text?: string;
  type?: string;
};

const EMPTY_MARKDOWN = "";
const MARKDOWN_BLOCK_SEPARATOR = "\n\n";

const renderMentionSuggestion = () => {
  let component: ReactRenderer<MentionListRef>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tippy.js instance type is complex
  let popup: any;

  return {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
    onStart: (props: any) => {
      component = new ReactRenderer(MentionList, {
        props,
        editor: props.editor,
      });

      if (!props.clientRect) {
        return;
      }

      popup = tippy("body", {
        getReferenceClientRect: props.clientRect,
        appendTo: () => document.body,
        content: component.element,
        showOnCreate: true,
        interactive: true,
        trigger: "manual",
        placement: "bottom-start",
      });
    },

    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
    onUpdate(props: any) {
      component.updateProps(props as Record<string, unknown>);
      if (!props.clientRect) {
        return;
      }

      popup?.[0]?.setProps({
        getReferenceClientRect: props.clientRect,
      });
    },

    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
    onKeyDown(props: any) {
      if (props.event.key === "Escape") {
        popup?.[0]?.hide();
        return true;
      }

      return component.ref?.onKeyDown(props.event as KeyboardEvent) ?? false;
    },

    onExit() {
      popup?.[0]?.destroy();
      component.destroy();
    },
  };
};

export const getStoryCommentEditorExtensions = ({
  enableMentions = false,
  mentionUsers = [],
  placeholder,
}: {
  enableMentions?: boolean;
  mentionUsers?: MentionItem[];
  placeholder: string;
}) => {
  const baseExtensions = [
    StarterKit,
    Underline,
    TaskList,
    TaskItem.configure({
      nested: true,
    }),
    Link.configure({
      autolink: true,
    }),
  ];
  const mentionExtensions = enableMentions
    ? [
        Mention.configure({
          HTMLAttributes: {
            class: "mention bg-surface-muted hover:bg-state-hover transition",
          },
          renderHTML({ node, options }) {
            return [
              "a",
              mergeAttributes(
                {
                  href: `/profile/${node.attrs.id}`,
                },
                options.HTMLAttributes,
              ),
              `${options.suggestion.char}${node.attrs.label ?? node.attrs.id}`,
            ];
          },
          suggestion: {
            items: ({ query }: { query: string }) => {
              if (!query || query.trim() === "") {
                return mentionUsers.slice(0, 6);
              }

              return mentionUsers
                .filter(
                  (user) =>
                    user.label.toLowerCase().includes(query.toLowerCase()) ||
                    user.username.toLowerCase().includes(query.toLowerCase()),
                )
                .slice(0, 6);
            },
            render: renderMentionSuggestion,
          },
        }),
      ]
    : [];

  return [
    ...baseExtensions,
    ...mentionExtensions,
    Placeholder.configure({
      placeholder,
    }),
  ];
};

const escapeMarkdownText = (value: string) =>
  value.replace(/\\/g, "\\\\").replace(/([`*_~[\]])/g, "\\$1");

const escapeMarkdownLinkText = (value: string) =>
  escapeMarkdownText(value).replace(/\(/g, "\\(").replace(/\)/g, "\\)");

const getNodeTextContent = (node?: RichTextNode): string => {
  if (!node) {
    return EMPTY_MARKDOWN;
  }

  if (node.type === "text") {
    return node.text ?? EMPTY_MARKDOWN;
  }

  return (node.content ?? []).map((child) => getNodeTextContent(child)).join("");
};

const serializeInlineNode = (node: RichTextNode): string => {
  if (node.type === "hardBreak") {
    return "\\\n";
  }

  if (node.type === "text") {
    const marks = node.marks ?? [];
    const hasCodeMark = marks.some((mark) => mark.type === "code");
    let output = hasCodeMark
      ? `\`${(node.text ?? EMPTY_MARKDOWN).replace(/`/g, "\\`")}\``
      : escapeMarkdownText(node.text ?? EMPTY_MARKDOWN);

    if (!hasCodeMark) {
      const isBold = marks.some((mark) => mark.type === "bold");
      const isItalic = marks.some((mark) => mark.type === "italic");
      const isStrike = marks.some((mark) => mark.type === "strike");
      const linkMark = marks.find((mark) => mark.type === "link");

      if (isBold) {
        output = `**${output}**`;
      }

      if (isItalic) {
        output = `*${output}*`;
      }

      if (isStrike) {
        output = `~~${output}~~`;
      }

      if (linkMark?.attrs?.href && typeof linkMark.attrs.href === "string") {
        output = `[${escapeMarkdownLinkText(output)}](${linkMark.attrs.href})`;
      }
    }

    return output;
  }

  if (node.type === "mention") {
    const label = node.attrs?.label;
    return typeof label === "string" ? `@${label}` : EMPTY_MARKDOWN;
  }

  if (node.type === "paragraph") {
    return (node.content ?? []).map((child) => serializeInlineNode(child)).join("");
  }

  if (node.type === "bulletList" || node.type === "orderedList") {
    return serializeBlockNode(node);
  }

  return (node.content ?? []).map((child) => serializeInlineNode(child)).join("");
};

const indentBlock = (value: string, depth: number) => {
  const indentation = "  ".repeat(depth);
  return value
    .split("\n")
    .map((line) => (line ? `${indentation}${line}` : line))
    .join("\n");
};

const serializeListNode = (
  node: RichTextNode,
  depth: number,
  type: "bullet" | "ordered" | "task",
): string =>
  (node.content ?? [])
    .map((child, index) => serializeListItemNode(child, depth, type, index))
    .filter(Boolean)
    .join("\n");

const serializeListItemNode = (
  node: RichTextNode,
  depth: number,
  type: "bullet" | "ordered" | "task",
  index: number,
): string => {
  const content = node.content ?? [];
  const blockNodes = content.filter(
    (child) =>
      child.type !== "bulletList" &&
      child.type !== "orderedList" &&
      child.type !== "taskList",
  );
  const nestedListNodes = content.filter(
    (child) =>
      child.type === "bulletList" ||
      child.type === "orderedList" ||
      child.type === "taskList",
  );
  const linePrefix =
    type === "ordered"
      ? `${index + 1}. `
      : type === "task"
        ? `- [${node.attrs?.checked ? "x" : " "}] `
        : "- ";
  const [firstBlockNode, ...remainingBlockNodes] = blockNodes;
  const nestedBlocks: string[] = [];
  let head = `${"  ".repeat(depth)}${linePrefix.trimEnd()}`;

  if (firstBlockNode?.type === "paragraph") {
    const contentText = (firstBlockNode.content ?? [])
      .map((child) => serializeInlineNode(child))
      .join("")
      .trim();

    head = contentText
      ? `${"  ".repeat(depth)}${linePrefix}${contentText}`
      : head;
  } else if (firstBlockNode?.type === "heading") {
    const level =
      typeof firstBlockNode.attrs?.level === "number"
        ? Math.min(firstBlockNode.attrs.level, 6)
        : 1;
    const contentText = (firstBlockNode.content ?? [])
      .map((child) => serializeInlineNode(child))
      .join("")
      .trim();

    head = contentText
      ? `${"  ".repeat(depth)}${linePrefix}${"#".repeat(level)} ${contentText}`
      : head;
  } else if (firstBlockNode) {
    nestedBlocks.push(serializeBlockNode(firstBlockNode, depth + 1));
  }

  nestedBlocks.push(
    ...remainingBlockNodes.map((child) => serializeBlockNode(child, depth + 1)),
  );

  const nested = [
    ...nestedBlocks,
    ...nestedListNodes.map((child) => serializeBlockNode(child, depth + 1)),
  ]
    .filter(Boolean)
    .join("\n");

  if (!nested) {
    return head;
  }

  return `${head}\n${nested}`;
};

const serializeBlockquoteNode = (node: RichTextNode, depth: number): string => {
  const content = (node.content ?? [])
    .map((child) => serializeBlockNode(child, depth))
    .filter(Boolean)
    .join(MARKDOWN_BLOCK_SEPARATOR);

  if (!content) {
    return EMPTY_MARKDOWN;
  }

  return content
    .split("\n")
    .map((line) => (line ? `${"  ".repeat(depth)}> ${line}` : `${"  ".repeat(depth)}>`))
    .join("\n");
};

const serializeCodeBlockNode = (node: RichTextNode, depth: number): string => {
  const code = getNodeTextContent(node);
  const language =
    typeof node.attrs?.language === "string" ? node.attrs.language : EMPTY_MARKDOWN;
  const fence = `${"  ".repeat(depth)}\`\`\`${language}`;
  const body = code
    .split("\n")
    .map((line) => `${"  ".repeat(depth)}${line}`)
    .join("\n");

  return `${fence}\n${body}\n${"  ".repeat(depth)}\`\`\``;
};

const serializeBlockNode = (node: RichTextNode, depth = 0): string => {
  switch (node.type) {
    case "bulletList":
      return serializeListNode(node, depth, "bullet");
    case "orderedList":
      return serializeListNode(node, depth, "ordered");
    case "taskList":
      return serializeListNode(node, depth, "task");
    case "heading": {
      const level =
        typeof node.attrs?.level === "number" ? Math.min(node.attrs.level, 6) : 1;
      const content = (node.content ?? [])
        .map((child) => serializeInlineNode(child))
        .join("")
        .trim();

      return content ? `${"  ".repeat(depth)}${"#".repeat(level)} ${content}` : EMPTY_MARKDOWN;
    }
    case "blockquote":
      return serializeBlockquoteNode(node, depth);
    case "codeBlock":
      return serializeCodeBlockNode(node, depth);
    case "horizontalRule":
      return `${"  ".repeat(depth)}---`;
    case "paragraph": {
      const content = (node.content ?? [])
        .map((child) => serializeInlineNode(child))
        .join("")
        .trim();

      return content ? `${"  ".repeat(depth)}${content}` : EMPTY_MARKDOWN;
    }
    default: {
      const content = (node.content ?? [])
        .map((child) => serializeBlockNode(child, depth))
        .filter(Boolean)
        .join(MARKDOWN_BLOCK_SEPARATOR);

      return content ? indentBlock(content, depth) : EMPTY_MARKDOWN;
    }
  }
};

export const serializeStoryCommentToGitHubMarkdown = (node: RichTextNode): string =>
  (node.content ?? [])
    .map((child) => serializeBlockNode(child))
    .filter(Boolean)
    .join(MARKDOWN_BLOCK_SEPARATOR)
    .replace(/\n{3,}/g, MARKDOWN_BLOCK_SEPARATOR)
    .trim();
