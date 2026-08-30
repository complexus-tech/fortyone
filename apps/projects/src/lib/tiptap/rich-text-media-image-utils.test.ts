import {
  getRichTextImageStyle,
  getRichTextImageStyleString,
  normalizeRichTextImageAlignment,
  normalizeRichTextImageWidth,
  parseRichTextImageWidthToPixels,
} from "./rich-text-media-image-utils";

describe("rich text media image utilities", () => {
  it("normalizes untrusted persisted dimensions and alignment", () => {
    expect(normalizeRichTextImageWidth(" 320px ")).toBe("320px");
    expect(normalizeRichTextImageWidth("calc(100% - 10px)")).toBe("100%");
    expect(normalizeRichTextImageAlignment("right")).toBe("right");
    expect(normalizeRichTextImageAlignment("diagonal")).toBe("center");
  });

  it("produces a bounded, serializable image layout", () => {
    expect(getRichTextImageStyle("50%", "left")).toEqual({
      display: "block",
      marginLeft: "0",
      marginRight: "auto",
      maxWidth: "100%",
      width: "50%",
    });
    expect(getRichTextImageStyleString("320px", "right")).toBe(
      "display:block;margin-left:auto;margin-right:0;max-width:100%;width:320px",
    );
    expect(parseRichTextImageWidthToPixels("50%", 800)).toBe(400);
  });
});
