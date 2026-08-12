import "server-only";

import { createHmac } from "node:crypto";
import { headers } from "next/headers";

const PROOF_VERSION = "v1";
const HEADER_NAME_PATTERN = /^[a-z0-9-]+$/;
const PORTAL_SLUG_PATTERN = /^(?=.{3,255}$)[a-z0-9](?:[a-z0-9-]*[a-z0-9])$/;

const sign = (secret: string, value: string) =>
  createHmac("sha256", secret).update(value).digest("hex");

const getIngressSecret = () => {
  const configured = process.env.FEEDBACK_INGRESS_SECRET?.trim();
  if (configured && configured.length >= 32) return configured;
  throw new Error(
    "FEEDBACK_INGRESS_SECRET must contain at least 32 characters",
  );
};

const getFirstAddress = (value?: string | null) =>
  value?.split(",")[0]?.trim().toLowerCase() ?? "";

const getTrustedClientAddress = (
  requestHeaders: Awaited<ReturnType<typeof headers>>,
) => {
  if (process.env.VERCEL) {
    const address = getFirstAddress(
      requestHeaders.get("x-vercel-forwarded-for"),
    );
    if (address) return address;
    throw new Error("Vercel did not provide a feedback client address");
  }

  const configuredHeader =
    process.env.FEEDBACK_TRUSTED_CLIENT_IP_HEADER?.trim().toLowerCase();
  if (configuredHeader) {
    if (!HEADER_NAME_PATTERN.test(configuredHeader)) {
      throw new Error("FEEDBACK_TRUSTED_CLIENT_IP_HEADER is invalid");
    }
    const address = getFirstAddress(requestHeaders.get(configuredHeader));
    if (address) return address;
    throw new Error(
      `Trusted proxy header ${configuredHeader} did not contain a client address`,
    );
  }

  if (process.env.NODE_ENV !== "production") {
    return (
      getFirstAddress(requestHeaders.get("x-forwarded-for")) ||
      getFirstAddress(requestHeaders.get("x-real-ip")) ||
      "local-development"
    );
  }

  throw new Error(
    "Anonymous feedback requires Vercel or FEEDBACK_TRUSTED_CLIENT_IP_HEADER",
  );
};

export const createFeedbackIngressHeaders = async (portalSlug: string) => {
  const requestHeaders = await headers();
  const secret = getIngressSecret();
  const normalizedPortalSlug = normalizeFeedbackPortalSlug(portalSlug);
  const clientAddress = getTrustedClientAddress(requestHeaders);
  const clientFingerprint = sign(
    secret,
    `feedback-client-v1\n${clientAddress}`,
  );
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const signature = sign(
    secret,
    `${PROOF_VERSION}\n${normalizedPortalSlug}\n${clientFingerprint}\n${timestamp}`,
  );

  return {
    "X-FortyOne-Feedback-Identity": clientFingerprint,
    "X-FortyOne-Feedback-Signature": signature,
    "X-FortyOne-Feedback-Timestamp": timestamp,
    "X-FortyOne-Feedback-Version": PROOF_VERSION,
  };
};

export const normalizeFeedbackPortalSlug = (portalSlug: string) => {
  const normalized = portalSlug.trim().toLowerCase();
  if (!PORTAL_SLUG_PATTERN.test(normalized)) {
    throw new Error("Feedback portal slug is invalid");
  }
  return normalized;
};
