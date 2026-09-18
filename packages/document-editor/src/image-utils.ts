export type RichTextImageAlignment = "left" | "center" | "right";

export const DEFAULT_RICH_TEXT_IMAGE_WIDTH = "100%";
export const DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT: RichTextImageAlignment =
  "center";
export const MIN_RICH_TEXT_IMAGE_WIDTH_PX = 160;

export const normalizeRichTextImageAlignment = (
  value: unknown,
): RichTextImageAlignment =>
  value === "left" || value === "right"
    ? value
    : DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT;

export const normalizeRichTextImageWidth = (value: unknown) => {
  if (typeof value !== "string") return DEFAULT_RICH_TEXT_IMAGE_WIDTH;
  const width = value.trim();
  if (/^\d+(?:\.\d+)?(?:px|%)$/.test(width)) return width;
  return DEFAULT_RICH_TEXT_IMAGE_WIDTH;
};

export const getRichTextImageStyle = (width: unknown, alignment: unknown) => {
  const align = normalizeRichTextImageAlignment(alignment);
  return {
    display: "block",
    marginLeft: align === "left" ? "0" : "auto",
    marginRight: align === "right" ? "0" : "auto",
    maxWidth: "100%",
    width: normalizeRichTextImageWidth(width),
  };
};

export const getRichTextImageStyleString = (
  width: unknown,
  alignment: unknown,
) =>
  Object.entries(getRichTextImageStyle(width, alignment))
    .map(([property, value]) => {
      const cssProperty = property.replace(
        /[A-Z]/g,
        (match) => `-${match.toLowerCase()}`,
      );
      return `${cssProperty}:${value}`;
    })
    .join(";");

export const parseRichTextImageWidthToPixels = (
  width: unknown,
  maxWidth: number,
) => {
  const normalized = normalizeRichTextImageWidth(width);
  if (normalized.endsWith("%")) {
    return (Number.parseFloat(normalized) / 100) * maxWidth;
  }
  return Number.parseFloat(normalized);
};

export const clampNumber = (value: number, minimum: number, maximum: number) =>
  Math.min(Math.max(value, minimum), maximum);
