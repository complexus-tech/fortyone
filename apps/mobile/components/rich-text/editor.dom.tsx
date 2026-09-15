"use dom";

import { useEffect, useLayoutEffect, useRef, useState, type Ref } from "react";
import { colors, themeColors } from "../../constants/colors";
import {
  EditorContent,
  useEditor,
  useEditorState,
  type Editor,
} from "@tiptap/react";
import { useDOMImperativeHandle, type DOMProps } from "expo/dom";
import { createMobileRichTextExtensions, getRichTextValue } from "./extensions";
import { getUnsupportedRichText, sanitizeRichText } from "./sanitize";
import type { RichTextValue } from "./content";
import { MetadataIcon, type EditorMetadata } from "./metadata-icon";

export type MentionOption = { id: string; label: string };
const ICON_PATHS = {
  Close: "M6 6l12 12M18 6 6 18",
  Save: "M12 20V4m-6 6 6-6 6 6",
  Bold: "M6 4h7a4 4 0 0 1 0 8H6m0 0h8a4 4 0 0 1 0 8H6V4",
  Italic: "M10 4h9M5 20h9M15 4 9 20",
  Underline: "M6 3v7a6 6 0 0 0 12 0V3M4 21h16",
  "Bullet list": "M9 6h12M9 12h12M9 18h12M3 6h.01M3 12h.01M3 18h.01",
  "Numbered list":
    "M10 6h11M10 12h11M10 18h11M3 3h1v6M2 9h4M2 14c0-3 5-3 4 0l-4 4h4",
  Checklist: "M10 6h11M10 12h11M10 18h11M2 5l2 2 3-4M2 17l2 2 3-4",
} as const;
const FormattingIcon = ({ name }: { name: keyof typeof ICON_PATHS }) => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.8"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
  >
    <path d={ICON_PATHS[name]} />
  </svg>
);
export type EditorDOMHandle = {
  flush: (requestId: unknown) => void;
  resume: () => void;
};
type Props = {
  ref?: Ref<EditorDOMHandle>;
  initialHtml: string;
  dark: boolean;
  title: string;
  saveLabel: string;
  placeholder: string;
  mentions: MentionOption[];
  closeRequest: number;
  inline?: boolean;
  disabled?: boolean;
  readOnly?: boolean;
  metadata?: EditorMetadata[];
  onMetadataPress: (key: string) => Promise<void>;
  onFlushComplete: (
    requestId: number,
    value: RichTextValue | null,
    error: string | null,
  ) => Promise<void>;
  onRuntimeError: () => Promise<void>;
  onDraft: (value: RichTextValue) => Promise<void>;
  onSave: (value: RichTextValue) => Promise<{ error?: string }>;
  onClose: () => Promise<void>;
  onReady: (readOnly: boolean) => Promise<void>;
  onDiscard: () => Promise<void>;
  dom?: DOMProps;
};

type EditorOperation = {
  editor: Editor | null;
  callbacks: { current: Omit<Props, "ref"> };
  saving: { current: boolean };
  unsupported: boolean;
  setSaving: (saving: boolean) => void;
  setError: (message: string | null) => void;
};

function resumeEditor(operation: EditorOperation) {
  operation.saving.current = false;
  operation.setSaving(false);
  const { editor, callbacks, unsupported } = operation;
  if (editor && !editor.isDestroyed)
    editor.setEditable(
      !unsupported &&
        !callbacks.current.disabled &&
        !callbacks.current.readOnly,
      false,
    );
}

async function flushEditor(operation: EditorOperation, requestId: number) {
  const { editor, callbacks, saving, unsupported, setSaving, setError } =
    operation;
  try {
    if (!editor || saving.current)
      throw new Error("The editor is busy. Please try again.");
    if (unsupported || callbacks.current.readOnly)
      throw new Error(
        "This document is read-only. Your original draft is preserved.",
      );
    saving.current = true;
    setSaving(true);
    editor.setEditable(false, false);
    const value = getRichTextValue(editor);
    await callbacks.current.onDraft(value);
    await callbacks.current.onFlushComplete(requestId, value, null);
  } catch (cause) {
    const message =
      cause instanceof Error
        ? cause.message
        : "Could not save your draft. Please try again.";
    setError(message);
    await callbacks.current.onFlushComplete(requestId, null, message);
  }
}

