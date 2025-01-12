const colors = [
  "#EA6060", // Red
  "#EAB308", // Amber
  "#06B6D4", // Cyan
  "#34D399", // Emerald
  "#FB7185", // Rose
  "#60A5FA", // Blue
  "#A78BFA", // Violet
  "#F472B6", // Pink
  "#4ADE80", // Green
  "#F59E0B", // Orange
  "#D946EF", // Fuchsia
  "#22D3EE", // Light Blue
  "#10B981", // Teal
  "#E879F9", // Purple
  "#FFB703", // Yellow
  "#118AB2", // Ocean Blue
  "#EF476F", // Coral
  "#073B4C", // Navy
  "#F4A261", // Peach
  "#264653", // Slate
];

export const generateRandomColor = ({
  exclude = [],
}: { exclude?: string[] } = {}) => {
  const finalColors = colors.filter((color) => !exclude.includes(color));
  return finalColors[Math.floor(Math.random() * finalColors.length)];
};
