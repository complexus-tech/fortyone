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
      `<img src=x onerror="actor()"> mentioned you: ${content}`,
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
        value: content,
      },
    ]);
    expect(result).not.toHaveProperty("html");
  });
});
