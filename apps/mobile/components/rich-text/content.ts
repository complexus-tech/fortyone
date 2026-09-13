export type RichTextValue = {
  html: string;
  text: string;
  mentions: string[];
};

export const EMPTY_RICH_TEXT: RichTextValue = {
  html: "",
  text: "",
  mentions: [],
};

export const isRichTextValue = (value: unknown): value is RichTextValue => {
  if (!value || typeof value !== "object") return false;
  const record = value as Record<string, unknown>;
  return (
    typeof record.html === "string" &&
    typeof record.text === "string" &&
    Array.isArray(record.mentions) &&
    record.mentions.every((id) => typeof id === "string")
  );
};

export const escapeHtml = (value: string) =>
  value.replace(/[&<>"']/g, (character) => {
    const entities: Record<string, string> = {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;",
    };
    return entities[character];
  });

// Legacy descriptions are plain text, not Markdown or trusted HTML.
export const plainTextToHtml = (text: string) =>
  text.trim()
    ? text
        .split(/\n{2,}/)
        .map((block) => `<p>${escapeHtml(block).replace(/\n/g, "<br>")}</p>`)
        .join("")
    : "";

export const getDescriptionHtml = (
  html?: string | null,
  text?: string | null,
) => (html?.trim() ? html : plainTextToHtml(text ?? ""));

export const isSafeLink = (value: string) => {
  try {
    const url = new URL(value);
    return ["https:", "http:", "mailto:"].includes(url.protocol);
  } catch {
    return false;
  }
};

export const isSafeMediaUrl = (value: string) => {
  try {
    return new URL(value).protocol === "https:";
  } catch {
    return false;
  }
};
