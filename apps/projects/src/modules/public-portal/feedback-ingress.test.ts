/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { headers } from "next/headers";
import {
  createFeedbackIngressHeaders,
  normalizeFeedbackPortalSlug,
} from "./feedback-ingress";

jest.mock("server-only", () => ({}));
jest.mock("next/headers", () => ({ headers: jest.fn() }));

const mockHeaders = jest.mocked(headers);
const originalSecret = process.env.FEEDBACK_INGRESS_SECRET;
const originalVercel = process.env.VERCEL;
const originalTrustedClientIpHeader =
  process.env.FEEDBACK_TRUSTED_CLIENT_IP_HEADER;

const restoreEnvironmentValue = (name: string, value?: string) => {
  if (value === undefined) {
    Reflect.deleteProperty(process.env, name);
    return;
  }
  process.env[name] = value;
};

describe("feedback ingress proof", () => {
  beforeEach(() => {
    process.env.FEEDBACK_INGRESS_SECRET =
      "test-feedback-ingress-secret-32-bytes";
    process.env.VERCEL = "1";
    mockHeaders.mockResolvedValue(
      new Headers({ "x-vercel-forwarded-for": "192.0.2.44" }),
    );
    jest.spyOn(Date, "now").mockReturnValue(1_800_000_000_000);
  });

  afterEach(() => {
    restoreEnvironmentValue("FEEDBACK_INGRESS_SECRET", originalSecret);
    restoreEnvironmentValue("VERCEL", originalVercel);
    restoreEnvironmentValue(
      "FEEDBACK_TRUSTED_CLIENT_IP_HEADER",
      originalTrustedClientIpHeader,
    );
    jest.restoreAllMocks();
  });

  it("binds the signed client fingerprint to the portal and time", async () => {
    const result = await createFeedbackIngressHeaders(" City-Roads ");

    expect(result).toEqual({
      "X-FortyOne-Feedback-Identity":
        "4de3dc762d2ec7ee7c2bdc0c9dccb06a6575d6baf14fd7780977fd7fb4e3c8d2",
      "X-FortyOne-Feedback-Signature":
        "e46dd5329fcf300bda60d49584a0637ae2107c78da79d57718901911d0acdb4b",
      "X-FortyOne-Feedback-Timestamp": "1800000000",
      "X-FortyOne-Feedback-Version": "v1",
    });
  });

  it("requires an explicitly trusted proxy outside Vercel", async () => {
    delete process.env.VERCEL;
    delete process.env.FEEDBACK_TRUSTED_CLIENT_IP_HEADER;
    const originalNodeEnv = process.env.NODE_ENV;
    Object.defineProperty(process.env, "NODE_ENV", {
      configurable: true,
      value: "production",
    });

    try {
      await expect(createFeedbackIngressHeaders("city-roads")).rejects.toThrow(
        "Anonymous feedback requires Vercel",
      );
    } finally {
      Object.defineProperty(process.env, "NODE_ENV", {
        configurable: true,
        value: originalNodeEnv,
      });
    }
  });

  it("rejects values that cannot be a workspace portal slug", () => {
    expect(normalizeFeedbackPortalSlug("city-roads")).toBe("city-roads");
    expect(() => normalizeFeedbackPortalSlug("../../users/verify")).toThrow(
      "Feedback portal slug is invalid",
    );
    expect(() => normalizeFeedbackPortalSlug("-city-roads")).toThrow(
      "Feedback portal slug is invalid",
    );
  });
});
