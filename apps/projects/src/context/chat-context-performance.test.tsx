/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { act, render, screen } from "@testing-library/react";
import type { MayaChatConfig } from "@/modules/maya";
import { useMayaChat } from "@/modules/maya/hooks/use-maya-chat";
import { MayaChat } from "@/modules/maya/components";
import {
  ChatProvider,
  useChatActions,
  useChatRuntime,
  useChatVisibility,
} from "./chat-context";

const mockRouterReplace = jest.fn();
let mockPathname = "/acme/my-work";
let mockSearchParams = new URLSearchParams();
jest.mock("next/navigation", () => ({
  usePathname: () => mockPathname,
  useSearchParams: () => mockSearchParams,
  useRouter: () => ({ replace: mockRouterReplace }),
}));

jest.mock("ai", () => ({ generateId: () => "draft-chat-id" }));
jest.mock("@/modules/maya/hooks/use-maya-chat", () => ({
  useMayaChat: jest.fn(),
}));
jest.mock("@/components/walkthrough/walkthrough-provider", () => ({
  useWalkthrough: () => ({ completeWalkthroughAction: jest.fn() }),
}));
jest.mock("ui", () => ({
  Box: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  Button: ({ children }: { children?: ReactNode }) => (
    <button type="button">{children}</button>
  ),
  Text: ({ children }: { children?: ReactNode }) => <span>{children}</span>,
}));
jest.mock("@/components/ui", () => ({
  NewStoryDialog: () => null,
  NewObjectiveDialog: () => null,
}));
jest.mock("@/components/ui/new-sprint-dialog", () => ({
  NewSprintDialog: () => null,
}));
jest.mock("@/components/shared", () => ({
  BodyContainer: ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  ),
}));
jest.mock("@/components/ui/chat/chat-messages", () => ({
  ChatMessages: () => null,
}));
jest.mock("@/components/ui/chat/chat-input", () => ({ ChatInput: () => null }));
jest.mock("@/components/ui/chat/suggested-prompts", () => ({
  SuggestedPrompts: () => null,
}));
jest.mock("@/components/ui/chat/limit-reached", () => ({
  LimitReached: () => null,
}));
jest.mock("@/modules/maya/components/header", () => ({ Header: () => null }));
jest.mock("@/modules/maya/components/recent-work", () => ({
  RecentWork: () => null,
}));
jest.mock("@/lib/hooks/profile", () => ({
  useProfile: () => ({ data: { fullName: "Joseph" } }),
}));
jest.mock("@/shared/maya-usage/use-maya-message-availability", () => ({
  useMayaMessageAvailability: () => ({ isLimited: false }),
}));

type ChatRuntime = ReturnType<typeof useMayaChat>;
type ChatActions = ReturnType<typeof useChatActions>;

const useMayaChatMock = jest.mocked(useMayaChat);
const actionRender = jest.fn();
const visibilityRender = jest.fn();
const runtimeRender = jest.fn();
const runtimeMount = jest.fn();
const runtimeUnmount = jest.fn();
const submitPrompt = jest.fn();
const stopResponse = jest.fn();
const addFile = jest.fn();
const selectConversation = jest.fn();
const disconnectVoice = jest.fn();
let actions: ChatActions;
let runtime: ChatRuntime;
let streamToken: () => void;

const linkedFile = {
  id: "file-reference-1",
  mimeType: "application/vnd.google-apps.document",
  name: "Launch plan",
};

