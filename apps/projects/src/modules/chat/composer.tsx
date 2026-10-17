"use client";

import type { Editor, JSONContent } from "@tiptap/core";
import type { ComponentType } from "react";
import { useEffect, useId, useMemo, useState } from "react";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Underline from "@tiptap/extension-underline";
import { EditorContent, useEditor, useEditorState } from "@tiptap/react";
import { SuggestionPluginKey } from "@tiptap/suggestion";
import {
  AtIcon,
  BoldIcon,
  CodeIcon,
  ItalicIcon,
  LinkIcon,
  OrderedListIcon,
  QuoteIcon,
  StrikeThroughIcon,
  UnderlineIcon,
  UnorderedListIcon,
} from "icons";
import { cn } from "lib";
import { Button, Input, Popover, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useChatCapabilities } from "@/lib/hooks/workspace/chat-capabilities";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import { getConversationMembers } from "./api";
import type { PreviewDraft } from "./chat-rich-text";
import { CHAT_CONTENT_CLASS, getChatLinkHref } from "./chat-rich-text";
import { useTypingActivity } from "./use-typing-activity";
import { ChatSendIcon } from "./send-icon";
import { VoiceRecorder } from "./voice-recorder";
import { ICON_BUTTON_CLASS } from "./chat-styles";
import {
  ChatMention,
  setChatMentionPeople,
  setChatMentionSearch,
} from "./mentions";
import { useChatIdentity } from "./chat-context";
import {
  AttachFileButton,
  ComposerAttachmentList,
  useComposerAttachments,
} from "./composer-attachments";

const MAX_MESSAGE_LENGTH = 4000;
const COMPOSER_BUTTON_CLASS =
  "size-7 rounded-sm text-text-muted [&_svg]:text-current";

const getDraftContent = (draft: PreviewDraft): JSONContent =>
  draft.content ?? {
    type: "doc",
    content: draft.text.split("\n").map((text) => ({
      type: "paragraph",
      content: text ? [{ type: "text", text }] : [],
    })),
  };

const FORMATTING_GROUPS = [
  [
    { name: "bold", label: "Bold", icon: BoldIcon },
    { name: "italic", label: "Italic", icon: ItalicIcon },
    { name: "underline", label: "Underline", icon: UnderlineIcon },
    { name: "strike", label: "Strikethrough", icon: StrikeThroughIcon },
  ],
  [
    { name: "orderedList", label: "Numbered list", icon: OrderedListIcon },
    { name: "bulletList", label: "Bulleted list", icon: UnorderedListIcon },
  ],
  [
    { name: "blockquote", label: "Quote", icon: QuoteIcon },
    { name: "code", label: "Inline code", icon: CodeIcon },
  ],
] as const;

type FormatName = (typeof FORMATTING_GROUPS)[number][number]["name"];

const toggleFormat = (editor: Editor, name: FormatName) => {
  const chain = editor.chain().focus();
  switch (name) {
    case "bold":
      return chain.toggleBold().run();
    case "italic":
      return chain.toggleItalic().run();
    case "underline":
      return chain.toggleUnderline().run();
    case "strike":
      return chain.toggleStrike().run();
    case "bulletList":
      return chain.toggleBulletList().run();
    case "orderedList":
      return chain.toggleOrderedList().run();
    case "blockquote":
      return chain.toggleBlockquote().run();
    case "code":
      return chain.toggleCode().run();
  }
};

const FormattingButton = ({
  active,
  icon: Icon,
  label,
  onClick,
  disabled,
}: {
  disabled: boolean;
  active: boolean;
  icon: ComponentType<{ className?: string }>;
  label: string;
  onClick: () => void;
}) => (
  <Tooltip title={label}>
    <Button
      aria-label={label}
      aria-pressed={active}
      asIcon
      className={cn(ICON_BUTTON_CLASS, COMPOSER_BUTTON_CLASS, {
        "bg-state-active text-foreground": active,
      })}
      color="tertiary"
      disabled={disabled}
      onClick={onClick}
      onMouseDown={(event) => {
        event.preventDefault();
      }}
      size="sm"
      type="button"
      variant="naked"
    >
      <Icon className="size-4" />
    </Button>
  </Tooltip>
);

