import type { Editor } from "@tiptap/react";
import { useEditorState } from "@tiptap/react";
import { BoldIcon, CodeIcon, ItalicIcon, UnorderedListIcon } from "icons";
import { Button } from "ui";

export function DocumentCommentFormatting({
  editor,
  disabled,
}: {
  editor: Editor;
  disabled: boolean;
}) {
  const active = useEditorState({
    editor,
    selector: ({ editor: current }) => ({
      bold: current.isActive("bold"),
      italic: current.isActive("italic"),
      code: current.isActive("code"),
      list: current.isActive("bulletList"),
    }),
  });
  const actions = [
    {
      label: "Bold",
      icon: BoldIcon,
      active: active.bold,
      run: () => editor.chain().focus().toggleBold().run(),
    },
    {
      label: "Italic",
      icon: ItalicIcon,
      active: active.italic,
      run: () => editor.chain().focus().toggleItalic().run(),
    },
    {
      label: "Code",
      icon: CodeIcon,
      active: active.code,
      run: () => editor.chain().focus().toggleCode().run(),
    },
    {
      label: "Bullet list",
      icon: UnorderedListIcon,
      active: active.list,
      run: () => editor.chain().focus().toggleBulletList().run(),
    },
  ];
  return (
    <div
      aria-label="Comment formatting"
      className="flex items-center gap-0.5"
      role="group"
    >
      {actions.map(({ label, icon: Icon, active: pressed, run }) => (
        <Button
          active={pressed}
          aria-label={label}
          aria-pressed={pressed}
          asIcon
          color="tertiary"
          disabled={disabled}
          key={label}
          onClick={run}
          onMouseDown={(event) => {
            event.preventDefault();
          }}
          size="sm"
          title={label}
          type="button"
          variant="naked"
        >
          <Icon className="size-4" strokeWidth={1.8} />
        </Button>
      ))}
    </div>
  );
}
