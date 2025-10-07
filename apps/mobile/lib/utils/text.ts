/**
 * Converts text to title case
 * @param text - The text to convert
 * @returns Text in title case
 */
export const toTitleCase = (text: string) => {
  if (!text) return "";
  if (text.length === 1) return text.toUpperCase();
  return text.charAt(0).toUpperCase() + text.slice(1);
};

/**
 * Truncates text to a specified length and adds ellipsis
 * @param text - The text to truncate
 * @param maxLength - Maximum number of characters to display
 * @returns Truncated text with ellipsis if needed
 */
export const truncateText = (text: string, maxLength: number) => {
  if (!text) return "";
  if (text.length <= maxLength) {
    return text;
  }

  return text.slice(0, maxLength) + "...";
};
