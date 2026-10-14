import type { Editor, JSONContent } from "@tiptap/core";
import {
  getHTMLFromFragment,
  getText,
  getTextSerializersFromSchema,
  isNodeEmpty,
} from "@tiptap/core";

// ProseMirror documents are immutable. Retain the current document and schema
// without traversing them, so deferred saves survive editor reuse or destruction.
export const createRichTextSnapshot = (editor: Editor) => {
  const document = editor.state.doc;
  const schema = editor.schema;

  return {
    schema,
    state: { doc: document },
    get isEmpty() {
      return isNodeEmpty(document);
    },
    getHTML: () => getHTMLFromFragment(document.content, schema),
    getJSON: (): JSONContent => document.toJSON(),
    getText: () =>
      getText(document, {
        blockSeparator: "\n\n",
        textSerializers: getTextSerializersFromSchema(schema),
      }),
  };
};

export type RichTextSnapshot = ReturnType<typeof createRichTextSnapshot>;
