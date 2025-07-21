export type MentionData = {
  id: string;
  fullName: string;
};

/**
 * Extracts mentioned users from HTML content containing mention spans
 * @param htmlContent - The HTML content from the editor
 * @returns Array of mentioned user data
 */
export const extractMentionsFromHTML = (htmlContent: string): MentionData[] => {
  if (!htmlContent) return [];

  // Create a temporary DOM element to parse the HTML
  const tempDiv = document.createElement("div");
  tempDiv.innerHTML = htmlContent;

  // Find all mention spans
  const mentionElements = tempDiv.querySelectorAll('a[data-type="mention"]');

  const mentions: MentionData[] = [];

  mentionElements.forEach((element) => {
    const id = element.getAttribute("data-id");
    const fullName = element.getAttribute("data-label") ?? "";
    if (id) {
      if (!mentions.find((m) => m.id === id)) {
        mentions.push({ id, fullName });
      }
    }
  });

  return mentions;
};
