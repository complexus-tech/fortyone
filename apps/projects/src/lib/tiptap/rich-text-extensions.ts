import type { Node, Editor } from "@tiptap/core";
import { createDocumentExtensions } from "@fortyone/document-editor";
import Placeholder from "@tiptap/extension-placeholder";
import { SlashCommand } from "./slash-command";
import { RichTextMarkdown, RichTextMarkdownPaste } from "./markdown";
import {
  RichTextImage,
  RichTextMediaDrop,
  RichTextVideo,
} from "./rich-text-media";

type CreateRichTextExtensionsOptions = {
  onMediaFiles?: (editor: Editor, files: File[], position?: number) => void;
  onMediaRequest?: (editor: Editor) => void;
  placeholder: string;
  collaborative?: boolean;
  mention?: Node;
  reference?: Node;
};

export const createRichTextExtensions = ({
  onMediaFiles,
  onMediaRequest,
  placeholder,
  collaborative = false,
  mention,
  reference,
}: CreateRichTextExtensionsOptions) => [
  ...createDocumentExtensions({
    collaborative,
    image: RichTextImage,
    video: RichTextVideo,
    ...(mention ? { mention } : {}),
    ...(reference ? { reference } : {}),
  }),
  RichTextMarkdown,
  RichTextMarkdownPaste,
  ...(onMediaFiles
    ? [RichTextMediaDrop.configure({ onFiles: onMediaFiles })]
    : []),
  Placeholder.configure({ placeholder }),
  SlashCommand.configure({ onMediaRequest: onMediaRequest ?? null }),
];
