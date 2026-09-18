import { strict as assert } from "node:assert";
import { test } from "node:test";
import { getSchema } from "@tiptap/core";
import { generateJSON } from "@tiptap/html/server";
import { prosemirrorJSONToYDoc } from "@tiptap/y-tiptap";
import { createDocumentExtensions } from "@fortyone/document-editor";
import * as Y from "yjs";
import {
  cleanDocumentJSON,
  parseDocumentName,
  renderDocument,
} from "./documents";

test("document rooms are scoped to an exact UUID and restore epoch", () => {
  assert.equal(
    parseDocumentName("12345678-1234-1234-1234-123456789abc:2").epoch,
    "2",
  );
  for (const name of [
    "../private",
    "abc:1",
    "12345678-1234-1234-1234-123456789abc:0",
  ])
    assert.throws(() => parseDocumentName(name));
});

test("shared schema round trips formatting and media without upload placeholders", () => {
  const extensions = createDocumentExtensions({ collaborative: true });
  const json = generateJSON(
    '<h2>Plan</h2><p><strong>Ship</strong> together</p><img src="https://example.com/image.png" data-width="50%" data-align="left" data-attachment-id="attachment"><table><tbody><tr><th>A</th><td>B</td></tr></tbody></table><video data-document-media-video="true" src="https://example.com/video.mp4"></video>',
    extensions,
  );
  const document = prosemirrorJSONToYDoc(
    getSchema(extensions),
    json,
    "default",
  );
  document.getText("title").insert(0, "Plan");
  const rendered = renderDocument(document);
  assert.match(rendered.html, /<strong>Ship<\/strong>/);
  assert.match(rendered.html, /data-width="50%"/);
  assert.match(rendered.html, /data-attachment-id="attachment"/);
  assert.match(rendered.html, /<table/);
  assert.match(rendered.html, /data-document-media-video/);
  assert.equal(
    cleanDocumentJSON({ type: "image", attrs: { isUploading: true } }),
    null,
  );
  document.destroy();
});

test("independent editors converge when updates arrive in either order", () => {
  const extensions = createDocumentExtensions({ collaborative: true });
  const initial = prosemirrorJSONToYDoc(
    getSchema(extensions),
    generateJSON("<p>First</p><p>Last</p>", extensions),
    "default",
  );
  initial.getText("title").insert(0, "Plan");
  const first = new Y.Doc();
  const second = new Y.Doc();
  Y.applyUpdate(first, Y.encodeStateAsUpdate(initial));
  Y.applyUpdate(second, Y.encodeStateAsUpdate(initial));
  (first.getXmlFragment("default").get(0) as Y.XmlElement)
    .toArray()
    .forEach((node) => {
      if (node instanceof Y.XmlText) node.insert(0, "Alice: ");
    });
  (second.getXmlFragment("default").get(1) as Y.XmlElement)
    .toArray()
    .forEach((node) => {
      if (node instanceof Y.XmlText) node.insert(0, "Bob: ");
    });
  const left = Y.encodeStateAsUpdate(first);
  const right = Y.encodeStateAsUpdate(second);
  Y.applyUpdate(first, right);
  Y.applyUpdate(second, left);
  assert.equal(renderDocument(first).html, renderDocument(second).html);
  assert.match(renderDocument(first).text, /Alice: First/);
  assert.match(renderDocument(first).text, /Bob: Last/);
  initial.destroy();
  first.destroy();
  second.destroy();
});
