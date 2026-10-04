/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { AppNotification } from "../types";
import { renderTemplate } from "./render-template";

describe("renderTemplate", () => {
  it("renders known values while preserving unknown placeholders", () => {
    const message: AppNotification["message"] = {
      template: "{actor} mentioned you in {story}.",
      variables: {
        actor: { value: "Ava" },
      },
    };

    expect(renderTemplate(message)).toEqual({
      segments: [
        {
          emphasized: true,
          key: "actor",
          kind: "variable",
          value: "Ava",
        },
        { kind: "text", value: " mentioned you in " },
        { kind: "text", value: "{story}" },
        { kind: "text", value: "." },
      ],
      text: "Ava mentioned you in {story}.",
    });
  });

  it("keeps untrusted values as data rather than synthesizing HTML", () => {
    const content =
      '<script>alert("xss")</script><img src=x onerror="alert(1)">&lt;entity&gt;';
    const message: AppNotification["message"] = {
      template: "{actor} mentioned you: {content}",
      variables: {
        actor: { value: '<img src=x onerror="actor()">' },
        content: { value: content, type: "text" },
      },
    };

    const result = renderTemplate(message);

    expect(result.text).toBe(
      '<img src=x onerror="actor()"> mentioned you: <entity>',
    );
    expect(result.segments).toEqual([
      {
        emphasized: true,
        key: "actor",
        kind: "variable",
        value: '<img src=x onerror="actor()">',
      },
      { kind: "text", value: " mentioned you: " },
      {
        emphasized: false,
        key: "content",
        kind: "variable",
        value: "<entity>",
      },
    ]);
    expect(result).not.toHaveProperty("html");
  });

  it("renders the screenshot's historical linked comment as a readable URL", () => {
    const url = "https://app.reversecontact.com/contact/find-phone-from-url";
    const result = renderTemplate({
      template: "{actor} left a comment: {content}",
      variables: {
        actor: { value: "hector", type: "actor" },
        content: {
          value: `<p><a target="_blank" rel="noopener noreferrer nofollow" href="${url}">${url}</a></p>`,
          type: "text",
        },
      },
    });
    expect(result.text).toBe(`hector left a comment: ${url}`);
    expect(result.text).not.toMatch(/target=|href=|<\/?[pa]>/);
  });

  it("preserves newly normalized literal text without decoding it again", () => {
    const result = renderTemplate({
      template: "{actor} replied: {content}",
      variables: {
        actor: { value: "hector" },
        content: { value: "<example> &amp;", type: "plain_text" },
      },
    });
    expect(result.text).toBe("hector replied: <example> &amp;");
    expect(result.segments[2]).toMatchObject({ emphasized: false });
  });

  it("does not expose a placeholder when an attachment-only comment has no text", () => {
    expect(
      renderTemplate({
        template: "{actor} left a comment: {content}",
        variables: {
          actor: { value: "hector" },
          content: { value: "", type: "plain_text" },
        },
      }).text,
    ).toBe("hector left a comment: ");
  });
});
