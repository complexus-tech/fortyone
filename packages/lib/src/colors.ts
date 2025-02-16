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
