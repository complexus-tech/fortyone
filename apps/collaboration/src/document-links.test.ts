import { strict as assert } from "node:assert";
import { test } from "node:test";
import { extractDocumentLinks } from "./document-links";

const id = "11111111-1111-4111-8111-111111111111";
test("typed links are deduplicated and prose that resembles a mention is not an entity", () => {
  const links = extractDocumentLinks({
    type: "doc",
    content: [
      {
        type: "paragraph",
        content: [
          { type: "text", text: "@Alice is ordinary text" },
          { type: "mention", attrs: { id, label: "Alice" } },
          {
            type: "documentReference",
            attrs: { id, entityType: "story", label: "Task" },
          },
          {
            type: "documentReference",
            attrs: { id, entityType: "story", label: "Task again" },
          },
        ],
      },
    ],
  });
  assert.deepEqual(links, {
    references: [{ id, entityType: "story" }],
    mentions: [id],
  });
});
test("unknown entity types and malformed typed IDs cannot enter the reference index", () => {
  assert.throws(
    () => extractDocumentLinks({ type: "mention", attrs: { id: "../other" } }),
    /Invalid document reference/,
  );
  assert.throws(
    () =>
      extractDocumentLinks({
        type: "documentReference",
        attrs: { id, entityType: "workspace" },
      }),
    /Invalid document reference/,
  );
});
