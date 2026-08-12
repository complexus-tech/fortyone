/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { post } from "api-client";
import { auth } from "@/auth";
import { createFeedbackIngressHeaders } from "@/modules/public-portal/feedback-ingress";
import { createWidgetFeedbackAction } from "./actions";

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
      isAnonymous: false,
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
      isAnonymous: true,
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
    expect(result.data?.isAnonymous).toBe(true);
  });
});
