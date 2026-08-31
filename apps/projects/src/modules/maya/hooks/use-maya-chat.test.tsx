/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useChat } from "@ai-sdk/react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useMayaChat } from "./use-maya-chat";

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/my-work",
  useRouter: () => ({ push: jest.fn() }),
}));

jest.mock("next-themes", () => ({
  useTheme: () => ({
    resolvedTheme: "dark",
    setTheme: jest.fn(),
    theme: "dark",
  }),
}));

jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

jest.mock("@ai-sdk/react", () => ({ useChat: jest.fn() }));

jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({
    fetchQuery: jest.fn(),
    invalidateQueries: jest.fn(),
  }),
}));

jest.mock("ai", () => ({
  DefaultChatTransport: class {
    body?: () => Record<string, unknown>;
  },
  generateId: () => "generated-chat-id",
  lastAssistantMessageIsCompleteWithApprovalResponses: jest.fn(),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { isInternal: false } } }),
}));

jest.mock("@/lib/hooks/subscriptions/subscription", () => ({
  useSubscription: () => ({ data: undefined }),
}));

jest.mock("@/lib/utils/files", () => ({ fileToBase64: jest.fn() }));

jest.mock("@/modules/ai-chats/hooks/use-ai-chat-messages", () => ({
  useAiChatMessages: () => ({ data: [] }),
}));

jest.mock("@/modules/ai-chats/queries/get-ai-chat-messages", () => ({
  getAiChatMessages: jest.fn(),
}));

jest.mock("@/modules/ai-chats/constants", () => ({
  aiChatKeys: {
    lists: () => ["ai-chats", "lists"],
    messages: (chatId: string) => ["ai-chats", "messages", chatId],
    totalMessages: () => ["ai-chats", "total-messages"],
  },
}));

jest.mock("@/lib/hooks/profile", () => ({
  useProfile: () => ({ data: undefined }),
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "work",
  }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
    workspaceSlug: "acme",
  }),
}));

jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: () => ({ workspace: { id: "workspace-1" } }),
}));

jest.mock("@/modules/ai-chats/hooks/use-memory", () => ({
  useMemories: () => ({ data: [] }),
}));

jest.mock("@/modules/ai-chats/hooks/use-total-messages", () => ({
  useTotalMessages: () => ({ data: 0 }),
}));

jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({
    displayTier: "Pro",
    getLimit: () => Infinity,
  }),
}));

jest.mock("../utils/tool-query-invalidation", () => ({
  getMayaToolInvalidationKeys: () => [],
}));

jest.mock("./use-maya-realtime-voice", () => ({
  useMayaRealtimeVoice: () => ({
    clearMessages: jest.fn(),
    disconnect: jest.fn(),
    messages: [],
  }),
}));

type ChatFinishHandler = (event: {
  isAbort?: boolean;
  isDisconnect?: boolean;
  isError?: boolean;
  message: { parts: unknown[] };
}) => void | Promise<void>;

const sendMessage = jest.fn(() => Promise.resolve());
const regenerate = jest.fn(() => Promise.resolve());
const mockedUseChat = jest.mocked(useChat);
let onFinish: ChatFinishHandler | undefined;

const renderMayaChat = (onUserMessageCompleted = jest.fn()) => {
  const hook = renderHook(() =>
    useMayaChat({
      clearChatRef: jest.fn(),
      currentChatId: "chat-1",
      onUserMessageCompleted,
      updateChatRef: jest.fn(),
    }),
  );

  return { ...hook, onUserMessageCompleted };
};

const submitUserMessage = async (
  handleSuggestedPrompt: (prompt: string) => void,
) => {
  act(() => {
    handleSuggestedPrompt("Help me plan my work");
  });

  await waitFor(() => {
    expect(sendMessage).toHaveBeenCalledWith({
      files: [],
      text: "Help me plan my work",
    });
  });
};

const finishResponse = async ({
  isAbort = false,
  isDisconnect = false,
  isError = false,
}: {
  isAbort?: boolean;
  isDisconnect?: boolean;
  isError?: boolean;
} = {}) => {
  const onFinishHandler = onFinish;
  if (!onFinishHandler) {
    throw new Error("Expected useChat to receive an onFinish callback.");
  }

  await act(async () => {
    await onFinishHandler({
      isAbort,
      isDisconnect,
      isError,
      message: { parts: [] },
    });
  });
};

describe("useMayaChat onboarding completion", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    onFinish = undefined;
    mockedUseChat.mockImplementation((options) => {
      const chatOptions = options as unknown as {
        onFinish?: ChatFinishHandler;
      };
      onFinish = chatOptions.onFinish;

      return {
        addToolApprovalResponse: jest.fn(),
        error: undefined,
        messages: [],
        regenerate,
        sendMessage,
        setMessages: jest.fn(),
        status: "ready",
        stop: jest.fn(),
      } as unknown as ReturnType<typeof useChat>;
    });
  });

  it("completes onboarding only after a user-sent message finishes successfully", async () => {
    const { onUserMessageCompleted, result } = renderMayaChat();

    await submitUserMessage(result.current.handleSuggestedPrompt);
    expect(onUserMessageCompleted).not.toHaveBeenCalled();

    await finishResponse();

    expect(onUserMessageCompleted).toHaveBeenCalledTimes(1);

    await finishResponse();
    expect(onUserMessageCompleted).toHaveBeenCalledTimes(1);
  });

  it.each([
    ["is aborted", { isAbort: true }],
    ["disconnects", { isDisconnect: true }],
    ["returns an error", { isError: true }],
  ])(
    "does not complete onboarding when a user message %s",
    async (_, result) => {
      const { onUserMessageCompleted, result: hook } = renderMayaChat();

      await submitUserMessage(hook.current.handleSuggestedPrompt);
      await finishResponse(result);
      await finishResponse();

      expect(onUserMessageCompleted).not.toHaveBeenCalled();
    },
  );

  it("does not treat a regenerated response as a user-sent onboarding message", async () => {
    const { onUserMessageCompleted, result } = renderMayaChat();

    await act(async () => {
      await result.current.regenerate();
    });
    expect(regenerate).toHaveBeenCalledTimes(1);

    await finishResponse();

    expect(onUserMessageCompleted).not.toHaveBeenCalled();
  });
});
