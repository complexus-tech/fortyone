import { generateJSON } from "@tiptap/html/server";
import { markdownDocumentExtensions, markdownManager } from "./markdown";

export const richTextHTMLToMarkdown = (html: string) =>
  markdownManager.serialize(generateJSON(html, markdownDocumentExtensions));
