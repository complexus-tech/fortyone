import Link from "@tiptap/extension-link";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import Placeholder from "@tiptap/extension-placeholder";
import Underline from "@tiptap/extension-underline";
import { createRichTextStarterKit } from "./starter-kit";

export const getCommentEditorBaseExtensions = () => [
  createRichTextStarterKit(),
  Underline,
  TaskList,
  TaskItem.configure({
    nested: true,
  }),
  Link.configure({
    autolink: true,
  }),
];

export const getCommentEditorExtensions = ({
  placeholder,
}: {
  placeholder: string;
}) => [
  ...getCommentEditorBaseExtensions(),
  Placeholder.configure({
    placeholder,
  }),
];
