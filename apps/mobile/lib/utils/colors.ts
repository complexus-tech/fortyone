/**
 * Converts hex color to rgba format with opacity
 * @param hex - Hex color (with or without #)
 * @param opacity - Opacity value between 0 and 1
 * @returns rgba color string
 */
export const hexToRgba = (
  hex: string = "#6B665C",
  opacity: number = 0.1
): string => {
  const cleanHex = hex.replace("#", "");

  if (!/^[0-9A-F]{6}$/i.test(cleanHex)) {
    // Return default gray if invalid
    return `rgba(107, 102, 92, ${opacity})`;
  }

  if (opacity < 0 || opacity > 1) {
    opacity = 0.1;
  }

  const r = parseInt(cleanHex.substring(0, 2), 16);
  const g = parseInt(cleanHex.substring(2, 4), 16);
  const b = parseInt(cleanHex.substring(4, 6), 16);

  return `rgba(${r}, ${g}, ${b}, ${opacity})`;
};
