/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { revalidatePath } from "next/cache";
import { post } from "api-client";
import { auth } from "@/auth";
import { createAnonymousFeedbackAction, createFeedbackAction } from "./actions";
import { createFeedbackIngressHeaders } from "./feedback-ingress";

jest.mock("next/cache", () => ({ revalidatePath: jest.fn() }));
jest.mock("api-client", () => ({ post: jest.fn() }));
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

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);
const ingressHeadersMock = jest.mocked(createFeedbackIngressHeaders);

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
});
