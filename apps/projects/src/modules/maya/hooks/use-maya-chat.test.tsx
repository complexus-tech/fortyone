/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useChat } from "@ai-sdk/react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useAiChatMessages } from "@/modules/ai-chats/hooks/use-ai-chat-messages";
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
  useAiChatMessages: jest.fn(() => ({ data: [] })),
}));

jest.mock("@/modules/ai-chats/queries/get-ai-chat-messages", () => ({
  getAiChatMessages: jest.fn(),
}));

jest.mock("@/modules/ai-chats/constants", () => ({
  aiChatKeys: {
    lists: (workspaceSlug: string) => [
      "ai-chats",
      "workspace",
      workspaceSlug,
      "list",
    ],
    messages: (workspaceSlug: string, chatId: string) => [
      "ai-chats",
      "workspace",
      workspaceSlug,
      "detail",
      chatId,
      "messages",
    ],
    totalMessages: (workspaceSlug: string) => [
      "ai-chats",
      "workspace",
      workspaceSlug,
      "total-messages",
    ],
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
  useTotalMessages: () => ({ data: mockTotalMessages }),
}));

jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({
    displayTier: "Pro",
    getLimit: () => mockMessageLimit,
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

let mockChatStatus: "error" | "ready" | "streaming" | "submitted" = "ready";
let mockMessageLimit = Infinity;
let mockTotalMessages = 0;
const sendMessage = jest.fn(() => Promise.resolve());
const regenerate = jest.fn(() => Promise.resolve());
const mockedUseChat = jest.mocked(useChat);
const mockedUseAiChatMessages = jest.mocked(useAiChatMessages);
let onFinish: ChatFinishHandler | undefined;

const renderMayaChat = (
  onUserMessageSubmitted = jest.fn(),
  hasSelectedChat = false,
) => {
  const hook = renderHook(() =>
    useMayaChat({
      clearChatRef: jest.fn(),
      currentChatId: "chat-1",
      hasSelectedChat,
      onUserMessageSubmitted,
      updateChatRef: jest.fn(),
    }),
  );

  return { ...hook, onUserMessageSubmitted };
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
    mockChatStatus = "ready";
    mockMessageLimit = Infinity;
    mockTotalMessages = 0;
    sendMessage.mockResolvedValue(undefined);
    regenerate.mockResolvedValue(undefined);
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
        status: mockChatStatus,
        stop: jest.fn(),
      } as unknown as ReturnType<typeof useChat>;
    });
  });

  it.each([
    ["draft", false],
    ["selected", true],
  ] as const)(
    "loads history only for a %s persisted chat",
    (_, hasSelectedChat) => {
      renderMayaChat(jest.fn(), hasSelectedChat);

      expect(mockedUseAiChatMessages).toHaveBeenLastCalledWith("chat-1", {
        enabled: hasSelectedChat,
      });
    },
  );

  it("completes onboarding as soon as a valid user prompt is submitted", async () => {
    const { onUserMessageSubmitted, result } = renderMayaChat();

    await submitUserMessage(result.current.handleSuggestedPrompt);
    expect(onUserMessageSubmitted).toHaveBeenCalledTimes(1);

    await finishResponse();
    expect(onUserMessageSubmitted).toHaveBeenCalledTimes(1);
  });

  it.each([
    ["is aborted", { isAbort: true }],
    ["disconnects", { isDisconnect: true }],
    ["returns an error", { isError: true }],
  ])(
    "keeps onboarding complete when the submitted request %s",
    async (_, result) => {
      const { onUserMessageSubmitted, result: hook } = renderMayaChat();

      await submitUserMessage(hook.current.handleSuggestedPrompt);
      await finishResponse(result);

      expect(onUserMessageSubmitted).toHaveBeenCalledTimes(1);
    },
  );

  it("completes onboarding when credits are exhausted without sending an AI request", async () => {
    mockMessageLimit = 10;
    mockTotalMessages = 10;
    const { onUserMessageSubmitted, result } = renderMayaChat();

    act(() => {
      result.current.handleSuggestedPrompt("Help me plan my work");
    });

    await waitFor(() => {
      expect(onUserMessageSubmitted).toHaveBeenCalledTimes(1);
    });
    expect(sendMessage).not.toHaveBeenCalled();
  });

  it("does not complete onboarding for a blank prompt", () => {
    const { onUserMessageSubmitted, result } = renderMayaChat();

    act(() => {
      result.current.handleSuggestedPrompt("   ");
    });

    expect(onUserMessageSubmitted).not.toHaveBeenCalled();
    expect(sendMessage).not.toHaveBeenCalled();
  });

  it.each(["submitted", "streaming"] as const)(
    "does not complete onboarding while chat status is %s",
    (status) => {
      mockChatStatus = status;
      const { onUserMessageSubmitted, result } = renderMayaChat();

      act(() => {
        result.current.handleSuggestedPrompt("Help me plan my work");
      });

      expect(onUserMessageSubmitted).not.toHaveBeenCalled();
      expect(sendMessage).not.toHaveBeenCalled();
    },
  );

  it("does not count a second prompt while a send is already in progress", async () => {
    let resolveSend: (() => void) | undefined;
    sendMessage.mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          resolveSend = resolve;
        }),
    );
    const { onUserMessageSubmitted, result } = renderMayaChat();

    act(() => {
      result.current.handleSuggestedPrompt("Help me plan my work");
    });
    await waitFor(() => {
      expect(onUserMessageSubmitted).toHaveBeenCalledTimes(1);
      expect(sendMessage).toHaveBeenCalledTimes(1);
    });

    act(() => {
      result.current.handleSuggestedPrompt("Send another message");
    });

    expect(onUserMessageSubmitted).toHaveBeenCalledTimes(1);
    expect(sendMessage).toHaveBeenCalledTimes(1);

    await act(async () => {
      resolveSend?.();
    });
  });

  it("does not treat a regenerated response as a user-sent onboarding message", async () => {
    const { onUserMessageSubmitted, result } = renderMayaChat();

    await act(async () => {
      await result.current.regenerate();
    });
    expect(regenerate).toHaveBeenCalledTimes(1);

    await finishResponse();

    expect(onUserMessageSubmitted).not.toHaveBeenCalled();
  });
});
