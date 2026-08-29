/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { get, post, remove } from "@/lib/http";
import {
  createPersonalToken,
  createWebhookEndpoint,
  listPersonalTokens,
  rotateServiceAccountKey,
  rotateWebhookSecret,
} from "./api";
import type { CreateCredentialInput } from "./types";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  remove: jest.fn(),
}));

const mockGet = jest.mocked(get);
const mockPost = jest.mocked(post);
const mockRemove = jest.mocked(remove);
const ctx = { workspaceSlug: "acme" };

beforeEach(() => {
  jest.clearAllMocks();
  mockRemove.mockResolvedValue({ data: null });
});

describe("developer settings API contracts", () => {
  it("uses the session-authenticated personal-token management route", async () => {
    mockGet.mockResolvedValue({ data: [] });

    await expect(listPersonalTokens(ctx)).resolves.toEqual([]);
    expect(mockGet).toHaveBeenCalledWith("personal-access-tokens", ctx);
  });

  it("sends the complete least-privilege token input", async () => {
    const input: CreateCredentialInput = {
      name: "CLI",
      scopes: ["stories:read"],
      teamIds: ["team-1"],
      expiresAt: "2027-01-01T00:00:00.000Z",
    };
    const issued = { credential: { id: "credential-1" }, token: "show-once" };
    mockPost.mockResolvedValue({ data: issued });

    await expect(createPersonalToken(ctx, input)).resolves.toBe(issued);
    expect(mockPost).toHaveBeenCalledWith("personal-access-tokens", input, ctx);
  });

  it("uses browser management routes for webhook creation and rotation", async () => {
    const createInput = {
      name: "Production",
      url: "https://example.com/webhooks/fortyone",
      subscriptions: ["story.created" as const],
    };
    const created = { endpoint: { id: "endpoint-1" }, signingSecret: "whsec" };
    const rotated = {
      signingSecret: "whsec-next",
      generation: 2,
      previousSecretExpiresAt: "2027-01-01T00:00:00.000Z",
    };
    mockPost.mockResolvedValueOnce({ data: created }).mockResolvedValueOnce({
      data: rotated,
    });

    await expect(createWebhookEndpoint(ctx, createInput)).resolves.toBe(
      created,
    );
    await expect(rotateWebhookSecret(ctx, "endpoint-1")).resolves.toBe(rotated);

    expect(mockPost).toHaveBeenNthCalledWith(
      1,
      "webhook-endpoints",
      createInput,
      ctx,
    );
    expect(mockPost).toHaveBeenNthCalledWith(
      2,
      "webhook-endpoints/endpoint-1/rotate-secret",
      {},
      ctx,
    );
  });

  it("gives rotated service-account keys a bounded overlap", async () => {
    mockPost.mockResolvedValue({
      data: { credential: { id: "key-2" }, token: "show-once" },
    });

    await rotateServiceAccountKey(
      ctx,
      "account-1",
      "key-1",
      "2027-01-01T00:00:00.000Z",
    );

    expect(mockPost).toHaveBeenCalledWith(
      "service-accounts/account-1/keys/key-1/rotate",
      {
        expiresAt: "2027-01-01T00:00:00.000Z",
        overlapSeconds: 3600,
      },
      ctx,
    );
  });
});
