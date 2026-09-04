/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { revalidatePath } from "next/cache";
import { get, post } from "api-client";
import { auth } from "@/auth";
import {
  confirmFeedbackVerificationAction,
  createAnonymousFeedbackAction,
  createFeedbackAction,
  createVerifiedGuestFeedbackAction,
  exchangeFeedbackPreferenceTokenAction,
  getCurrentFeedbackGuestAction,
  requestFeedbackVerificationAction,
} from "./actions";
import { createFeedbackIngressHeaders } from "./feedback-ingress";
import {
  getFeedbackSessionAuthorization,
  setFeedbackPreferenceSessionToken,
  setFeedbackSessionToken,
} from "./guest-session";

jest.mock("next/cache", () => ({ revalidatePath: jest.fn() }));
jest.mock("api-client", () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  remove: jest.fn(),
}));
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/utils", () => ({
  getApiError: (error: Error) => ({
    data: null,
    error: { message: error.message },
  }),
}));
jest.mock("./data", () => ({ toPublicRequest: (item: unknown) => item }));
jest.mock("./feedback-ingress", () => ({
  createFeedbackIngressHeaders: jest.fn(async () => ({
    "X-FortyOne-Feedback-Signature": "signed",
  })),
  normalizeFeedbackPortalSlug: (slug: string) => slug.trim().toLowerCase(),
}));
jest.mock("./guest-session", () => ({
  clearFeedbackSessionToken: jest.fn(),
  getFeedbackPreferenceAuthorization: jest.fn(),
  getFeedbackSessionAuthorization: jest.fn(),
  setFeedbackPreferenceSessionToken: jest.fn(),
  setFeedbackSessionToken: jest.fn(),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);
const getMock = jest.mocked(get);
const ingressHeadersMock = jest.mocked(createFeedbackIngressHeaders);
const guestAuthorizationMock = jest.mocked(getFeedbackSessionAuthorization);
const setFeedbackSessionTokenMock = jest.mocked(setFeedbackSessionToken);
const setFeedbackPreferenceSessionTokenMock = jest.mocked(
  setFeedbackPreferenceSessionToken,
);

const input = {
  boardId: "road-repairs",
  description: "The crossing is unsafe.",
  portalSlug: "city-roads",
  title: "Repair the crossing signal",
};

describe("public feedback participation intent", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("does not downgrade an expired account submission to anonymous", async () => {
    authMock.mockResolvedValue(null);

    const result = await createFeedbackAction(input);

    expect(result).toEqual({
      data: null,
      error: { message: "Your session expired. Please log in and try again." },
    });
    expect(postMock).not.toHaveBeenCalled();
    expect(ingressHeadersMock).not.toHaveBeenCalled();
  });

  it("submits public feedback attachments as multipart form data", async () => {
    authMock.mockResolvedValue({} as Awaited<ReturnType<typeof auth>>);
    postMock.mockResolvedValue({
      data: {
        id: "feedback-with-file",
        participantKind: "account",
        slug: "repair-the-crossing-signal",
      },
    });
    const attachment = new File(["evidence"], "evidence.txt", {
      type: "text/plain",
    });
    const attachmentData = new FormData();
    attachmentData.append("files", attachment);

    await createFeedbackAction(
      { ...input, descriptionHTML: "<p>The crossing is unsafe.</p>" },
      attachmentData,
    );

    const payload = postMock.mock.calls[0]?.[1];
    expect(payload).toBeInstanceOf(FormData);
    expect((payload as FormData).get("descriptionHTML")).toBe(
      "<p>The crossing is unsafe.</p>",
    );
    expect((payload as FormData).getAll("files")).toEqual([attachment]);
  });

  it("preserves anonymous intent when a session appears before submit", async () => {
    authMock.mockResolvedValue({} as Awaited<ReturnType<typeof auth>>);
    postMock.mockResolvedValue({
      data: {
        anonymous: true,
        id: "feedback-1",
        slug: "repair-the-crossing-signal",
      },
    });

    const result = await createAnonymousFeedbackAction(input);

    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/feedback/items",
      expect.objectContaining({ participationIntent: "anonymous" }),
      {
        credentials: "omit",
        headers: { "X-FortyOne-Feedback-Signature": "signed" },
      },
    );
    expect(ingressHeadersMock).toHaveBeenCalledWith("city-roads");
    expect(result.data).toEqual(expect.objectContaining({ kind: "anonymous" }));
  });

  it("signs a genuinely logged-out anonymous submission", async () => {
    authMock.mockResolvedValue(null);
    postMock.mockResolvedValue({
      data: {
        anonymous: true,
        id: "feedback-2",
        slug: "repair-the-crossing-signal",
      },
    });

    await createAnonymousFeedbackAction(input);

    expect(ingressHeadersMock).toHaveBeenCalledWith("city-roads");
    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/feedback/items",
      expect.objectContaining({ participationIntent: "anonymous" }),
      {
        credentials: "omit",
        headers: { "X-FortyOne-Feedback-Signature": "signed" },
      },
    );
    expect(revalidatePath).toHaveBeenCalled();
  });

  it("requests guest verification with private identity choices and signed ingress", async () => {
    postMock.mockResolvedValue({
      data: {
        accepted: true,
        expiresAt: "2026-08-12T12:15:00.000Z",
      },
    });

    await requestFeedbackVerificationAction({
      displayName: "Ada Ndlovu",
      email: " ada@example.com ",
      hideNamePublicly: true,
      portalSlug: " City-Roads ",
    });

    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/feedback/verifications",
      {
        displayName: "Ada Ndlovu",
        email: "ada@example.com",
        hideNamePublicly: true,
        source: "portal",
      },
      {
        credentials: "omit",
        headers: { "X-FortyOne-Feedback-Signature": "signed" },
      },
    );
  });

  it("stores a confirmed guest token only in the server-side cookie boundary", async () => {
    postMock.mockResolvedValue({
      data: {
        participant: {
          avatarUrl: null,
          displayName: "Ada Ndlovu",
          email: "ada@example.com",
          id: "contributor-1",
          kind: "verified_guest",
          masked: true,
          publicName: "Anonymous",
        },
        session: {
          expiresAt: "2026-09-12T12:00:00.000Z",
          token: "opaque-session-token",
        },
        unreadUpdateCount: 3,
      },
    });

    const response = await confirmFeedbackVerificationAction({
      code: "123 456",
      email: "ada@example.com",
      portalSlug: "city-roads",
    });

    expect(setFeedbackSessionTokenMock).toHaveBeenCalledWith({
      expiresAt: "2026-09-12T12:00:00.000Z",
      portalSlug: "city-roads",
      token: "opaque-session-token",
    });
    expect(response.data).toEqual({
      participant: expect.objectContaining({
        kind: "verified_guest",
        masked: true,
        name: "Anonymous",
        unreadUpdateCount: 3,
      }),
    });
    expect(JSON.stringify(response)).not.toContain("opaque-session-token");
  });

  it("uses only the locked guest session for a verified guest submission", async () => {
    guestAuthorizationMock.mockResolvedValue(
      "FeedbackSession opaque-session-token",
    );
    postMock.mockResolvedValue({
      data: {
        following: true,
        id: "feedback-guest",
        participantKind: "verified_guest",
        slug: "repair-the-crossing-signal",
      },
    });

    const response = await createVerifiedGuestFeedbackAction(input);

    expect(authMock).not.toHaveBeenCalled();
    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/feedback/items",
      expect.objectContaining({ participationIntent: "verified_guest" }),
      {
        credentials: "omit",
        headers: {
          Authorization: "FeedbackSession opaque-session-token",
        },
      },
    );
    expect(response.data).toEqual(
      expect.objectContaining({ kind: "verified_guest", following: true }),
    );
  });

  it("recovers a magic-link guest session without exposing its token to the draft", async () => {
    guestAuthorizationMock.mockResolvedValue(
      "FeedbackSession opaque-session-token",
    );
    getMock.mockResolvedValue({
      data: {
        participant: {
          avatarUrl: null,
          displayName: "Ada Ndlovu",
          id: "contributor-1",
          kind: "verified_guest",
          masked: false,
          publicName: "Ada Ndlovu",
        },
        session: { expiresAt: "2026-09-12T12:00:00.000Z" },
        unreadUpdateCount: 2,
      },
    });

    const response = await getCurrentFeedbackGuestAction("city-roads");

    expect(getMock).toHaveBeenCalledWith(
      "portals/city-roads/feedback/session",
      {
        credentials: "omit",
        headers: {
          Authorization: "FeedbackSession opaque-session-token",
        },
      },
    );
    expect(response.data?.participant).toEqual(
      expect.objectContaining({
        kind: "verified_guest",
        unreadUpdateCount: 2,
      }),
    );
    expect(JSON.stringify(response)).not.toContain("opaque-session-token");
  });

  it("stores the nested preference session returned by token exchange", async () => {
    postMock.mockResolvedValue({
      data: {
        participant: {
          displayName: "Ada",
          id: "contributor-1",
          kind: "verified_guest",
          masked: false,
          publicName: "Ada",
        },
        session: {
          expiresAt: "2026-08-12T12:30:00.000Z",
          token: "preference-session",
        },
      },
    });

    const response = await exchangeFeedbackPreferenceTokenAction({
      portalSlug: "city-roads",
      token: "email-token",
    });

    expect(setFeedbackPreferenceSessionTokenMock).toHaveBeenCalledWith({
      expiresAt: "2026-08-12T12:30:00.000Z",
      portalSlug: "city-roads",
      token: "preference-session",
    });
    expect(response.data).toEqual({ exchanged: true });
  });
});
