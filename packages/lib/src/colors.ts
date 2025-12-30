export const colors = [
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

/**
 * Determines if a background color is light or dark based on luminance
 * @param color - CSS color value (hex, rgb, rgba, hsl, etc.) - optional
 * @returns 'light' for light backgrounds, 'dark' for dark backgrounds
 */
export const getColorBrightness = (color?: string): "light" | "dark" => {
  if (!color) return "light"; // Default to light when no color provided

  const rgb = parseColorToRgb(color);
  if (!rgb) return "light";

  const luminance = (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) / 255;
  return luminance > 0.5 ? "light" : "dark";
};

/**
 * Parses CSS color formats to RGB values
 */
const parseColorToRgb = (
  color: string
): { r: number; g: number; b: number } | null => {
  if (color.startsWith("#")) {
    const hex = color.slice(1);
    if (hex.length === 3) {
      const r = parseInt(hex[0] + hex[0], 16);
      const g = parseInt(hex[1] + hex[1], 16);
      const b = parseInt(hex[2] + hex[2], 16);
      return { r, g, b };
    } else if (hex.length === 6) {
      const r = parseInt(hex.slice(0, 2), 16);
      const g = parseInt(hex.slice(2, 4), 16);
      const b = parseInt(hex.slice(4, 6), 16);
      return { r, g, b };
    }
  }

  if (color.startsWith("rgb")) {
    const match = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
    if (match) {
      return {
        r: parseInt(match[1]),
        g: parseInt(match[2]),
        b: parseInt(match[3]),
      };
    }
  }

  if (typeof window !== "undefined") {
    const temp = document.createElement("div");
    temp.style.color = color;
    document.body.appendChild(temp);
    const computed = window.getComputedStyle(temp).color;
    document.body.removeChild(temp);
    const match = computed.match(/rgb\((\d+),\s*(\d+),\s*(\d+)\)/);
    if (match) {
      return {
        r: parseInt(match[1]),
        g: parseInt(match[2]),
        b: parseInt(match[3]),
      };
    }
  }

  return null;
};
