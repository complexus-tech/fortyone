import { createHash } from "node:crypto";
import Redis from "ioredis";

export type SharedPreviewCache = {
  read: (key: string) => Promise<string | null>;
  write: (key: string, value: string, ttlMs: number) => Promise<void>;
  claim: (key: string, token: string, ttlMs: number) => Promise<boolean | null>;
  release: (key: string, token: string) => Promise<void>;
  consume?: (
    key: string,
    limit: number,
    ttlMs: number,
  ) => Promise<boolean | null>;
};

const PREFIX = "fortyone:previews:v1:";
export const previewCacheKey = (value: string) =>
  createHash("sha256").update(value).digest("hex");

let client: Redis | undefined;
let retryAfter = 0;
let warnedAt = 0;

const redisUrl = () => {
  if (process.env.PREVIEW_REDIS_URL) return process.env.PREVIEW_REDIS_URL;
  if (!process.env.APP_REDIS_HOST) return undefined;
  const url = new URL(
    `${process.env.APP_REDIS_DISABLE_TLS === "true" ? "redis" : "rediss"}://${process.env.APP_REDIS_HOST}`,
  );
  url.port = process.env.APP_REDIS_PORT || "6379";
  url.password = process.env.APP_REDIS_PASSWORD || "";
  url.pathname = `/${process.env.APP_REDIS_DB || "0"}`;
  return url.href;
};

const getClient = () => {
  if (Date.now() < retryAfter) return undefined;
  if (client) return client;
  const url = redisUrl();
  if (!url) return undefined;
  client = new Redis(url, {
    lazyConnect: true,
    connectTimeout: 750,
    commandTimeout: 1_000,
    maxRetriesPerRequest: 0,
    retryStrategy: () => null,
  });
  // Requests handle failures below; never log connection URLs or credentials.
  client.on("error", () => undefined);
  return client;
};

const attempt = async <T>(operation: (redis: Redis) => Promise<T>) => {
  let redis: Redis | undefined;
  try {
    redis = getClient();
    if (!redis) return undefined;
    return await operation(redis);
  } catch {
    redis?.disconnect();
    client = undefined;
    retryAfter = Date.now() + 30_000;
    if (Date.now() - warnedAt > 30_000) {
      warnedAt = Date.now();
      // eslint-disable-next-line no-console -- Operational fallback must be observable without logging URLs or credentials.
      console.warn(
        "Shared preview cache unavailable; using bounded local cache.",
      );
    }
    return undefined;
  }
};

export const sharedPreviewCache: SharedPreviewCache = {
  read: async (key) =>
    (await attempt((redis) => redis.get(PREFIX + key))) ?? null,
  write: async (key, value, ttlMs) => {
    await attempt((redis) => redis.set(PREFIX + key, value, "PX", ttlMs));
  },
  claim: async (key, token, ttlMs) => {
    const claimed = await attempt((redis) =>
      redis.set(PREFIX + key, token, "PX", ttlMs, "NX"),
    );
    return claimed === undefined ? null : claimed === "OK";
  },
  consume: async (key, limit, ttlMs) => {
    const count = await attempt((redis) =>
      redis.eval(
        "local n = redis.call('incr', KEYS[1]); if n == 1 then redis.call('pexpire', KEYS[1], ARGV[1]); end; return n",
        1,
        PREFIX + key,
        ttlMs,
      ),
    );
    return count === undefined
      ? null
      : typeof count === "number" && count <= limit;
  },
  release: async (key, token) => {
    await attempt((redis) =>
      redis.eval(
        "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end",
        1,
        PREFIX + key,
        token,
      ),
    );
  },
};
