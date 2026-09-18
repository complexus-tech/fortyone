import { readFileSync } from "node:fs";
import { Pool } from "pg";
import { Redis } from "@hocuspocus/extension-redis";
import { DocumentStore } from "./documents";
import { createCollaborationServer } from "./server";

const databaseURL = process.env.DATABASE_URL;
const origins = (process.env.COLLABORATION_ALLOWED_ORIGINS ?? "")
  .split(",")
  .map((origin) => origin.trim())
  .filter(Boolean);
if (!databaseURL || origins.length === 0)
  throw new Error(
    "DATABASE_URL and COLLABORATION_ALLOWED_ORIGINS are required",
  );
const production = process.env.NODE_ENV === "production";
const caPath = process.env.DATABASE_CA_FILE;
if (production && !caPath)
  throw new Error("DATABASE_CA_FILE is required in production");
if (
  caPath &&
  ["sslmode", "sslcert", "sslkey", "sslrootcert", "ssl"].some((key) =>
    new URL(databaseURL).searchParams.has(key),
  )
) {
  throw new Error(
    "Use DATABASE_CA_FILE without SSL parameters in DATABASE_URL",
  );
}
const pool = new Pool({
  connectionString: databaseURL,
  max: 10,
  connectionTimeoutMillis: 10_000,
  statement_timeout: 10_000,
  ssl: caPath
    ? { ca: readFileSync(caPath, "utf8"), rejectUnauthorized: true }
    : undefined,
});
pool.on("error", () => {
  console.error("Collaboration database connection failed");
});
const redisHost = process.env.COLLABORATION_REDIS_HOST;
const redis = redisHost
  ? new Redis({
      host: redisHost,
      port: Number(process.env.COLLABORATION_REDIS_PORT ?? 6379),
      prefix: "fortyone:documents",
      options: {
        password: process.env.COLLABORATION_REDIS_PASSWORD,
        ...(process.env.COLLABORATION_REDIS_TLS === "true" ? { tls: {} } : {}),
      },
    })
  : undefined;
const server = createCollaborationServer(new DocumentStore(pool), {
  port: Number(process.env.PORT ?? 1234),
  origins,
  redis,
});
await pool.query("SELECT 1");
await server.listen();
console.info("Document collaboration service is ready");
let stopping = false;
const shutdown = async () => {
  if (stopping) return;
  stopping = true;
  await server.destroy();
  await pool.end();
};
process.on("SIGTERM", () => void shutdown());
process.on("SIGINT", () => void shutdown());