async function closeEditor(operation: EditorOperation) {
  const { editor, callbacks, saving, unsupported, setSaving, setError } =
    operation;
  if (saving.current) return;
  saving.current = true;
  setSaving(true);
  editor?.setEditable(false, false);
  try {
    if (editor && !unsupported && !callbacks.current.readOnly)
      await callbacks.current.onDraft(getRichTextValue(editor));
    await callbacks.current.onClose();
  } catch (cause) {
    resumeEditor(operation);
    setError(
      cause instanceof Error
        ? cause.message
        : "Your draft could not be saved. Keep the editor open and try again.",
    );
  }
}

async function saveEditor(operation: EditorOperation) {
  const { editor, callbacks, saving, unsupported, setSaving, setError } =
    operation;
  if (!editor || saving.current || unsupported) return;
  saving.current = true;
  setSaving(true);
  setError(null);
  editor.setEditable(false, false);
  try {
    const value = getRichTextValue(editor);
    if (!callbacks.current.readOnly) await callbacks.current.onDraft(value);
    const result = await callbacks.current.onSave(value);
    if (result.error) setError(result.error);
  } catch (cause) {
    setError(
      cause instanceof Error
        ? cause.message
        : "Could not save your changes. Your draft is still here; try again.",
    );
  } finally {
    resumeEditor(operation);
  }
}

