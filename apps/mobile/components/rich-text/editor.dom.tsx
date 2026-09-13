"use dom";

import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { colors } from "../../constants/colors";
import { EditorContent, useEditor, useEditorState } from "@tiptap/react";
import type { DOMProps } from "expo/dom";
import { createMobileRichTextExtensions, getRichTextValue } from "./extensions";
import { getUnsupportedRichText, sanitizeRichText } from "./sanitize";
import { isSafeLink, type RichTextValue } from "./content";

export type MentionOption = { id: string; label: string };

type Props = {
  initialHtml: string;
  dark: boolean;
  title: string;
  saveLabel: string;
  placeholder: string;
  mentions: MentionOption[];
  closeRequest: number;
  onDraft: (value: RichTextValue) => Promise<void>;
  onSave: (value: RichTextValue) => Promise<{ error?: string }>;
  onClose: () => Promise<void>;
  onReady: () => Promise<void>;
  onDiscard: () => Promise<void>;
  dom?: DOMProps;
};

const CSS = `
*{box-sizing:border-box}html,body,#root{margin:0;height:100%;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
body{overscroll-behavior:none}.editor-shell{--fg:${colors.black};--muted:${colors.gray.DEFAULT};--bg:${colors.white};--line:${colors.gray[100]};--active:${colors.gray[50]};--accent:color-mix(in srgb,${colors.primary} 75%,${colors.black});color:var(--fg);background:var(--bg);height:100dvh;display:flex;flex-direction:column}
.editor-shell.dark{--fg:${colors.white};--muted:${colors.gray[300]};--bg:${colors.dark.DEFAULT};--line:${colors.dark[50]};--active:${colors.dark[100]};--accent:${colors.primary}}
button,input{font:inherit}button{cursor:pointer;color:inherit;background:none;border:0;border-radius:10px;min-height:44px;padding:8px 12px;touch-action:manipulation}button:disabled{opacity:.4;cursor:default}button:focus-visible,input:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.editor-header{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:7px 8px;border-bottom:1px solid var(--line);flex:none}.editor-header strong{font-size:16px;text-align:center}.editor-header button{font-size:15px;color:var(--accent);font-weight:600}
.editor-toolbar{display:flex;gap:3px;overflow-x:auto;padding:5px 10px;border-bottom:1px solid var(--line);flex:none;scrollbar-width:none}.editor-toolbar button{min-width:44px;white-space:nowrap;flex-shrink:0;font-size:16px}.editor-toolbar button[aria-pressed=true]{background:var(--active);color:var(--accent)}
.editor-document{flex:1;overflow:auto;-webkit-overflow-scrolling:touch;padding:12px 20px 60px;scroll-padding:30px}.tiptap{outline:none;min-height:100%;line-height:1.6;font-size:17px;overflow-wrap:anywhere}.tiptap p{margin:.65em 0}.tiptap h1{font-size:29px}.tiptap h2{font-size:24px}.tiptap h3{font-size:21px}.tiptap h1,.tiptap h2,.tiptap h3{line-height:1.25;margin:1em 0 .45em}
.tiptap a,.tiptap .mention{color:var(--accent)}.tiptap .mention{border-radius:5px;background:var(--active);padding:2px 3px}.tiptap blockquote{border-left:3px solid var(--line);padding-left:16px;margin:14px 0;color:var(--muted)}.tiptap pre{background:var(--active);border-radius:10px;padding:14px;white-space:pre-wrap}.tiptap code{background:var(--active);border-radius:4px;padding:2px 4px;font-size:.85em}.tiptap pre code{padding:0}.tiptap ul,.tiptap ol{padding-left:24px}.tiptap li>p{margin:3px 0}
.tiptap ul[data-type=taskList]{list-style:none;padding-left:0}.tiptap ul[data-type=taskList]>li{display:flex;gap:10px;align-items:flex-start}.tiptap ul[data-type=taskList]>li>label{padding-top:4px;flex:none}.tiptap ul[data-type=taskList]>li>div{flex:1;min-width:0}.tiptap input[type=checkbox]{width:18px;height:18px;accent-color:var(--accent)}
.tiptap img,.tiptap video{max-width:100%;height:auto;border-radius:12px}.tiptap table{border-collapse:collapse;table-layout:fixed;width:100%;margin:16px 0}.tiptap th,.tiptap td{border:1px solid var(--line);padding:8px;vertical-align:top;min-width:50px}.tiptap th{background:var(--active)}.tiptap .selectedCell{background:var(--active)}.tiptap hr{border:0;border-top:1px solid var(--line);margin:24px 0}.tiptap p.is-editor-empty:first-child:before{content:attr(data-placeholder);float:left;color:var(--muted);pointer-events:none;height:0}
.editor-notice{margin:0;padding:12px 16px;color:var(--fg);background:var(--active);font-size:14px;line-height:1.45}.editor-error{color:var(--accent)}.editor-status{padding:7px 16px;font-size:12px;color:var(--muted);border-top:1px solid var(--line);flex:none}
.editor-picker{padding:10px 16px;border-bottom:1px solid var(--line);flex:none}.editor-picker input{width:100%;border:1px solid var(--line);border-radius:10px;min-height:44px;padding:10px;color:var(--fg);background:var(--bg)}.editor-picker-actions{display:flex;justify-content:flex-end;gap:8px}.editor-people{max-height:180px;overflow:auto}.editor-people button{display:block;width:100%;text-align:left}
`;

