import { strict as assert } from "node:assert";
import { test } from "node:test";
import * as Y from "yjs";
import { applyAgentOperations, parseAgentOperations } from "./agent-operations";
import { renderDocument, restoreDocument } from "./documents";

const alice = "11111111-1111-4111-8111-111111111111";
const story = "22222222-2222-4222-8222-222222222222";

test("text edits preserve formatted content, media, human mentions and linked work", () => {
  const document = restoreDocument(
    null,
    "Plan",
    `<p>Make <strong>onboarding</strong> easy. Read <a href="https://example.com">the guide</a>.</p><p><span data-type="mention" data-mention-kind="user" data-id="${alice}" data-label="Alice">@Alice</span> owns <span data-type="document-reference" data-entity-type="story" data-id="${story}" data-label="Ship onboarding" data-display-id="PRO-98">PRO-98 Ship onboarding</span>.</p><img src="/api/media/image" data-attachment-id="image"><video data-document-media-video src="/api/media/video"></video>`,
  );
  try {
    assert.deepEqual(
      applyAgentOperations(document, {
        version: 1,
        title: "Onboarding",
        replacements: [
          { before: "Make onboarding easy.", after: "Make onboarding clear." },
        ],
        appendParagraphs: ["Review with the team."],
      }),
      ["title", "content"],
    );
    const rendered = renderDocument(document);
    assert.match(rendered.html, /<strong>onboarding<\/strong>/);
    assert.match(rendered.html, /https:\/\/example.com/);
    assert.match(rendered.html, /data-attachment-id="image"/);
    assert.match(rendered.html, /data-document-media-video/);
    assert.deepEqual(rendered.mentions, [alice]);
    assert.match(rendered.text, /@Alice owns Ship onboarding/);
    assert.match(rendered.html, /data-display-id="PRO-98"/);
    assert.match(rendered.html, />Ship onboarding<\/span>/);
    assert.doesNotMatch(rendered.text, /PRO-98/);
    assert.deepEqual(rendered.references, [{ entityType: "story", id: story }]);
    assert.match(rendered.html, /data-type="document-reference"/);
    assert.equal(rendered.title, "Onboarding");
    assert.match(rendered.text, /Review with the team/);
  } finally {
    document.destroy();
  }
});

test("ambiguous text and replacement of reference links are rejected", () => {
  const document = restoreDocument(
    null,
    "Plan",
    '<p>Repeated</p><p>Repeated</p><p>Read <a href="/docs/guide">the guide</a>.</p>',
  );
  try {
    assert.throws(
      () =>
        applyAgentOperations(document, {
          version: 1,
          replacements: [{ before: "Repeated", after: "New" }],
        }),
      /text_not_unique/,
    );
    assert.throws(
      () =>
        applyAgentOperations(document, {
          version: 1,
          replacements: [
            { before: "Read the guide.", after: "Read something else." },
          ],
        }),
      /protected_reference/,
    );
    assert.match(renderDocument(document).html, /href="\/docs\/guide"/);
  } finally {
    document.destroy();
  }
});

test("concurrent Yjs edits outside the replaced span merge in either order", () => {
  const initial = restoreDocument(
    null,
    "Plan",
    "<p>Original line.</p><p>Last line.</p>",
  );
  const agent = restoreDocument(Y.encodeStateAsUpdate(initial), "", "");
  const human = restoreDocument(Y.encodeStateAsUpdate(initial), "", "");
  try {
    applyAgentOperations(agent, {
      version: 1,
      replacements: [{ before: "Original", after: "Updated" }],
    });
    const last = human.getXmlFragment("default").get(1) as Y.XmlElement;
    (last.get(0) as Y.XmlText).insert(0, "Human: ");
    const agentUpdate = Y.encodeStateAsUpdate(agent);
    const humanUpdate = Y.encodeStateAsUpdate(human);
    Y.applyUpdate(agent, humanUpdate);
    Y.applyUpdate(human, agentUpdate);
    assert.equal(renderDocument(agent).html, renderDocument(human).html);
    assert.equal(
      renderDocument(agent).text,
      "Updated line.\nHuman: Last line.",
    );
  } finally {
    initial.destroy();
    agent.destroy();
    human.destroy();
  }
});

test("raw HTML is never accepted as an operation and appended text remains inert", () => {
  assert.throws(
    () => parseAgentOperations({ version: 1, html: "<p>Overwrite</p>" }),
    /invalid_operations/,
  );
  assert.throws(
    () => parseAgentOperations({ version: 1, title: "" }),
    /invalid_title/,
  );
  assert.throws(
    () =>
      parseAgentOperations({
        version: 1,
        appendParagraphs: ["two\nparagraphs"],
      }),
    /invalid_paragraphs/,
  );
  const document = restoreDocument(null, "Plan", "<p>Keep</p>");
  try {
    applyAgentOperations(document, {
      version: 1,
      appendParagraphs: ["<script>alert(1)</script>"],
    });
    assert.doesNotMatch(renderDocument(document).html, /<script>/);
    assert.match(renderDocument(document).html, /&lt;script&gt;/);
  } finally {
    document.destroy();
  }
});

test("emoji replacements retain whole Unicode characters at minimal edit boundaries", () => {
  for (const [before, after] of [
    ["😀", "😁"],
    ["😀", "\uD83E\uDE00"],
  ]) {
    const document = restoreDocument(
      null,
      "Plan",
      `<p>Ship ${before} today.</p>`,
    );
    try {
      applyAgentOperations(document, {
        version: 1,
        replacements: [
          { before: `Ship ${before} today.`, after: `Ship ${after} today.` },
        ],
      });
      assert.equal(renderDocument(document).text, `Ship ${after} today.`);
      const restored = restoreDocument(Y.encodeStateAsUpdate(document), "", "");
      try {
        assert.equal(renderDocument(restored).text, `Ship ${after} today.`);
      } finally {
        restored.destroy();
      }
    } finally {
      document.destroy();
    }
  }
});
