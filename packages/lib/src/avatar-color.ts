export type AvatarColor = {
  backgroundColor: string;
  foregroundColor: string;
};

export const AVATAR_COLORS = [
  { backgroundColor: "#E67E22", foregroundColor: "#111827" },
  { backgroundColor: "#C0392B", foregroundColor: "#FFFFFF" },
  { backgroundColor: "#8E44AD", foregroundColor: "#FFFFFF" },
  { backgroundColor: "#27AE60", foregroundColor: "#111827" },
  { backgroundColor: "#4A90E2", foregroundColor: "#111827" },
  { backgroundColor: "#30336B", foregroundColor: "#FFFFFF" },
] as const satisfies readonly AvatarColor[];

const normalizeAvatarSeed = (value: string) =>
  value.normalize("NFKC").trim().replace(/\s+/g, " ").toLowerCase() || "user";

export const getAvatarColor = (value: string): AvatarColor => {
  const seed = normalizeAvatarSeed(value);
  let hash = 0x811c9dc5;

  for (let index = 0; index < seed.length; index++) {
    hash ^= seed.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }

  return AVATAR_COLORS[(hash >>> 0) % AVATAR_COLORS.length] ?? AVATAR_COLORS[0];
};