export default function RichTextEditorDOM(props: Props) {
  const [unsupported] = useState(() =>
    getUnsupportedRichText(props.initialHtml),
  );
  const [initialHtml] = useState(() => sanitizeRichText(props.initialHtml));
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [picker, setPicker] = useState<"link" | "mention" | null>(null);
  const [input, setInput] = useState("");
  const callbacks = useRef(props);
  useLayoutEffect(() => {
    callbacks.current = props;
  }, [props]);
  const savingRef = useRef(false);
  const editor = useEditor({
    extensions: createMobileRichTextExtensions(props.placeholder),
    content: initialHtml,
    editable: unsupported.length === 0,
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
    onCreate: () => {
      void callbacks.current.onReady();
    },
    onUpdate: ({ editor: current }) => {
      void callbacks.current
        .onDraft(getRichTextValue(current))
        .catch(() =>
          setError(
            "Could not save your draft on this device. Keep the editor open and try Save again.",
          ),
        );
    },
  });

  const state = useEditorState({
    editor,
    selector: ({ editor: current }) => ({
      bold: current?.isActive("bold") ?? false,
      italic: current?.isActive("italic") ?? false,
      underline: current?.isActive("underline") ?? false,
      strike: current?.isActive("strike") ?? false,
      heading: current?.isActive("heading", { level: 2 }) ?? false,
      bulletList: current?.isActive("bulletList") ?? false,
      orderedList: current?.isActive("orderedList") ?? false,
      taskList: current?.isActive("taskList") ?? false,
      blockquote: current?.isActive("blockquote") ?? false,
      codeBlock: current?.isActive("codeBlock") ?? false,
      table: current?.isActive("table") ?? false,
      link: current?.isActive("link") ?? false,
      canUndo: current?.can().undo() ?? false,
      canRedo: current?.can().redo() ?? false,
    }),
  });

  const close = async () => {
    if (savingRef.current) return;
    try {
      if (editor && unsupported.length === 0)
        await callbacks.current.onDraft(getRichTextValue(editor));
      await callbacks.current.onClose();
    } catch {
      setError(
        "Your draft could not be saved. Keep the editor open and try again.",
      );
    }
  };
  const closeRef = useRef(close);
  useLayoutEffect(() => {
    closeRef.current = close;
  });
  useEffect(() => {
    if (props.closeRequest > 0) void closeRef.current();
  }, [props.closeRequest]);

  const save = async () => {
    if (!editor || savingRef.current || unsupported.length > 0) return;
    savingRef.current = true;
    setSaving(true);
    setError(null);
    editor.setEditable(false);
    try {
      const value = getRichTextValue(editor);
      await callbacks.current.onDraft(value);
      const result = await callbacks.current.onSave(value);
      if (result.error) setError(result.error);
    } catch {
      setError(
        "Could not save your changes. Your draft is still here; try again.",
      );
    } finally {
      savingRef.current = false;
      setSaving(false);
      if (!editor.isDestroyed) editor.setEditable(true);
    }
  };

  const tool = (
    label: string,
    content: string,
    active: boolean,
    action: () => void,
    disabled = false,
  ) => (
    <button
      key={label}
      type="button"
      aria-label={label}
      aria-pressed={active}
      disabled={!editor || saving || unsupported.length > 0 || disabled}
      onPointerDown={(event) => event.preventDefault()}
      onClick={action}
    >
      {content}
    </button>
  );

  return (
    <div className={`editor-shell${props.dark ? " dark" : ""}`}>
      <style>{CSS}</style>
      <header className="editor-header">
        <button type="button" disabled={saving} onClick={() => void close()}>
          Close
        </button>
        <strong>{props.title}</strong>
        <button
          type="button"
          disabled={!editor || saving || unsupported.length > 0}
          onClick={() => void save()}
        >
          {saving ? "Saving…" : props.saveLabel}
        </button>
      </header>
      <div
        className="editor-toolbar"
        role="toolbar"
        aria-label="Text formatting"
      >
        {tool("Bold", "B", state?.bold ?? false, () => {
          editor?.chain().focus().toggleBold().run();
        })}
        {tool("Italic", "𝑰", state?.italic ?? false, () => {
          editor?.chain().focus().toggleItalic().run();
        })}
        {tool("Underline", "U̲", state?.underline ?? false, () => {
          editor?.chain().focus().toggleUnderline().run();
        })}
        {tool("Strikethrough", "S̶", state?.strike ?? false, () => {
          editor?.chain().focus().toggleStrike().run();
        })}
        {tool("Heading", "H2", state?.heading ?? false, () => {
          editor?.chain().focus().toggleHeading({ level: 2 }).run();
        })}
        {tool("Bullet list", "• List", state?.bulletList ?? false, () => {
          editor?.chain().focus().toggleBulletList().run();
        })}
        {tool("Numbered list", "1. List", state?.orderedList ?? false, () => {
          editor?.chain().focus().toggleOrderedList().run();
        })}
        {tool("Checklist", "☑", state?.taskList ?? false, () => {
          editor?.chain().focus().toggleTaskList().run();
        })}
        {tool("Quote", "❝", state?.blockquote ?? false, () => {
          editor?.chain().focus().toggleBlockquote().run();
        })}
        {tool("Code block", "</>", state?.codeBlock ?? false, () => {
          editor?.chain().focus().toggleCodeBlock().run();
        })}
        {tool("Link", "Link", state?.link ?? false, () => {
          setInput(editor?.getAttributes("link").href ?? "");
          setPicker("link");
        })}
        {props.mentions.length > 0 &&
          tool("Mention someone", "@", picker === "mention", () => {
            setInput("");
            setPicker("mention");
          })}
        {tool(
          "Insert table",
          "Table",
          state?.table ?? false,
          () => {
            editor
              ?.chain()
              .focus()
              .insertTable({ rows: 3, cols: 2, withHeaderRow: true })
              .run();
          },
          state?.table,
        )}
        {tool(
          "Undo",
          "↶",
          false,
          () => {
            editor?.chain().focus().undo().run();
          },
          !state?.canUndo,
        )}
        {tool(
          "Redo",
          "↷",
          false,
          () => {
            editor?.chain().focus().redo().run();
          },
          !state?.canRedo,
        )}
      </div>
      {state?.table && (
        <div
          className="editor-toolbar"
          role="toolbar"
          aria-label="Table formatting"
        >
          {tool("Add row below", "+ Row", false, () => {
            editor?.chain().focus().addRowAfter().run();
          })}
          {tool("Add column right", "+ Column", false, () => {
            editor?.chain().focus().addColumnAfter().run();
          })}
          {tool("Delete row", "− Row", false, () => {
            editor?.chain().focus().deleteRow().run();
          })}
          {tool("Delete column", "− Column", false, () => {
            editor?.chain().focus().deleteColumn().run();
          })}
        </div>
      )}
      {picker && (
        <div className="editor-picker">
          <input
            aria-label={picker === "link" ? "Link URL" : "Find a person"}
            placeholder={
              picker === "link" ? "https://example.com" : "Find a person"
            }
            value={input}
            onChange={(event) => setInput(event.target.value)}
            autoFocus
          />
          {picker === "mention" && (
            <div className="editor-people">
              {props.mentions
                .filter((person) =>
                  person.label.toLowerCase().includes(input.toLowerCase()),
                )
                .slice(0, 20)
                .map((person) => (
                  <button
                    type="button"
                    key={person.id}
                    onClick={() => {
                      editor
                        ?.chain()
                        .focus()
                        .insertContent([
                          { type: "mention", attrs: person },
                          { type: "text", text: " " },
                        ])
                        .run();
                      setPicker(null);
                    }}
                  >
                    @{person.label}
                  </button>
                ))}
            </div>
          )}
          <div className="editor-picker-actions">
            <button type="button" onClick={() => setPicker(null)}>
              Cancel
            </button>
            {picker === "link" && (
              <>
                <button
                  type="button"
                  onClick={() => {
                    editor
                      ?.chain()
                      .focus()
                      .extendMarkRange("link")
                      .unsetLink()
                      .run();
                    setPicker(null);
                  }}
                >
                  Remove link
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (!isSafeLink(input.trim())) {
                      setError("Enter a complete http, https, or mailto link.");
                      return;
                    }
                    editor
                      ?.chain()
                      .focus()
                      .extendMarkRange("link")
                      .setLink({ href: input.trim() })
                      .run();
                    setError(null);
                    setPicker(null);
                  }}
                >
                  Apply link
                </button>
              </>
            )}
          </div>
        </div>
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
          disabled={saving}
          onClick={() =>
            void props
              .onDiscard()
              .catch(() => setError("Could not discard the draft. Try again."))
          }
        >
          Discard draft
        </button>
      </div>
    </div>
  );
}