// A stateful chat double exercises React subscriptions and lifecycle without
// opening a network stream. Runtime callbacks deliberately change each render.
const useRuntimeFixture = (config: MayaChatConfig): ChatRuntime => {
  const [input, setInput] = useState("");
  const [tokenCount, setTokenCount] = useState(0);
  const [attachments, setAttachments] = useState<File[]>([]);
  const [googleDriveFiles, setGoogleDriveFiles] = useState<
    ChatRuntime["googleDriveFiles"]
  >([]);
  const selectionVersion = useRef(0);
  const [displayMessages, setDisplayMessages] = useState<
    ChatRuntime["displayMessages"]
  >([
    {
      id: "assistant-message-1",
      role: "assistant",
      parts: [
        {
          type: "dynamic-tool",
          toolName: "createStory",
          toolCallId: "tool-call-1",
          state: "approval-requested",
          input: { title: "Review launch plan" },
          approval: { id: "approval-1" },
        },
      ],
    },
  ]);

  useEffect(() => {
    runtimeMount();
    return () => {
      runtimeUnmount();
    };
  }, []);

  streamToken = () => {
    setTokenCount((count) => count + 1);
    setDisplayMessages((messages) =>
      messages.map((message) => ({
        ...message,
        parts: [...message.parts, { type: "text", text: "next token" }],
      })),
    );
  };

  return {
    currentChatId: config.currentChatId,
    input,
    setInput,
    attachments,
    setAttachments,
    googleDriveFiles,
    displayMessages,
    status: "streaming",
    realtimeVoice: {
      error: null,
      isSpeaking: true,
      disconnect: disconnectVoice,
    },
    handleStop: stopResponse,
    cancelPendingChatSelection: () => {
      selectionVersion.current += 1;
    },
    handleNewChat: () => {
      config.clearChatRef("next-draft-id");
      setInput("");
      setAttachments([]);
      setGoogleDriveFiles([]);
      setDisplayMessages([]);
    },
    handleChatSelect: async (chatId: string) => {
      const version = ++selectionVersion.current;
      await selectConversation(chatId);
      if (version !== selectionVersion.current) return;
      config.updateChatRef(chatId);
      setInput("");
      setAttachments([]);
      setGoogleDriveFiles([]);
    },
    handleSuggestedPrompt: (prompt: string) => {
      submitPrompt(prompt, input, tokenCount);
    },
    addGoogleDriveFile: (
      file: Parameters<ChatRuntime["addGoogleDriveFile"]>[0],
    ) => {
      addFile(file, input, tokenCount);
      setGoogleDriveFiles((files) => [...files, file]);
    },
    // Components not relevant to context subscriptions are stubbed above;
    // their unrelated runtime fields are intentionally omitted from this double.
  } as unknown as ChatRuntime;
};

const ActionsProbe = () => {
  actions = useChatActions();
  actionRender();
  return null;
};

const VisibilityProbe = () => {
  const isOpen = useChatVisibility();
  visibilityRender();
  return <span>{isOpen ? "Maya open" : "Maya closed"}</span>;
};

const RuntimeProbe = () => {
  runtime = useChatRuntime();
  runtimeRender();
  return <span data-testid="draft">{runtime.input}</span>;
};

const WorkspaceFixture = ({
  route = "my-work",
  workspace = "acme",
  children,
}: {
  route?: string;
  workspace?: string;
  children?: ReactNode;
}) => (
  <ChatProvider key={workspace}>
    <ActionsProbe />
    <VisibilityProbe />
    <RuntimeProbe />
    <main key={route}>{route === "maya" ? <MayaChat /> : route}</main>
    {children}
  </ChatProvider>
);

