import type { JSONContent } from "@tiptap/core";

type RichTextMark = {
  attrs?: Record<string, unknown>;
  type: string;
};

type RichTextNode = JSONContent & {
  attrs?: Record<string, unknown>;
  content?: RichTextNode[];
  marks?: RichTextMark[];
  text?: string;
  type?: string;
};

const EMPTY_MARKDOWN = "";
const MARKDOWN_BLOCK_SEPARATOR = "\n\n";

const escapeMarkdownText = (value: string) =>
  value.replace(/\\/g, "\\\\").replace(/[`*_~[\]]/g, "\\$&");

const escapeMarkdownLinkText = (value: string) =>
  escapeMarkdownText(value).replace(/\(/g, "\\(").replace(/\)/g, "\\)");

const getNodeTextContent = (node?: RichTextNode): string => {
  if (!node) return EMPTY_MARKDOWN;
  if (node.type === "text") return node.text ?? EMPTY_MARKDOWN;

  return (node.content ?? [])
    .map((child) => getNodeTextContent(child))
    .join("");
};

const serializeInlineNode = (node: RichTextNode): string => {
  if (node.type === "hardBreak") return "\\\n";

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

      if (isBold) output = `**${output}**`;
      if (isItalic) output = `*${output}*`;
      if (isStrike) output = `~~${output}~~`;
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
    return (node.content ?? [])
      .map((child) => serializeInlineNode(child))
      .join("");
  }

  if (node.type === "bulletList" || node.type === "orderedList") {
    return serializeBlockNode(node);
  }

  return (node.content ?? [])
    .map((child) => serializeInlineNode(child))
    .join("");
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
  let linePrefix = "- ";
  if (type === "ordered") {
    linePrefix = `${index + 1}. `;
  } else if (type === "task") {
    linePrefix = `- [${node.attrs?.checked ? "x" : " "}] `;
  }

  const firstBlockNode = blockNodes.at(0);
  const remainingBlockNodes = blockNodes.slice(1);
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

  return nested ? `${head}\n${nested}` : head;
};

const serializeBlockquoteNode = (node: RichTextNode, depth: number): string => {
  const content = (node.content ?? [])
    .map((child) => serializeBlockNode(child, depth))
    .filter(Boolean)
    .join(MARKDOWN_BLOCK_SEPARATOR);
  if (!content) return EMPTY_MARKDOWN;

  return content
    .split("\n")
    .map((line) =>
      line ? `${"  ".repeat(depth)}> ${line}` : `${"  ".repeat(depth)}>`,
    )
    .join("\n");
};

const serializeCodeBlockNode = (node: RichTextNode, depth: number): string => {
  const code = getNodeTextContent(node);
  const language =
    typeof node.attrs?.language === "string"
      ? node.attrs.language
      : EMPTY_MARKDOWN;
  const indentation = "  ".repeat(depth);
  const fence = `${indentation}\`\`\`${language}`;
  const body = code
    .split("\n")
    .map((line) => `${indentation}${line}`)
    .join("\n");

  return `${fence}\n${body}\n${indentation}\`\`\``;
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
        typeof node.attrs?.level === "number"
          ? Math.min(node.attrs.level, 6)
          : 1;
      const content = (node.content ?? [])
        .map((child) => serializeInlineNode(child))
        .join("")
        .trim();

      return content
        ? `${"  ".repeat(depth)}${"#".repeat(level)} ${content}`
        : EMPTY_MARKDOWN;
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

/** Serializes the shared Tiptap comment document into GitHub-flavored Markdown. */
export const serializeCommentToGitHubMarkdown = (node: RichTextNode): string =>
  (node.content ?? [])
    .map((child) => serializeBlockNode(child))
    .filter(Boolean)
    .join(MARKDOWN_BLOCK_SEPARATOR)
    .replace(/\n{3,}/g, MARKDOWN_BLOCK_SEPARATOR)
    .trim();
