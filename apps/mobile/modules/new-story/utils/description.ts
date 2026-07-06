const escapeHtml = (value: string) =>
  value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");

const inlineMarkdownToHtml = (value: string) =>
  escapeHtml(value)
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(/_(.+?)_/g, "<em>$1</em>")
    .replace(/`(.+?)`/g, "<code>$1</code>");

export const descriptionToHtml = (value: string) => {
  const blocks = value
    .split(/\n{2,}/)
    .map((block) => block.trim())
    .filter(Boolean);

  if (blocks.length === 0) {
    return "";
  }

  return blocks
    .map((block) => {
      const lines = block.split("\n").map((line) => line.trim());
      const isBulleted = lines.every((line) => line.startsWith("- "));

      if (isBulleted) {
        const items = lines
          .map((line) => `<li>${inlineMarkdownToHtml(line.slice(2))}</li>`)
          .join("");
        return `<ul>${items}</ul>`;
      }

      return `<p>${inlineMarkdownToHtml(lines.join("<br />"))}</p>`;
    })
    .join("");
};
