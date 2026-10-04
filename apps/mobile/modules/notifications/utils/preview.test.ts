import assert from "node:assert/strict";
import test from "node:test";
import { htmlToPlainText } from "lib/src/html-to-plain-text";
import { renderNotificationTemplate } from "lib/src/notification-template";

test("historical comment HTML yields the same readable native and web preview", () => {
  const url = "https://app.reversecontact.com/contact/find-phone-from-url";
  const message = renderNotificationTemplate({
    template: "{actor} left a comment: {content}",
    variables: {
      actor: { value: "hector", type: "actor" },
      content: {
        value: `<p><a href="${url}" target="_blank">${url}</a></p>`,
        type: "text",
      },
    },
  });
  assert.equal(message.text, `hector left a comment: ${url}`);
  assert.equal(message.segments[2].kind, "variable");
});

test("previews omit executable and hidden content, decode entities once, and separate blocks", () => {
  assert.equal(
    htmlToPlainText(
      '<script>alert(1)</script><style>bad</style><iframe>hidden</iframe><p>Review &amp; update</p><ul><li>First</li><li>&lt;example&gt; &#x1f44b; &amp;amp;</li></ul><img onerror="alert(2)">',
    ),
    "Review & update First <example> 👋 &amp;",
  );
  assert.equal(
    htmlToPlainText("Priority < 5 & delivery > 2"),
    "Priority < 5 & delivery > 2",
  );
  assert.equal(
    htmlToPlainText('<p title="a > b">Hello <strong>team'),
    "Hello team",
  );
});

test("plain text content preserves literal markup and does not become emphasized", () => {
  const result = renderNotificationTemplate({
    template: "{actor} mentioned you: {content}",
    variables: {
      actor: { value: "hector" },
      content: { value: "<example> &amp;", type: "plain_text" },
    },
  });
  assert.equal(result.text, "hector mentioned you: <example> &amp;");
  assert.deepEqual(result.segments[2], {
    kind: "variable",
    key: "content",
    value: "<example> &amp;",
    emphasized: false,
  });
});
