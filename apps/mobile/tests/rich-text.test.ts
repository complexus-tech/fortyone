import assert from "node:assert/strict";
import { before, test } from "node:test";
import { createRequire } from "node:module";
import { JSDOM } from "jsdom";
import type { Editor as EditorType } from "@tiptap/core";
import {
  getDescriptionHtml,
  plainTextToHtml,
  isSafeLink,
} from "../components/rich-text/content";

let Editor: typeof EditorType;
let createMobileRichTextExtensions: typeof import("../components/rich-text/extensions").createMobileRichTextExtensions;
let getRichTextValue: typeof import("../components/rich-text/extensions").getRichTextValue;
let sanitizeRichText: typeof import("../components/rich-text/sanitize").sanitizeRichText;
let getUnsupportedRichText: typeof import("../components/rich-text/sanitize").getUnsupportedRichText;

before(async () => {
  const dom = new JSDOM("<!DOCTYPE html><html><body></body></html>", {
    url: "https://editor.fortyone.test",
  });
  Object.assign(globalThis, {
    window: dom.window,
    document: dom.window.document,
    DOMParser: dom.window.DOMParser,
    Node: dom.window.Node,
    HTMLElement: dom.window.HTMLElement,
    Element: dom.window.Element,
    MutationObserver: dom.window.MutationObserver,
    getComputedStyle: dom.window.getComputedStyle.bind(dom.window),
    requestAnimationFrame: (callback: FrameRequestCallback) =>
      setTimeout(callback, 0),
    cancelAnimationFrame: clearTimeout,
  });
  Object.defineProperty(globalThis, "navigator", {
    value: dom.window.navigator,
    configurable: true,
  });
  // Keep ProseMirror in one module system. Mixing an ESM Editor import with
  // tsx's CommonJS extension imports creates two independent plugin registries.
  const require = createRequire(__filename);
  ({ Editor } = require("@tiptap/core"));
  ({
    createMobileRichTextExtensions,
    getRichTextValue,
  } = require("../components/rich-text/extensions"));
  ({
    sanitizeRichText,
    getUnsupportedRichText,
  } = require("../components/rich-text/sanitize"));
});

test("plain legacy text is escaped while line breaks remain real HTML breaks", () => {
  assert.equal(
    plainTextToHtml("First <step>\nNext & final\n\n**literal**"),
    "<p>First &lt;step&gt;<br>Next &amp; final</p><p>**literal**</p>",
  );
  assert.equal(
    getDescriptionHtml("<p>Rich <strong>text</strong></p>", "stale plain text"),
    "<p>Rich <strong>text</strong></p>",
  );
  assert.equal(
    getDescriptionHtml(null, "<script>alert(1)</script>"),
    "<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>",
  );
});

test("web documents retain nested checklists, tables, mentions and attachment attributes through editing", () => {
  const webHtml =
    '<h2>Acceptance criteria</h2><p><strong>Keep</strong> <em>this</em> <u>format</u> <s>old</s> <a href="https://fortyone.app">link</a></p>' +
    '<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><label><input type="checkbox" checked><span></span></label><div><p>Parent</p><ul data-type="taskList"><li data-type="taskItem" data-checked="false"><label><input type="checkbox"><span></span></label><div><p>Nested task</p></div></li></ul></div></li></ul>' +
    '<table><tbody><tr><th colspan="2" rowspan="1" colwidth="120,140"><p>Heading</p></th></tr><tr><td><p>A</p></td><td><p>B</p></td></tr></tbody></table>' +
    '<p><a class="mention" href="/profile/1234">@Joseph</a></p>' +
    '<img src="https://cdn.fortyone.app/image.jpg" data-width="62%" data-align="right" data-attachment-id="asset-image" alt="Reference">' +
    '<video src="https://cdn.fortyone.app/clip.mp4" data-document-media-video="true" data-attachment-id="asset-video" controls></video><p>End</p>';
  assert.deepEqual(getUnsupportedRichText(webHtml), []);
  const editor = new Editor({
    extensions: createMobileRichTextExtensions(),
    content: sanitizeRichText(webHtml),
  });
  editor.commands.insertContentAt(
    editor.state.doc.content.size - 1,
    " updated",
  );
  const value = getRichTextValue(editor);
  const output = new DOMParser().parseFromString(value.html, "text/html");
  assert.equal(output.querySelector("h2")?.textContent, "Acceptance criteria");
  assert.ok(
    output.querySelector('li[data-type="taskItem"] li[data-type="taskItem"]'),
  );
  assert.equal(output.querySelector("th")?.getAttribute("colspan"), "2");
  assert.equal(output.querySelector("th")?.getAttribute("colwidth"), "120,140");
  assert.equal(output.querySelector("img")?.getAttribute("data-width"), "62%");
  assert.equal(
    output.querySelector("img")?.getAttribute("data-align"),
    "right",
  );
  assert.equal(
    output.querySelector("img")?.getAttribute("data-attachment-id"),
    "asset-image",
  );
  assert.equal(
    output.querySelector("video")?.getAttribute("data-attachment-id"),
    "asset-video",
  );
  assert.equal(
    output.querySelector('a[href="/profile/1234"]')?.textContent,
    "@Joseph",
  );
  assert.deepEqual(value.mentions, ["1234"]);
  assert.match(value.text, /Nested task/);
  assert.match(value.text, /End updated/);
  const second = new Editor({
    extensions: createMobileRichTextExtensions(),
    content: value.html,
  });
  assert.deepEqual(second.getJSON(), editor.getJSON());
  second.destroy();
  editor.destroy();
});

test("unknown blocks and custom formatting are detected before an editable schema can strip them", () => {
  assert.deepEqual(
    getUnsupportedRichText(
      "<p>Supported</p><section><p>Custom section</p></section>",
    ),
    ["section"],
  );
  assert.deepEqual(
    getUnsupportedRichText('<p><span data-type="equation">x²</span></p>'),
    ["equation"],
  );
  assert.deepEqual(
    getUnsupportedRichText('<p style="color:red">Colored text</p>'),
    ["custom text styling"],
  );
});

test("pasted and displayed HTML cannot retain scripts, event handlers, unsafe URLs or injected CSS", () => {
  const html = sanitizeRichText(
    '<script>alert(1)</script><p onclick="alert(1)" style="background:url(javascript:alert(1))">Hello</p><img src="file:///private/secret" onerror="alert(1)"><a href="javascript:alert(1)">Bad</a><video src="data:video/mp4;base64,AAAA" autoplay></video><a href="https://example.com">Good</a>',
  );
  assert.doesNotMatch(
    html,
    /script|onerror|onclick|style=|file:|data:|autoplay/,
  );
  const document = new DOMParser().parseFromString(html, "text/html");
  assert.equal(document.querySelector("img")?.getAttribute("src"), null);
  assert.equal(document.querySelector("a")?.getAttribute("href"), null);
  assert.equal(
    document
      .querySelector('a[href="https://example.com"]')
      ?.getAttribute("rel"),
    "noopener noreferrer",
  );
  assert.equal(isSafeLink("javascript:alert(1)"), false);
  assert.equal(isSafeLink("https://fortyone.app"), true);
});
