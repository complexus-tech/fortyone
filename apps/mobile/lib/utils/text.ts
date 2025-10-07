export const toTitleCase = (text = "") => {
  if (!text) return "";
  if (text.length === 1) return text.toUpperCase();
  return text.charAt(0).toUpperCase() + text.slice(1);
};
