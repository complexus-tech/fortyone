import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import { Markdown } from "@tiptap/markdown";
import { createDocumentExtensions } from "@fortyone/document-editor";
import { DocumentCommentMention } from "./document-mentions";
import { renderDocumentComment } from "./document-comment-markdown";

const id = "1d4f475b-3dfa-458f-a315-0513149e481a";

test("comment mention identities survive Markdown draft persistence and rehydration", () => {
  const editor = new Editor({
    extensions: [StarterKit, DocumentCommentMention, Markdown],
  });
  try {
    editor.commands.setContent({
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            {
              type: "mention",
              attrs: { id, label: "Joseph [Design]", mentionKind: "user" },
            },
            { type: "text", text: " ask " },
            {
              type: "mention",
              attrs: {
                id,
                label: "Maya",
                mentionKind: "agent",
                agentKey: "maya",
              },
            },
          ],
        },
      ],
    });
    const markdown = editor.getMarkdown();
    expect(markdown).toContain(`](mention:${id})`);
    expect(markdown).toContain("[@Maya](agent:maya)");
    editor.commands.setContent(renderDocumentComment(markdown));
    expect(editor.getMarkdown()).toBe(markdown);
    expect(
      editor
        .getJSON()
        .content[0].content?.filter((node) => node.type === "mention"),
    ).toHaveLength(2);
  } finally {
    editor.destroy();
  }
});

test("document body preserves people and work references as distinct inline nodes through HTML saves", () => {
  const editor = new Editor({ extensions: createDocumentExtensions() });
  try {
    editor.commands.setContent({
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            { type: "mention", attrs: { id, label: "Joseph" } },
            { type: "text", text: " owns " },
            {
              type: "documentReference",
              attrs: {
                id,
                entityType: "story",
                label: "Refine onboarding",
                displayId: "PRO-98",
              },
            },
            { type: "text", text: " supporting " },
            {
              type: "documentReference",
              attrs: {
                id,
                entityType: "objective",
                label: "Improve activation",
              },
            },
          ],
        },
      ],
    });
    const original = editor.getJSON();
    const html = editor.getHTML();
    const saved = document.createElement("div");
    saved.innerHTML = html;
    const references = saved.querySelectorAll(
      '[data-type="document-reference"]',
    );
    expect([...references].map((node) => node.textContent)).toEqual([
      "Refine onboarding",
      "Improve activation",
    ]);
    expect(references[0].getAttribute("data-display-id")).toBe("PRO-98");
    editor.commands.setContent(html);
    expect(editor.getJSON()).toEqual(original);
    expect(editor.getText()).toContain(
      "@Joseph owns Refine onboarding supporting Improve activation",
    );
  } finally {
    editor.destroy();
  }
});
