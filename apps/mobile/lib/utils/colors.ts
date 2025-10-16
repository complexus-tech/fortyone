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

const colors = [
  "#FFE066",
  "#FF6B6B",
  "#C0392B",
  "#FFA07A",
  "#FFB6C1",
  "#E056FD",
  "#686DE0",
  "#E67E22",
  "#A8E6CF",
  "#9B59B6",
  "#8E44AD",
  "#6BCB77",
  "#4ECDC4",
  "#4A90E2",
  "#95A5A6",
  "#27AE60",
  "#2ECC71",
  "#30336B",
  "#B4A6AB",
  "#636E72",
  "#34495E",
  "#2C3E50",
];

export const generateRandomColor = ({
  exclude = [],
}: { exclude?: string[] } = {}) => {
  const finalColors = colors.filter((color) => !exclude.includes(color));
  return finalColors[Math.floor(Math.random() * finalColors.length)];
};
