/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ModelMessage } from "ai";
import { normalizeInlineFileData } from "./normalize-file-data";

describe("normalizeInlineFileData", () => {
  it("converts inline image data URLs to raw base64 model content", () => {
    const messages: ModelMessage[] = [
      {
        role: "user",
        content: [
          { type: "text", text: "What is in this image?" },
          {
            type: "file",
            data: "data:image/png;base64,aGVsbG8=",
            filename: "Group 691314619.png",
            mediaType: "image/png",
          },
        ],
      },
    ];

    expect(normalizeInlineFileData(messages)).toEqual([
      {
        role: "user",
        content: [
          { type: "text", text: "What is in this image?" },
          {
            type: "file",
            data: "aGVsbG8=",
            filename: "Group 691314619.png",
            mediaType: "image/png",
          },
        ],
      },
    ]);
  });

  it("leaves hosted attachment URLs unchanged", () => {
    const messages: ModelMessage[] = [
      {
        role: "user",
        content: [
          {
            type: "file",
            data: new URL("https://example.com/image.png"),
            filename: "image.png",
            mediaType: "image/png",
          },
        ],
      },
    ];

    expect(normalizeInlineFileData(messages)).toEqual(messages);
  });
});
