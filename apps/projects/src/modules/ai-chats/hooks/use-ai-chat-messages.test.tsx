/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useQuery } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";
import { useSession } from "@/lib/auth/client";
import { useAiChatMessages } from "./use-ai-chat-messages";

jest.mock("@tanstack/react-query", () => ({ useQuery: jest.fn() }));

jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("../queries/get-ai-chat-messages", () => ({
  getAiChatMessages: jest.fn(),
}));

const useQueryMock = jest.mocked(useQuery);
const useSessionMock = jest.mocked(useSession);

describe("useAiChatMessages", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useSessionMock.mockReturnValue({
      data: { user: { id: "user-1" } },
    } as ReturnType<typeof useSession>);
    useQueryMock.mockReturnValue({} as ReturnType<typeof useQuery>);
  });

  it("does not query a client-only draft chat", () => {
    renderHook(() => useAiChatMessages("draft-chat-id-01", { enabled: false }));

    expect(useQueryMock).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
        queryKey: [
          "ai-chats",
          "workspace",
          "acme",
          "detail",
          "draft-chat-id-01",
          "messages",
        ],
      }),
    );
  });

  it("queries a selected persisted chat", () => {
    renderHook(() => useAiChatMessages("saved-chat-id-01", { enabled: true }));

    expect(useQueryMock).toHaveBeenCalledWith(
      expect.objectContaining({ enabled: true }),
    );
  });
});
