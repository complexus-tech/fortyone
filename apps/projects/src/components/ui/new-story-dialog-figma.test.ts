/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Editor } from "@tiptap/react";
import { toast } from "sonner";
import {
  attachFigmaDesigns,
  insertFigmaDescription,
} from "./new-story-dialog-figma";

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

describe("new story dialog Figma description", () => {
  it("inserts only populated structured sections", () => {
    const run = jest.fn();
    const insertContent = jest.fn(() => ({ run }));
    const focus = jest.fn(() => ({ insertContent }));
    const chain = jest.fn(() => ({ focus }));
    const editor = { chain } as unknown as Editor;

    insertFigmaDescription(editor, {
      overview: "A concise overview",
      requirements: ["Support saved filters"],
      acceptanceCriteria: [],
      implementationNotes: ["Reuse the existing query owner"],
    });

    expect(insertContent).toHaveBeenCalledWith([
      {
        type: "heading",
        attrs: { level: 3 },
        content: [{ type: "text", text: "Overview" }],
      },
      {
        type: "paragraph",
        content: [{ type: "text", text: "A concise overview" }],
      },
      {
        type: "heading",
        attrs: { level: 3 },
        content: [{ type: "text", text: "Requirements" }],
      },
      {
        type: "bulletList",
        content: [
          {
            type: "listItem",
            content: [
              {
                type: "paragraph",
                content: [{ type: "text", text: "Support saved filters" }],
              },
            ],
          },
        ],
      },
      {
        type: "heading",
        attrs: { level: 3 },
        content: [{ type: "text", text: "Implementation notes" }],
      },
      {
        type: "bulletList",
        content: [
          {
            type: "listItem",
            content: [
              {
                type: "paragraph",
                content: [
                  { type: "text", text: "Reuse the existing query owner" },
                ],
              },
            ],
          },
        ],
      },
    ]);
    expect(run).toHaveBeenCalledTimes(1);
  });

  it("does nothing until the editor is available", () => {
    expect(() => {
      insertFigmaDescription(null, {
        overview: "Overview",
        requirements: [],
        acceptanceCriteria: [],
        implementationNotes: [],
      });
    }).not.toThrow();
  });
});

describe("new story dialog Figma attachments", () => {
  it("links every selected design and reports completion", async () => {
    const linkDesign = jest.fn().mockResolvedValue(undefined);

    await attachFigmaDesigns({
      artifacts: [
        { canonicalUrl: "https://www.figma.com/design/first" },
        { canonicalUrl: "https://www.figma.com/design/second" },
      ],
      linkDesign,
      storyId: "story-1",
    });

    expect(linkDesign).toHaveBeenCalledTimes(2);
    expect(linkDesign).toHaveBeenNthCalledWith(1, {
      storyId: "story-1",
      url: "https://www.figma.com/design/first",
    });
    expect(linkDesign).toHaveBeenNthCalledWith(2, {
      storyId: "story-1",
      url: "https://www.figma.com/design/second",
    });
    expect(toast.success).toHaveBeenCalledWith("2 Figma designs attached");
  });

  it("keeps story creation committed when an attachment fails", async () => {
    const linkDesign = jest
      .fn()
      .mockResolvedValueOnce(undefined)
      .mockRejectedValueOnce(new Error("Figma unavailable"));

    await attachFigmaDesigns({
      artifacts: [
        { canonicalUrl: "https://www.figma.com/design/first" },
        { canonicalUrl: "https://www.figma.com/design/second" },
      ],
      linkDesign,
      storyId: "story-1",
    });

    expect(toast.error).toHaveBeenCalledWith(
      "1 of 2 Figma designs attached",
      expect.objectContaining({
        action: expect.objectContaining({ label: "Retry" }),
        description:
          "The story was created. Retry the remaining design attachments.",
      }),
    );
  });
});
