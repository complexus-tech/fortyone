import { previewCacheKey, sharedPreviewCache } from "./shared-cache";

const entries = new Map<string, { count: number; expiresAt: number }>();
const localBudget = (key: string, limit: number, ttlMs: number) => {
  const now = Date.now();
  const current = entries.get(key);
  const entry =
    current && current.expiresAt > now
      ? current
      : { count: 0, expiresAt: now + ttlMs };
  entry.count++;
  entries.delete(key);
  entries.set(key, entry);
  while (entries.size > 2000) entries.delete(entries.keys().next().value!);
  return entry.count <= limit;
};

export const allowPreviewRefresh = async (
  userId: string,
  normalizedUrl: string,
) => {
  const userKey = previewCacheKey(userId);
  const urlKey = previewCacheKey(`${userId}\n${normalizedUrl}`);
  const budget = await sharedPreviewCache.consume?.(
    `refresh-user:${userKey}`,
    20,
    60_000,
  );
  if (
    budget === false ||
    ((budget === null || budget === undefined) &&
      !localBudget(`user:${userKey}`, 20, 60_000))
  )
    return false;
  const perLink = await sharedPreviewCache.claim(
    `refresh-url:${urlKey}`,
    "refresh",
    5_000,
  );
  return perLink ?? localBudget(`url:${urlKey}`, 1, 5_000);
};
