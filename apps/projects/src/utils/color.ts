/**
 * Converts a six-digit hexadecimal color into an rgba() value.
 */
export const hexToRgba = (hex = "#6B665C", opacity = 0.1): string => {
  const cleanHex = hex.replace("#", "");

  if (!/^[0-9A-F]{6}$/i.test(cleanHex)) {
    throw new Error("Invalid hex color format");
  }

  if (opacity < 0 || opacity > 1) {
    throw new Error("Opacity must be between 0 and 1");
  }

  const r = parseInt(cleanHex.substring(0, 2), 16);
  const g = parseInt(cleanHex.substring(2, 4), 16);
  const b = parseInt(cleanHex.substring(4, 6), 16);

  return `rgba(${r}, ${g}, ${b}, ${opacity})`;
};