const LinkControl = ({
  editor,
  disabled,
}: {
  editor: Editor;
  disabled: boolean;
}) => {
  const [open, setOpen] = useState(false);
  const [url, setUrl] = useState("");
  const [error, setError] = useState("");
  const [selection, setSelection] = useState({ from: 0, to: 0, linked: false });
  const errorId = useId();

  const applyLink = () => {
    const href = getChatLinkHref(url);
    if (!href) {
      setError("Enter a full https://, http:// or mailto: link.");
      return;
    }
    const { from, to, linked } = selection;
    const chain = editor.chain().focus().setTextSelection({ from, to });
    if (from === to && !linked) {
      chain
        .insertContent({
          type: "text",
          text: href,
          marks: [{ type: "link", attrs: { href } }],
        })
        .run();
    } else {
      chain.extendMarkRange("link").setLink({ href }).run();
    }
    setOpen(false);
  };

  return (
    <Popover
      onOpenChange={(nextOpen) => {
        if (nextOpen) {
          const { from, to } = editor.state.selection;
          setSelection({ from, to, linked: editor.isActive("link") });
          setUrl(String(editor.getAttributes("link").href ?? ""));
          setError("");
        }
        setOpen(nextOpen);
      }}
      open={open}
    >
      <Tooltip title="Add link">
        <Popover.Trigger asChild>
          <Button
            aria-label="Add link"
            asIcon
            className={cn(ICON_BUTTON_CLASS, COMPOSER_BUTTON_CLASS)}
            color="tertiary"
            disabled={disabled}
            onMouseDown={(event) => {
              event.preventDefault();
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            <LinkIcon className="size-4" />
          </Button>
        </Popover.Trigger>
      </Tooltip>
      <Popover.Content
        align="start"
        aria-label="Add a link to your message"
        className="w-80 max-w-[calc(100vw-2rem)] p-3"
        onCloseAutoFocus={(event) => {
          event.preventDefault();
          editor.commands.focus();
        }}
        side="top"
      >
        <Input
          aria-describedby={error ? errorId : undefined}
          aria-invalid={Boolean(error)}
          label="Link address"
          labelClassName="font-medium"
          onChange={(event) => {
            setUrl(event.target.value);
            setError("");
          }}
          onKeyDown={(event) => {
            if (event.key === "Enter" && !event.nativeEvent.isComposing) {
              event.preventDefault();
              event.stopPropagation();
              applyLink();
            }
          }}
          placeholder="https://"
          size="sm"
          value={url}
        />
        {error ? (
          <p className="text-danger mt-2" id={errorId} role="alert">
            {error}
          </p>
        ) : null}
        <div className="mt-3 flex justify-end gap-2">
          {selection.linked ? (
            <Button
              color="tertiary"
              onClick={() => {
                editor
                  .chain()
                  .focus()
                  .setTextSelection(selection)
                  .extendMarkRange("link")
                  .unsetLink()
                  .run();
                setOpen(false);
              }}
              size="sm"
              type="button"
              variant="naked"
            >
              Remove
            </Button>
          ) : null}
          <Button onClick={applyLink} size="sm" type="button">
            Apply link
          </Button>
        </div>
      </Popover.Content>
    </Popover>
  );
};

export const Composer = ({
  label,
  draft,
  onDraftChange,
  onSend,
  allowMaya,
  intent = "send",
  onCancel,
  autoFocus,
  disabled = false,
  threadRootId,
}: {
  label: string;
  draft: PreviewDraft;
  onDraftChange: (value: PreviewDraft) => void;
  onSend: () => void;
  allowMaya: boolean;
  intent?: "send" | "edit";
  onCancel?: () => void;
  autoFocus?: boolean;
  disabled?: boolean;
  threadRootId?: string;
}) => {
  const identity = useChatIdentity();
  const { attachmentsEnabled } = useChatCapabilities();
  const hasRecipient = Boolean(identity.conversationId);
  const { workspaceSlug } = useWorkspacePath();
  const isEditing = intent === "edit";
  const typing = useTypingActivity(
    identity.conversationId,
    threadRootId,
    !isEditing,
  );
  const { uploading, uploadFiles, removeFile } = useComposerAttachments(
    draft,
    onDraftChange,
    isEditing,
  );
  const busy = disabled || uploading;
  const [voiceActive, setVoiceActive] = useState(false);
  const [showFormatting, setShowFormatting] = useState(true);
  const formattingId = useId();
  const helperId = useId();
  const extensions = useMemo(
    () => [
      createRichTextStarterKit().configure({
        heading: false,
        horizontalRule: false,
      }),
      Underline,
      ChatMention,
      Link.configure({
        autolink: true,
        linkOnPaste: true,
        defaultProtocol: "https",
        openOnClick: false,
        isAllowedUri: (url) => Boolean(getChatLinkHref(url)),
      }),
      Placeholder.configure({ placeholder: label }),
    ],
    [label],
  );
  const editor = useEditor({
    extensions,
    content: getDraftContent(draft),
    autofocus: autoFocus ?? isEditing ? "end" : false,
    immediatelyRender: false,
    shouldRerenderOnTransaction: false,
    editable: !busy,
    editorProps: {
      attributes: {
        "aria-label": label,
        "aria-describedby": helperId,
        "aria-multiline": "true",
        role: "textbox",
        class: cn(
          CHAT_CONTENT_CLASS,
          "max-h-48 min-h-[44px] overflow-y-auto px-3 py-[5px] outline-none [&_p]:my-0 [&>p.is-editor-empty:first-child]:before:text-text-muted [&>p.is-editor-empty:first-child]:before:pointer-events-none [&>p.is-editor-empty:first-child]:before:float-left [&>p.is-editor-empty:first-child]:before:h-0 [&>p.is-editor-empty:first-child]:before:content-[attr(data-placeholder)]",
        ),
      },
      handleKeyDown: (view, event) => {
        if (busy) return true;
        if (!hasRecipient) return false;
        if (SuggestionPluginKey.getState(view.state)?.active) return false;
        if (
          isEditing &&
          onCancel &&
          event.key === "Escape" &&
          !event.isComposing &&
          !view.composing
        ) {
          event.preventDefault();
          onCancel();
          return true;
        }
        if (
          event.key !== "Enter" ||
          event.shiftKey ||
          event.isComposing ||
          view.composing ||
          event.keyCode === 229
        ) {
          return false;
        }
        const isEditingBlock = ["listItem", "blockquote", "codeBlock"].some(
          (type) => editor?.isActive(type),
        );
        if (isEditingBlock && !event.metaKey && !event.ctrlKey) return false;

        event.preventDefault();
        if (voiceActive) return true;
        const text = editor?.getText({ blockSeparator: "\n" }) ?? "";
        if (
          (text.trim() || draft.attachments?.length) &&
          text.length <= MAX_MESSAGE_LENGTH
        ) {
          typing.stop();
          onSend();
        }
        return true;
      },
    },
    onBlur: () => {
      typing.stop();
    },
    onUpdate: ({ editor: current }) => {
      if (current.getText().trim()) typing.changed();
      else typing.stop();
      onDraftChange({
        ...draft,
        text: current.getText({ blockSeparator: "\n" }),
        content: current.getJSON(),
      });
    },
  });
  useEffect(() => {
    if (!editor) return;
    editor.setEditable(!busy, false);
    const toMentionItems = (people: typeof identity.members) =>
      people
        .filter(
          (person) => !person.isSystem || (allowMaya && identity.mayaEnabled),
        )
        .map((person) => ({
          id: person.id,
          label: person.fullName,
          username: person.fullName,
          avatar: person.avatarUrl,
          ...(person.isSystem
            ? { mentionKind: "agent" as const, agentKey: "maya" as const }
            : { mentionKind: "user" as const }),
        }));
    setChatMentionPeople(editor, toMentionItems(identity.members));
    setChatMentionSearch(editor, async (query, signal) => {
      if (!identity.conversationId) return [];
      const result = await getConversationMembers(
        { workspaceSlug },
        identity.conversationId,
        query,
        signal,
      );
      return toMentionItems(result.items);
    });
  }, [allowMaya, busy, editor, identity, workspaceSlug]);
  const editorState = useEditorState({
    editor,
    selector: ({ editor: current }) => ({
      active: Object.fromEntries(
        FORMATTING_GROUPS.flat().map(({ name }) => [
          name,
          current?.isActive(name) ?? false,
        ]),
      ),
      editingBlock: ["listItem", "blockquote", "codeBlock"].some((type) =>
        current?.isActive(type),
      ),
    }),
  });

  useEffect(() => {
    if (!editor) return;
    const content = getDraftContent(draft);
    if (JSON.stringify(editor.getJSON()) !== JSON.stringify(content)) {
      editor.commands.setContent(content, { emitUpdate: false });
    }
  }, [draft, editor]);

  const overLimit = draft.text.length > MAX_MESSAGE_LENGTH;
  const canSubmit =
    hasRecipient &&
    Boolean(draft.text.trim() || draft.attachments?.length) &&
    !overLimit &&
    !voiceActive &&
    !busy;
  let sessionHint = hasRecipient
    ? "Messages are shared with this conversation."
    : "Choose a person or team before sending your message.";
  if (isEditing) {
    sessionHint = "Editing a message. Save changes when you are ready.";
  }
  let keyboardHint = isEditing
    ? "Enter to save changes. Shift + Enter for a new line."
    : "Shift + Enter for a new line";
  if (editorState?.editingBlock) {
    keyboardHint = `⌘ / Ctrl + Enter to ${isEditing ? "save changes" : "send"}`;
  }

  return (
    <div className={cn("shrink-0", isEditing ? "mt-2" : "px-[20px] pt-2")}>
      <form
        aria-label={isEditing ? "Edit message" : undefined}
        className="border-border/40 bg-surface focus-within:border-border-strong dark:bg-surface-elevated overflow-hidden rounded-xl border transition-colors"
        onSubmit={(event) => {
          event.preventDefault();
          if (canSubmit) {
            typing.stop();
            onSend();
          }
        }}
      >
        {showFormatting ? (
          <div
            aria-label="Message formatting"
            className="flex min-h-[38px] flex-wrap items-center px-1.5 py-[5px]"
            id={formattingId}
            role="group"
          >
            {FORMATTING_GROUPS.map((group, index) => (
              <div className="flex items-center" key={group[0].name}>
                {index > 0 ? (
                  <span className="bg-border-strong mx-[5px] h-4 w-px" />
                ) : null}
                {index === 1 && editor ? (
                  <>
                    <LinkControl disabled={busy} editor={editor} />
                    <span className="bg-border-strong mx-[5px] h-4 w-px" />
                  </>
                ) : null}
                {group.map(({ name, label: formatLabel, icon }) => (
                  <FormattingButton
                    active={editorState?.active[name] ?? false}
                    disabled={busy}
                    icon={icon}
                    key={name}
                    label={formatLabel}
                    onClick={() => {
                      if (editor) toggleFormat(editor, name);
                    }}
                  />
                ))}
              </div>
            ))}
          </div>
        ) : null}
        <EditorContent
          className={cn(!showFormatting && "pt-[8px]")}
          editor={editor}
        />
        {draft.attachments?.length ? (
          <ComposerAttachmentList
            attachments={draft.attachments}
            busy={busy}
            onRemove={removeFile}
          />
        ) : null}
        {uploading ? (
          <p className="text-text-muted px-3 pb-2" role="status">
            Uploading files…
          </p>
        ) : null}
        <div
          className={cn(
            "flex items-center justify-between gap-2 px-3 pt-1 pb-3",
            isEditing ? "min-h-[44px] flex-wrap" : "min-h-[44px]",
          )}
        >
          <div className="flex items-center gap-1">
            {attachmentsEnabled ? (
              <AttachFileButton
                busy={busy}
                onUpload={uploadFiles}
                uploading={uploading}
              />
            ) : null}
            <Tooltip
              title={showFormatting ? "Hide formatting" : "Show formatting"}
            >
              <Button
                aria-controls={showFormatting ? formattingId : undefined}
                aria-expanded={showFormatting}
                aria-label="Formatting options"
                asIcon
                className={cn(ICON_BUTTON_CLASS, COMPOSER_BUTTON_CLASS)}
                color="tertiary"
                onClick={() => {
                  setShowFormatting((value) => !value);
                }}
                size="sm"
                type="button"
                variant="naked"
              >
                <span
                  aria-hidden="true"
                  className="text-[16px] leading-5 underline underline-offset-2"
                >
                  Aa
                </span>
              </Button>
            </Tooltip>
            {identity.members.length > 0 ? (
              <Tooltip title="Mention someone">
                <Button
                  aria-label="Mention someone"
                  asIcon
                  className={cn(ICON_BUTTON_CLASS, COMPOSER_BUTTON_CLASS)}
                  color="tertiary"
                  disabled={busy}
                  onClick={() => {
                    editor
                      ?.chain()
                      .focus()
                      .insertContent({ type: "text", text: "@" })
                      .run();
                  }}
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  <AtIcon className="size-4" />
                </Button>
              </Tooltip>
            ) : null}
            {!isEditing && attachmentsEnabled ? (
              <VoiceRecorder
                disabled={
                  busy ||
                  !hasRecipient ||
                  (draft.attachments?.length ?? 0) >= 10
                }
                onActiveChange={setVoiceActive}
                onAttach={(file, durationSeconds) =>
                  uploadFiles([file], {
                    kind: "voice-note",
                    durationMs: Math.max(
                      1,
                      Math.min(300_000, Math.round(durationSeconds * 1000)),
                    ),
                  })
                }
              />
            ) : null}
          </div>
          <div className="ml-auto flex items-center gap-3">
            {draft.text.length >= 3500 ? (
              <span
                className={cn(
                  "text-text-muted tabular-nums",
                  overLimit && "text-danger",
                )}
              >
                {draft.text.length.toLocaleString()} / 4,000
              </span>
            ) : null}
            {isEditing ? (
              <div className="flex items-center gap-2">
                <Button
                  color="tertiary"
                  disabled={!onCancel || busy}
                  onClick={onCancel}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  Cancel
                </Button>
                <Button disabled={!canSubmit} size="sm" type="submit">
                  Save changes
                </Button>
              </div>
            ) : (
              <Button
                aria-label="Send message"
                asIcon
                className={cn(
                  ICON_BUTTON_CLASS,
                  COMPOSER_BUTTON_CLASS,
                  "size-8 p-1",
                )}
                color="tertiary"
                disabled={!canSubmit}
                size="sm"
                title="Send message"
                type="submit"
                variant="naked"
              >
                <ChatSendIcon className="size-6" />
              </Button>
            )}
          </div>
        </div>
      </form>
      <div className="sr-only" id={helperId}>
        <span>
          {overLimit
            ? "Keep your message within 4,000 characters"
            : sessionHint}
        </span>{" "}
        <span>
          {keyboardHint}
          {isEditing && onCancel ? " Escape to cancel." : null}
        </span>
      </div>
    </div>
  );
};
