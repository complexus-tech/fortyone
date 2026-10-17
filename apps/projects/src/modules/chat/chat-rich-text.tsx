import type { JSONContent } from "@tiptap/core";
import type { ReactNode } from "react";
import { Fragment } from "react";
import { findChatLinks, getChatLinkHref } from "./chat-links";
import type { ChatAttachment } from "./types";
import { MessageLink } from "./message-link";

export { getChatLinkHref } from "./chat-links";

export type PreviewDraft = {
  text: string;
  content: JSONContent | null;
  attachments?: ChatAttachment[];
  baseVersion?: number;
};

export type ChatDraft = PreviewDraft;

export const EMPTY_PREVIEW_DRAFT: PreviewDraft = { text: "", content: null };

export const CHAT_CONTENT_CLASS =
  "text-foreground text-base leading-6 [overflow-wrap:anywhere] [&_p]:my-1 [&_p]:min-h-6 [&_strong]:font-semibold [&_a]:text-primary [&_a]:underline [&_a]:underline-offset-2 [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-6 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-6 [&_li>p]:my-0 [&_blockquote]:border-border-strong [&_blockquote]:text-text-muted [&_blockquote]:my-2 [&_blockquote]:border-l-2 [&_blockquote]:pl-3 [&_code]:bg-surface-muted [&_code]:rounded [&_code]:px-1 [&_code]:py-0.5 [&_code]:font-mono [&_pre]:bg-surface-muted [&_pre]:my-2 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:p-3 [&_pre]:whitespace-pre-wrap [&_pre_code]:bg-transparent [&_pre_code]:p-0";

export const ChatMention = ({ children }: { children: ReactNode }) => (
  <span className="bg-primary/10 text-primary rounded px-1 py-0.5 font-medium">
    {children}
  </span>
);

const renderMentions = (text: string): ReactNode =>
  text.split("@Maya").map((part, index) => (
    <Fragment key={index}>
      {index > 0 ? <ChatMention>@Maya</ChatMention> : null}
      {part}
    </Fragment>
  ));

const renderLinkedText = (text: string): ReactNode => {
  const children: ReactNode[] = [];
  let offset = 0;
  for (const link of findChatLinks(text)) {
    children.push(
      <Fragment key={`text:${offset}`}>
        {renderMentions(text.slice(offset, link.start))}
      </Fragment>,
    );
    children.push(
      <MessageLink href={link.href} key={`link:${link.start}`}>
        {link.text}
      </MessageLink>,
    );
    offset = link.end;
  }
  children.push(
    <Fragment key={`text:${offset}`}>
      {renderMentions(text.slice(offset))}
    </Fragment>,
  );
  return children;
};

const renderText = (node: JSONContent, inCode: boolean): ReactNode => {
  const code =
    inCode || Boolean(node.marks?.some((mark) => mark.type === "code"));
  const explicitlyLinked = Boolean(
    node.marks?.some((mark) => mark.type === "link"),
  );
  let result: ReactNode = node.text ?? "";
  if (!code)
    result = explicitlyLinked
      ? renderMentions(node.text ?? "")
      : renderLinkedText(node.text ?? "");
  for (const mark of node.marks ?? []) {
    switch (mark.type) {
      case "bold":
        result = <strong>{result}</strong>;
        break;
      case "italic":
        result = <em>{result}</em>;
        break;
      case "underline":
        result = <u>{result}</u>;
        break;
      case "strike":
        result = <s>{result}</s>;
        break;
      case "code":
        result = <code>{result}</code>;
        break;
      case "link": {
        const href = code ? null : getChatLinkHref(mark.attrs?.href);
        if (href) {
          result = (
            <MessageLink
              href={href}
              shorten={getChatLinkHref(node.text) === href}
            >
              {result}
            </MessageLink>
          );
        }
        break;
      }
    }
  }

  return result;
};

const renderNode = (node: JSONContent, inCode = false): ReactNode => {
  if (node.type === "text") return renderText(node, inCode);
  const children = node.content?.map((child, index) => (
    <Fragment key={index}>
      {renderNode(child, inCode || node.type === "codeBlock")}
    </Fragment>
  ));

  switch (node.type) {
    case "mention":
      return (
        <ChatMention>
          @{typeof node.attrs?.label === "string" ? node.attrs.label : "Member"}
        </ChatMention>
      );
    case "doc":
      return children;
    case "paragraph":
      return <p>{children ?? <br />}</p>;
    case "hardBreak":
      return <br />;
    case "bulletList":
      return <ul>{children}</ul>;
    case "orderedList":
      return (
        <ol
          start={
            typeof node.attrs?.start === "number" ? node.attrs.start : undefined
          }
        >
          {children}
        </ol>
      );
    case "listItem":
      return <li>{children}</li>;
    case "blockquote":
      return <blockquote>{children}</blockquote>;
    case "codeBlock":
      return (
        <pre>
          <code>{children}</code>
        </pre>
      );
    default:
      return children;
  }
};

export const ChatRichText = ({ content }: { content: JSONContent }) => (
  <div className={CHAT_CONTENT_CLASS}>{renderNode(content)}</div>
);

export const ChatPlainText = ({ text }: { text: string }) => (
  <p className={`${CHAT_CONTENT_CLASS} break-words whitespace-pre-wrap`}>
    {renderLinkedText(text)}
  </p>
);
