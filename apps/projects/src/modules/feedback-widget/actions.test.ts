/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { post } from "api-client";
import { auth } from "@/auth";
import { createFeedbackIngressHeaders } from "@/modules/public-portal/feedback-ingress";
import {
  createWidgetFeedbackAction,
  exchangeWidgetIdentityAction,
  markWidgetFeedbackUpdatesSeenAction,
  revokeWidgetIdentityAction,
} from "./actions";

jest.mock("next/cache", () => ({ revalidatePath: jest.fn() }));
jest.mock("api-client", () => ({ post: jest.fn() }));
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/utils", () => ({
  getApiError: (error: Error) => ({
    data: null,
    error: { message: error.message },
  }),
}));
jest.mock("@/modules/public-portal/data", () => ({
  toPublicRequest: (item: unknown) => item,
}));
jest.mock("@/modules/public-portal/feedback-ingress", () => ({
  createFeedbackIngressHeaders: jest.fn(async () => ({
    "X-FortyOne-Feedback-Signature": "signed",
  })),
  normalizeFeedbackPortalSlug: (slug: string) => slug.trim().toLowerCase(),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);
const ingressHeadersMock = jest.mocked(createFeedbackIngressHeaders);

const input = {
  boardId: "road-repairs",
  description: "The crossing is unsafe.",
  portalSlug: "city-roads",
  title: "Repair the crossing signal",
};

describe("feedback widget participation intent", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("does not create an anonymous item when an identified session expires", async () => {
    authMock.mockResolvedValue(null);

    const result = await createWidgetFeedbackAction({
      ...input,
      participationIntent: "account",
    });

    expect(result.error?.message).toContain("session expired");
    expect(postMock).not.toHaveBeenCalled();
    expect(ingressHeadersMock).not.toHaveBeenCalled();
  });

  it("keeps an anonymous widget draft anonymous when a session appears", async () => {
    authMock.mockResolvedValue({} as Awaited<ReturnType<typeof auth>>);
    postMock.mockResolvedValue({
      data: {
        anonymous: true,
        id: "feedback-1",
        slug: "repair-the-crossing-signal",
      },
    });

    const result = await createWidgetFeedbackAction({
      ...input,
      participationIntent: "anonymous",
    });

    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/widget/feedback/items",
      expect.objectContaining({ participationIntent: "anonymous" }),
      {
        credentials: "omit",
        headers: { "X-FortyOne-Feedback-Signature": "signed" },
      },
    );
    expect(ingressHeadersMock).toHaveBeenCalledWith("city-roads");
    expect(result.data?.participantKind).toBe("anonymous");
  });

  it("uses an iframe-only contributor session for identified widget users", async () => {
    postMock.mockResolvedValue({
      data: {
        anonymous: false,
        id: "feedback-1",
        participantKind: "external",
        slug: "repair-the-crossing-signal",
      },
    });

    await createWidgetFeedbackAction({
      ...input,
      participationIntent: "external",
      sessionToken: "iframe-session",
    });

    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/widget/feedback/items",
      expect.objectContaining({ participationIntent: "external" }),
      {
        credentials: "omit",
        headers: { Authorization: "FeedbackSession iframe-session" },
      },
    );
    expect(authMock).not.toHaveBeenCalled();
  });

  it("exchanges signed identity without putting it in credentials or a URL", async () => {
    postMock.mockResolvedValue({
      data: {
        participant: {
          displayName: "Amina",
          id: "contributor-1",
          kind: "external",
          masked: false,
          publicName: "Amina",
        },
        session: { expiresAt: "2026-08-13T00:00:00Z", token: "session" },
      },
    });

    const result = await exchangeWidgetIdentityAction({
      assertion: "payload.signature",
      parentOrigin: "https://app.example.com",
      portalSlug: "city-roads",
    });

    expect(postMock).toHaveBeenCalledWith(
      "portals/city-roads/feedback/widget/sessions",
      {
        assertion: "payload.signature",
        parentOrigin: "https://app.example.com",
      },
      {
        credentials: "omit",
        headers: { "X-FortyOne-Feedback-Signature": "signed" },
      },
    );
    expect(result.data?.participant.kind).toBe("external");
  });

  it("uses the iframe-only session to revoke identity and mark updates seen", async () => {
    postMock.mockResolvedValue({
      data: { lastSeenAt: "2026-08-13T00:00:00Z", unreadUpdateCount: 0 },
    });

    await markWidgetFeedbackUpdatesSeenAction({
      portalSlug: "city-roads",
      sessionToken: "iframe-session",
    });
    await revokeWidgetIdentityAction({
      portalSlug: "city-roads",
      sessionToken: "iframe-session",
    });

    const options = {
      credentials: "omit",
      headers: { Authorization: "FeedbackSession iframe-session" },
    };
    expect(postMock).toHaveBeenNthCalledWith(
      1,
      "portals/city-roads/feedback/updates/seen",
      {},
      options,
    );
    expect(postMock).toHaveBeenNthCalledWith(
      2,
      "portals/city-roads/feedback/sessions/revoke",
      {},
      options,
    );
  });
});
