const avatarColors = [
  "#F97316",
  "#0EA5E9",
  "#22C55E",
  "#A855F7",
  "#EC4899",
  "#14B8A6",
  "#EAB308",
  "#EF4444",
];

export const getPublicAvatarColor = (value: string) => {
  const seed = value || "Feedback";
  let hash = 0;
  for (const character of seed) {
    hash = (hash * 31 + character.charCodeAt(0)) % avatarColors.length;
  }
  return avatarColors[hash] ?? avatarColors[0];
};
