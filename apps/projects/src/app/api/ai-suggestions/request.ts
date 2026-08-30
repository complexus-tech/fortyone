import type { z } from "zod";

export const MAX_SUGGESTION_REQUEST_BYTES = 1024;

const readBoundedBody = async (request: Request): Promise<string | null> => {
  const declaredLength = request.headers.get("content-length");
  if (declaredLength !== null) {
    if (!/^\d+$/.test(declaredLength)) return null;

    const length = Number(declaredLength);
    if (
      !Number.isSafeInteger(length) ||
      length < 0 ||
      length > MAX_SUGGESTION_REQUEST_BYTES
    ) {
      return null;
    }
  }

  const reader = request.body?.getReader();
  if (!reader) return null;

  const chunks: Uint8Array[] = [];
  let byteLength = 0;

  const readNextChunk = async (): Promise<boolean> => {
    const { done, value } = await reader.read();
    if (done) return true;

    byteLength += value.byteLength;
    if (byteLength > MAX_SUGGESTION_REQUEST_BYTES) {
      await reader.cancel();
      return false;
    }

    chunks.push(value);
    return readNextChunk();
  };

  try {
    if (!(await readNextChunk())) return null;
  } finally {
    reader.releaseLock();
  }

  const body = new Uint8Array(byteLength);
  let offset = 0;
  for (const chunk of chunks) {
    body.set(chunk, offset);
    offset += chunk.byteLength;
  }

  return new TextDecoder().decode(body);
};

export const parseSuggestionRequest = async <T>(
  request: Request,
  schema: z.ZodType<T>,
): Promise<T | null> => {
  let body: string | null;
  try {
    body = await readBoundedBody(request);
  } catch {
    return null;
  }
  if (body === null) return null;

  try {
    const parsed = schema.safeParse(JSON.parse(body));
    return parsed.success ? parsed.data : null;
  } catch {
    return null;
  }
};

export const truncateSuggestionContext = (value: unknown, limit: number) => {
  const text = typeof value === "string" ? value : "";
  return text.length <= limit ? text : `${text.slice(0, limit)}…`;
};