const CSS = `
*{box-sizing:border-box}html,body,#root{margin:0;padding:0;width:100%;height:100%;min-width:0;min-height:0;overflow:hidden;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}#root{display:flex;flex-direction:column}
body{overscroll-behavior:none}.editor-shell{--fg:${themeColors.light.foreground};--muted:${themeColors.light.textMuted};--bg:${themeColors.light.background};--line:${themeColors.light.border};--active:${themeColors.light.surfaceMuted};--accent:color-mix(in srgb,${colors.primary} 75%,${colors.black});color:var(--fg);background:var(--bg);width:100%;max-width:100%;height:100%;min-width:0;min-height:0;display:flex;flex-direction:column;overflow:hidden}
.editor-shell.dark{--fg:${themeColors.dark.foreground};--muted:${themeColors.dark.textMuted};--bg:${themeColors.dark.background};--line:${themeColors.dark.border};--active:${themeColors.dark.surfaceMuted};--accent:${colors.primary}}
.editor-shell.inline.dark{--bg:${themeColors.dark.surface};--active:${themeColors.dark.surfaceProminent}}
button,input{font:inherit}button{cursor:pointer;color:inherit;background:none;border:0;border-radius:10px;min-height:44px;padding:8px 12px;touch-action:manipulation}button:disabled{opacity:.4;cursor:default}button:focus-visible,input:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.editor-header{display:grid;grid-template-columns:44px minmax(0,1fr) 44px;align-items:center;gap:8px;padding:10px 20px;flex:none;min-width:0}.editor-header strong{font-size:16px;font-weight:600;text-align:center;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.editor-header button{display:grid;place-items:center;width:44px;height:44px;padding:0;border-radius:50%;background:var(--active)}
.editor-toolbar{display:grid;grid-template-columns:repeat(6,minmax(44px,1fr));align-items:center;gap:2px;min-width:0;width:100%;padding:5px 10px;border-top:1px solid var(--line);flex:none}.editor-toolbar button{display:grid;place-items:center;width:100%;min-width:44px;min-height:44px;white-space:nowrap;padding:8px}.editor-toolbar button[aria-pressed=true]{background:var(--active);color:var(--accent)}.editor-spinner{width:18px;height:18px;border:2px solid var(--line);border-top-color:var(--fg);border-radius:50%;animation:editor-spin .8s linear infinite}@keyframes editor-spin{to{transform:rotate(360deg)}}@media(prefers-reduced-motion:reduce){.editor-spinner{animation:none}}
.editor-document{flex:1;min-width:0;min-height:0;overflow:auto;-webkit-overflow-scrolling:touch;padding:12px 20px 24px;scroll-padding:24px}.editor-document>div{min-height:100%;min-width:0}.tiptap{outline:none;min-height:100%;line-height:1.6;font-size:16px;overflow-wrap:anywhere}.tiptap p{margin:.65em 0}.tiptap h1{font-size:26px}.tiptap h2{font-size:22px}.tiptap h3{font-size:19px}.tiptap h1,.tiptap h2,.tiptap h3{line-height:1.25;margin:1em 0 .45em;font-weight:600}
.tiptap a,.tiptap .mention{color:var(--accent)}.tiptap .mention{border-radius:5px;background:var(--active);padding:2px 3px}.tiptap blockquote{border-left:3px solid var(--line);padding-left:16px;margin:14px 0;color:var(--muted)}.tiptap pre{background:var(--active);border-radius:10px;padding:14px;white-space:pre-wrap}.tiptap code{background:var(--active);border-radius:4px;padding:2px 4px;font-size:.85em}.tiptap pre code{padding:0}.tiptap ul,.tiptap ol{padding-left:24px}.tiptap li>p{margin:3px 0}
.tiptap ul[data-type=taskList]{list-style:none;padding-left:0}.tiptap ul[data-type=taskList]>li{display:flex;gap:10px;align-items:flex-start}.tiptap ul[data-type=taskList]>li>label{padding-top:4px;flex:none}.tiptap ul[data-type=taskList]>li>div{flex:1;min-width:0}.tiptap input[type=checkbox]{width:18px;height:18px;accent-color:var(--accent)}
.tiptap img,.tiptap video{max-width:100%;height:auto;border-radius:12px}.tiptap table{border-collapse:collapse;table-layout:fixed;width:100%;margin:16px 0}.tiptap th,.tiptap td{border:1px solid var(--line);padding:4px 8px;vertical-align:top;min-width:50px}.tiptap th>:first-child,.tiptap td>:first-child{margin-top:0}.tiptap th>:last-child,.tiptap td>:last-child{margin-bottom:0}.tiptap th{background:var(--active)}.tiptap .selectedCell{background:var(--active)}.tiptap hr{border:0;border-top:1px solid var(--line);margin:24px 0}.tiptap p.is-editor-empty:first-child:before{content:attr(data-placeholder);float:left;color:var(--muted);pointer-events:none;height:0}
.editor-notice{margin:0;max-height:25%;overflow:auto;padding:12px 20px;color:var(--fg);background:var(--active);font-size:14px;line-height:1.45}.editor-error{color:var(--accent)}.editor-status{padding:2px 12px 2px 20px;font-size:12px;color:var(--muted);flex:none;min-width:0}.editor-status span{min-width:0;overflow-wrap:anywhere}.editor-status button{flex:none;font-size:13px}
.editor-metadata{display:flex;align-items:center;gap:8px;padding:8px 20px;overflow-x:auto;min-width:0;flex:none;scrollbar-width:none}.editor-metadata button{display:flex;align-items:center;gap:7px;flex:none;border-radius:24px;padding:8px 12px;background:var(--active);font-size:15px;white-space:nowrap}.editor-metadata button.muted{color:var(--muted)}.editor-metadata svg{flex:none}.metadata-avatar{display:inline-flex;align-items:center;justify-content:center;flex:none;width:18px;height:18px;border-radius:50%;overflow:hidden;font-size:9px;font-weight:600;line-height:1}.metadata-avatar img{width:100%;height:100%;object-fit:cover;object-position:top center}
.editor-shell.inline .editor-document{padding-top:0}.editor-shell.inline .editor-toolbar{margin:0 12px 8px;padding:5px 8px;border:0;border-radius:28px;background:var(--active);width:calc(100% - 24px)}
`;

