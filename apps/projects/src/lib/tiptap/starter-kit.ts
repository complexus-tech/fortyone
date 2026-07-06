import StarterKit from "@tiptap/starter-kit";

export const createRichTextStarterKit = () =>
  StarterKit.configure({
    link: false,
    underline: false,
  });
