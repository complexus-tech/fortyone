import {
  GetObjectCommand,
  PutObjectCommand,
  S3Client,
} from "@aws-sdk/client-s3";
import { MAX_METADATA_IMAGE_RESPONSE_BYTES } from "../safe-fetch";

export type PreviewImage = { body: Buffer; contentType: string };
export type PreviewImageStore = {
  available?: () => boolean;
  read: (key: string) => Promise<PreviewImage | undefined>;
  write: (key: string, image: PreviewImage) => Promise<void>;
};
const PREFIX = "previews/external/v1/";
const IMAGE_TTL_MS = 7 * 24 * 60 * 60 * 1000;
const READ_TIMEOUT_MS = 1_500;
const WRITE_TIMEOUT_MS = 5_000;
type StoreOperation = "read" | "write";
let client: S3Client | undefined;
const retryAfter: Record<StoreOperation, number> = { read: 0, write: 0 };
const bucket = () =>
  process.env.PREVIEW_S3_BUCKET || process.env.APP_AWS_BUCKET;
const getClient = (operation: StoreOperation) => {
  if (!bucket() || Date.now() < retryAfter[operation]) return undefined;
  if (client) return client;
  const accessKeyId =
    process.env.PREVIEW_S3_ACCESS_KEY_ID || process.env.APP_AWS_ACCESS_KEY_ID;
  const secretAccessKey =
    process.env.PREVIEW_S3_SECRET_ACCESS_KEY ||
    process.env.APP_AWS_SECRET_ACCESS_KEY;
  client = new S3Client({
    region:
      process.env.PREVIEW_S3_REGION ||
      process.env.APP_AWS_REGION ||
      "us-east-1",
    endpoint:
      process.env.PREVIEW_S3_ENDPOINT ||
      process.env.APP_AWS_ENDPOINT ||
      undefined,
    forcePathStyle:
      (process.env.PREVIEW_S3_FORCE_PATH_STYLE ||
        process.env.APP_AWS_FORCE_PATH_STYLE) === "true",
    credentials:
      accessKeyId && secretAccessKey
        ? { accessKeyId, secretAccessKey }
        : undefined,
    maxAttempts: 1,
  });
  return client;
};
const unavailable = (operation: StoreOperation) => {
  retryAfter[operation] = Date.now() + 30_000;
  // eslint-disable-next-line no-console -- Operational fallback must be observable without logging URLs or credentials.
  console.warn("Preview image storage unavailable; using bounded local cache.");
};

export const previewImageStore: PreviewImageStore = {
  available: () => Boolean(bucket()) && Date.now() >= retryAfter.read,
  read: async (key) => {
    const s3 = getClient("read");
    if (!s3) return undefined;
    try {
      const response = await s3.send(
        new GetObjectCommand({ Bucket: bucket(), Key: PREFIX + key }),
        { abortSignal: AbortSignal.timeout(READ_TIMEOUT_MS) },
      );
      const createdAt = Number(response.Metadata?.createdat);
      if (
        !response.Body ||
        !response.ContentType ||
        !Number.isFinite(createdAt) ||
        Date.now() - createdAt > IMAGE_TTL_MS ||
        (response.ContentLength ?? Infinity) > MAX_METADATA_IMAGE_RESPONSE_BYTES
      )
        return undefined;
      const chunks: Buffer[] = [];
      let bytes = 0;
      for await (const chunk of response.Body as AsyncIterable<Uint8Array>) {
        bytes += chunk.byteLength;
        if (bytes > MAX_METADATA_IMAGE_RESPONSE_BYTES) return undefined;
        chunks.push(Buffer.from(chunk));
      }
      return {
        body: Buffer.concat(chunks, bytes),
        contentType: response.ContentType,
      };
    } catch (error) {
      if (
        !(
          error instanceof Error &&
          (error.name === "NoSuchKey" || error.name === "NotFound")
        )
      )
        unavailable("read");
      return undefined;
    }
  },
  write: async (key, image) => {
    // A short cache-read deadline can expire while a cold S3 connection starts.
    // Still publish the fetched image so the next process can reuse it.
    const s3 = getClient("write");
    if (!s3) return;
    try {
      await s3.send(
        new PutObjectCommand({
          Bucket: bucket(),
          Key: PREFIX + key,
          Body: image.body,
          ContentType: image.contentType,
          CacheControl: "private, max-age=86400",
          Metadata: { createdat: String(Date.now()) },
        }),
        { abortSignal: AbortSignal.timeout(WRITE_TIMEOUT_MS) },
      );
    } catch {
      unavailable("write");
    }
  },
};