export default function RichTextEditorDOM({ ref, ...props }: Props) {
  const [unsupported] = useState(() =>
    getUnsupportedRichText(props.initialHtml),
  );
  const [initialHtml] = useState(() => sanitizeRichText(props.initialHtml));
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const callbacks = useRef(props);
  useLayoutEffect(() => {
    callbacks.current = props;
  }, [props]);
  const savingRef = useRef(false);
  const editor = useEditor({
    extensions: createMobileRichTextExtensions(props.placeholder),
    content: initialHtml,
    editable: unsupported.length === 0 && !props.disabled && !props.readOnly,
    immediatelyRender: false,
    editorProps: {
      attributes: {
        role: "textbox",
        "aria-label": props.title,
        "aria-multiline": "true",
        spellcheck: "true",
        autocapitalize: "sentences",
      },
      transformPastedHTML: sanitizeRichText,
      handlePaste: (_view, event) => {
        const html = event.clipboardData?.getData("text/html");
        if (html && getUnsupportedRichText(html).length > 0) {
          setError(
            "This paste contains formatting we cannot preserve. Paste plain text or edit it on the web.",
          );
          return true;
        }
        if (event.clipboardData?.files.length) {
          setError(
            "Upload attachments from the web app. Existing images and videos stay in your document.",
          );
          return true;
        }
        return false;
      },
      handleDOMEvents: {
        click: (_view, event) => {
          if (event.target instanceof Element && event.target.closest("a"))
            event.preventDefault();
          return false;
        },
      },
    },
    onUpdate: ({ editor: current }) => {
      if (
        savingRef.current ||
        callbacks.current.readOnly ||
        callbacks.current.disabled
      )
        return;
      void callbacks.current
        .onDraft(getRichTextValue(current))
        .catch(() =>
          setError(
            "Could not save your draft on this device. Keep the editor open and try Save again.",
          ),
        );
    },
  });

  useDOMImperativeHandle<EditorDOMHandle>(
    ref ?? null,
    () => ({
      flush: (requestId) => {
        if (typeof requestId !== "number" || !Number.isInteger(requestId))
          return;
        void flushEditor(
          {
            editor,
            callbacks,
            saving: savingRef,
            unsupported: unsupported.length > 0,
            setSaving,
            setError,
          },
          requestId,
        );
      },
      resume: () =>
        resumeEditor({
          editor,
          callbacks,
          saving: savingRef,
          unsupported: unsupported.length > 0,
          setSaving,
          setError,
        }),
    }),
    [editor, unsupported],
  );

  useEffect(() => {
    if (editor) void callbacks.current.onReady(unsupported.length > 0);
  }, [editor, unsupported]);

  useEffect(() => {
    if (editor && !savingRef.current)
      editor.setEditable(
        unsupported.length === 0 && !props.disabled && !props.readOnly,
        false,
      );
  }, [editor, props.disabled, props.readOnly, unsupported]);

  useEffect(() => {
    const report = () => {
      void callbacks.current.onRuntimeError().catch(() => undefined);
    };
    window.addEventListener("error", report);
    window.addEventListener("unhandledrejection", report);
    return () => {
      window.removeEventListener("error", report);
      window.removeEventListener("unhandledrejection", report);
    };
  }, []);

  const state = useEditorState({
    editor,
    selector: ({ editor: current }) => ({
      bold: current?.isActive("bold") ?? false,
      italic: current?.isActive("italic") ?? false,
      underline: current?.isActive("underline") ?? false,
      bulletList: current?.isActive("bulletList") ?? false,
      orderedList: current?.isActive("orderedList") ?? false,
      taskList: current?.isActive("taskList") ?? false,
    }),
  });

  const close = () =>
    closeEditor({
      editor,
      callbacks,
      saving: savingRef,
      unsupported: unsupported.length > 0,
      setSaving,
      setError,
    });
  const closeRef = useRef(close);
  useLayoutEffect(() => {
    closeRef.current = close;
  });
  useEffect(() => {
    if (props.closeRequest > 0) void closeRef.current();
  }, [props.closeRequest]);

  const save = () =>
    saveEditor({
      editor,
      callbacks,
      saving: savingRef,
      unsupported: unsupported.length > 0,
      setSaving,
      setError,
    });

  const tool = (
    label: keyof typeof ICON_PATHS,
    active: boolean,
    action: () => void,
  ) => (
    <button
      key={label}
      type="button"
      aria-label={label}
      aria-pressed={active}
      disabled={
        !editor ||
        saving ||
        props.disabled ||
        props.readOnly ||
        unsupported.length > 0
      }
      onPointerDown={(event) => event.preventDefault()}
      onClick={() => {
        if (
          !savingRef.current &&
          !callbacks.current.disabled &&
          !callbacks.current.readOnly
        )
          action();
      }}
    >
      <FormattingIcon name={label} />
    </button>
  );

  return (
    <div
      className={`editor-shell${props.dark ? " dark" : ""}${props.inline ? " inline" : ""}`}
    >
      <style>{CSS}</style>
      {!props.inline && (
        <header className="editor-header">
          <button
            type="button"
            aria-label="Close editor"
            disabled={saving}
            onClick={() => void close()}
          >
            <FormattingIcon name="Close" />
          </button>
          <strong>{props.title}</strong>
          <button
            type="button"
            aria-label={saving ? "Saving changes" : props.saveLabel}
            aria-busy={saving}
            disabled={!editor || saving || unsupported.length > 0}
            onClick={() => void save()}
          >
            {saving ? (
              <span className="editor-spinner" aria-hidden="true" />
            ) : (
              <FormattingIcon name="Save" />
            )}
          </button>
        </header>
      )}
      {unsupported.length > 0 && (
        <p className="editor-notice" role="status">
          This document contains formatting that mobile cannot safely edit yet.
          Open it on the web to edit it. Your original content is preserved.
        </p>
      )}
      {error && (
        <p className="editor-notice editor-error" role="alert">
          {error}
        </p>
      )}
      <div className="editor-document">
        <EditorContent editor={editor} />
      </div>
      {!props.inline && (
        <div
          className="editor-status"
          aria-live="polite"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 8,
          }}
        >
          <span>
            {saving ? "Saving your changes…" : "Draft kept on this device"}
          </span>
          <button
            type="button"
            disabled={saving || props.readOnly}
            onClick={() =>
              void props
                .onDiscard()
                .catch(() =>
                  setError("Could not discard the draft. Try again."),
                )
            }
          >
            Discard draft
          </button>
        </div>
      )}
      {props.metadata && props.metadata.length > 0 && (
        <div
          className="editor-metadata"
          role="group"
          aria-label="Task properties"
        >
          {props.metadata.map((item) => (
            <button
              key={item.key}
              type="button"
              className={item.muted ? "muted" : undefined}
              disabled={saving || props.disabled || props.readOnly}
              aria-label={`Change ${item.label.toLowerCase()}: ${item.accessibilityValue ?? item.value}`}
              onPointerDown={(event) => event.preventDefault()}
              onClick={() => {
                if (
                  !savingRef.current &&
                  !callbacks.current.disabled &&
                  !callbacks.current.readOnly
                )
                  void props.onMetadataPress(item.key);
              }}
            >
              <MetadataIcon item={item} dark={props.dark} />
              {item.value}
            </button>
          ))}
        </div>
      )}
      <div
        className="editor-toolbar"
        role="toolbar"
        aria-label="Text formatting"
      >
        {tool("Bold", state?.bold ?? false, () => {
          editor?.chain().focus().toggleBold().run();
        })}
        {tool("Italic", state?.italic ?? false, () => {
          editor?.chain().focus().toggleItalic().run();
        })}
        {tool("Underline", state?.underline ?? false, () => {
          editor?.chain().focus().toggleUnderline().run();
        })}
        {tool("Bullet list", state?.bulletList ?? false, () => {
          editor?.chain().focus().toggleBulletList().run();
        })}
        {tool("Numbered list", state?.orderedList ?? false, () => {
          editor?.chain().focus().toggleOrderedList().run();
        })}
        {tool("Checklist", state?.taskList ?? false, () => {
          editor?.chain().focus().toggleTaskList().run();
        })}
      </div>
    </div>
  );
}