describe("Maya update isolation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockPathname = "/acme/my-work";
    mockSearchParams = new URLSearchParams();
    selectConversation.mockResolvedValue(undefined);
    useMayaChatMock.mockImplementation(useRuntimeFixture);
  });

  it("keeps action and visibility subscribers out of draft and stream updates", () => {
    render(<WorkspaceFixture />);
    const initialActions = actions;
    const initialActionRenders = actionRender.mock.calls.length;
    const initialVisibilityRenders = visibilityRender.mock.calls.length;
    const initialRuntimeRenders = runtimeRender.mock.calls.length;

    act(() => {
      runtime.setInput("Keep this draft while I navigate");
    });
    act(() => {
      streamToken();
    });

    expect(screen.getByTestId("draft")).toHaveTextContent(
      "Keep this draft while I navigate",
    );
    expect(runtimeRender.mock.calls.length).toBeGreaterThan(
      initialRuntimeRenders,
    );
    expect(actionRender).toHaveBeenCalledTimes(initialActionRenders);
    expect(visibilityRender).toHaveBeenCalledTimes(initialVisibilityRenders);
    expect(actions).toBe(initialActions);

    act(() => {
      actions.openChat();
    });
    expect(screen.getByText("Maya open")).toBeInTheDocument();
    expect(visibilityRender.mock.calls.length).toBeGreaterThan(
      initialVisibilityRenders,
    );
    expect(actionRender).toHaveBeenCalledTimes(initialActionRenders);
  });

  it("keeps one runtime, draft, files and approval state through page and popup changes", () => {
    mockPathname = "/acme/maya";
    const { rerender } = render(<WorkspaceFixture route="maya" />);
    const attachment = new File(["plan"], "plan.txt", {
      type: "text/plain",
    });
    const pendingApproval = runtime.displayMessages[0].parts[0];

    act(() => {
      actions.openChatWithDraft("Finish this after reviewing the board");
      runtime.setAttachments([attachment]);
    });
    act(() => {
      actions.openChatWithGoogleDriveFile(linkedFile);
    });
    mockPathname = "/acme/roadmap";
    rerender(<WorkspaceFixture route="roadmap" />);
    act(() => {
      actions.closeChat();
      streamToken();
    });
    mockPathname = "/acme/maya";
    rerender(<WorkspaceFixture route="maya" />);
    act(() => {
      actions.openChat();
    });

    expect(runtimeMount).toHaveBeenCalledTimes(1);
    expect(runtimeUnmount).not.toHaveBeenCalled();
    expect(stopResponse).not.toHaveBeenCalled();
    expect(disconnectVoice).not.toHaveBeenCalled();
    expect(selectConversation).not.toHaveBeenCalled();
    expect(runtime.currentChatId).toBe("draft-chat-id");
    expect(runtime.status).toBe("streaming");
    expect(runtime.input).toBe("Finish this after reviewing the board");
    expect(runtime.attachments).toEqual([attachment]);
    expect(runtime.googleDriveFiles).toEqual([
      {
        referenceId: linkedFile.id,
        mimeType: linkedFile.mimeType,
        name: linkedFile.name,
      },
    ]);
    expect(runtime.displayMessages[0].parts[0]).toBe(pendingApproval);
    expect(runtime.displayMessages[0].parts).toHaveLength(2);
    expect(submitPrompt).not.toHaveBeenCalled();
  });

  it("initializes the shared runtime from a conversation deep link", () => {
    mockPathname = "/acme/maya";
    mockSearchParams = new URLSearchParams("chatRef=saved-chat-1");
    render(<WorkspaceFixture route="maya" />);

    expect(runtimeMount).toHaveBeenCalledTimes(1);
    expect(runtime.currentChatId).toBe("saved-chat-1");
    expect(useMayaChatMock).toHaveBeenLastCalledWith(
      expect.objectContaining({
        hasSelectedChat: true,
        currentChatId: "saved-chat-1",
      }),
    );
    expect(selectConversation).not.toHaveBeenCalled();
  });

  it("selects a changed deep link once and preserves the current session on a bare return", async () => {
    const { rerender } = render(<WorkspaceFixture />);
    mockPathname = "/acme/maya";
    mockSearchParams = new URLSearchParams("chatRef=saved-chat-1");
    await act(async () => {
      rerender(<WorkspaceFixture route="maya" />);
    });
    expect(selectConversation).toHaveBeenCalledTimes(1);
    expect(selectConversation).toHaveBeenLastCalledWith("saved-chat-1");
    expect(runtime.currentChatId).toBe("saved-chat-1");

    act(() => {
      runtime.setInput("Keep this new draft");
      streamToken();
    });
    mockPathname = "/acme/roadmap";
    mockSearchParams = new URLSearchParams();
    rerender(<WorkspaceFixture route="roadmap" />);
    mockPathname = "/acme/maya";
    rerender(<WorkspaceFixture route="maya" />);
    expect(runtime.currentChatId).toBe("saved-chat-1");
    expect(runtime.input).toBe("Keep this new draft");
    expect(selectConversation).toHaveBeenCalledTimes(1);

    mockSearchParams = new URLSearchParams("chatRef=saved-chat-2");
    await act(async () => {
      rerender(<WorkspaceFixture route="maya" />);
    });
    expect(selectConversation).toHaveBeenCalledTimes(2);
    expect(selectConversation).toHaveBeenLastCalledWith("saved-chat-2");
    expect(runtime.currentChatId).toBe("saved-chat-2");
    expect(mockRouterReplace).not.toHaveBeenCalled();
    expect(runtimeMount).toHaveBeenCalledTimes(1);
  });

  it("does not restore the old URL conversation while a new-chat URL update is pending", () => {
    mockPathname = "/acme/maya";
    mockSearchParams = new URLSearchParams(
      "chatRef=saved-chat-1&panel=history",
    );
    const { rerender } = render(<WorkspaceFixture route="maya" />);
    act(() => {
      runtime.handleNewChat();
    });
    expect(runtime.currentChatId).toBe("next-draft-id");
    expect(selectConversation).not.toHaveBeenCalled();
    expect(mockRouterReplace).toHaveBeenCalledWith("/acme/maya?panel=history", {
      scroll: false,
    });

    mockSearchParams = new URLSearchParams("panel=history");
    rerender(<WorkspaceFixture route="maya" />);
    expect(runtime.currentChatId).toBe("next-draft-id");
    expect(selectConversation).not.toHaveBeenCalled();
  });

  it("does not navigate back when a history selection finishes after leaving Maya", async () => {
    let resolveSelection: (() => void) | undefined;
    selectConversation.mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          resolveSelection = resolve;
        }),
    );
    const { rerender } = render(<WorkspaceFixture />);
    mockPathname = "/acme/maya";
    mockSearchParams = new URLSearchParams("chatRef=saved-chat-1");
    rerender(<WorkspaceFixture route="maya" />);
    mockPathname = "/acme/roadmap";
    mockSearchParams = new URLSearchParams();
    rerender(<WorkspaceFixture route="roadmap" />);

    await act(async () => {
      resolveSelection?.();
    });
    expect(runtime.currentChatId).toBe("saved-chat-1");
    expect(mockRouterReplace).not.toHaveBeenCalled();
  });

  it("starts a new draft without refreshing an already bare Maya URL", () => {
    mockPathname = "/acme/maya";
    render(<WorkspaceFixture route="maya" />);
    act(() => {
      runtime.handleNewChat();
    });

    expect(runtime.currentChatId).toBe("next-draft-id");
    expect(mockRouterReplace).not.toHaveBeenCalled();
  });

  it.each([
    ["linked conversation", "chatRef=saved-chat-1", "saved-chat-1"],
    ["bare draft", "", "draft-chat-id"],
  ])(
    "cancels pending history when Back returns to the active %s",
    async (_, query, chatId) => {
      let resolveSelection: (() => void) | undefined;
      selectConversation.mockImplementationOnce(
        () =>
          new Promise<void>((resolve) => {
            resolveSelection = resolve;
          }),
      );
      mockPathname = "/acme/maya";
      mockSearchParams = new URLSearchParams(query);
      const { rerender } = render(<WorkspaceFixture route="maya" />);
      act(() => {
        runtime.setInput("Keep my active draft");
      });

      mockSearchParams = new URLSearchParams("chatRef=saved-chat-2");
      rerender(<WorkspaceFixture route="maya" />);
      mockSearchParams = new URLSearchParams(query);
      rerender(<WorkspaceFixture route="maya" />);
      await act(async () => {
        resolveSelection?.();
      });

      expect(runtime.currentChatId).toBe(chatId);
      expect(runtime.input).toBe("Keep my active draft");
      expect(mockRouterReplace).not.toHaveBeenCalled();
      expect(stopResponse).not.toHaveBeenCalled();
      expect(disconnectVoice).not.toHaveBeenCalled();
    },
  );

  it("uses the latest runtime callbacks and submits a queued prompt only once", () => {
    render(<WorkspaceFixture />);
    const initialActions = actions;

    act(() => {
      runtime.setInput("Latest draft");
      streamToken();
    });
    act(() => {
      initialActions.openChat("Plan my day");
    });
    expect(submitPrompt).toHaveBeenNthCalledWith(
      1,
      "Plan my day",
      "Latest draft",
      1,
    );

    act(() => {
      streamToken();
    });
    expect(submitPrompt).toHaveBeenCalledTimes(1);
    act(() => {
      initialActions.openChat("Review my priorities");
      initialActions.openChatWithGoogleDriveFile(linkedFile);
    });
    expect(submitPrompt).toHaveBeenNthCalledWith(
      2,
      "Review my priorities",
      "Latest draft",
      2,
    );
    expect(addFile).toHaveBeenCalledWith(
      {
        referenceId: linkedFile.id,
        mimeType: linkedFile.mimeType,
        name: linkedFile.name,
      },
      "Latest draft",
      2,
    );
  });

  it.each(["close", "draft", "file"])(
    "cancels a queued prompt when replaced by the %s action before opening",
    (action) => {
      render(<WorkspaceFixture />);
      act(() => {
        actions.openChat("Do not send this queued prompt");
        if (action === "close") {
          actions.closeChat();
        } else if (action === "draft") {
          actions.openChatWithDraft("Review before sending");
        } else {
          actions.openChatWithGoogleDriveFile(linkedFile, "Review this file");
        }
      });
      expect(submitPrompt).not.toHaveBeenCalled();
      if (action === "close") {
        act(() => {
          actions.openChat();
        });
        expect(submitPrompt).not.toHaveBeenCalled();
      } else {
        expect(runtime.input).toBe(
          action === "draft" ? "Review before sending" : "Review this file",
        );
      }
    },
  );

  it("resets the runtime when the workspace identity changes", () => {
    const { rerender } = render(<WorkspaceFixture />);
    act(() => {
      actions.openChatWithDraft("Workspace-specific draft");
    });
    rerender(<WorkspaceFixture workspace="another-workspace" />);

    expect(runtimeMount).toHaveBeenCalledTimes(2);
    expect(runtimeUnmount).toHaveBeenCalledTimes(1);
    expect(runtime.input).toBe("");
    expect(screen.getByText("Maya closed")).toBeInTheDocument();
  });
});
