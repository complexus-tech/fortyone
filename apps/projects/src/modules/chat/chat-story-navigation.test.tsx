import type { ReactNode } from "react";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useNativeLinkNavigation } from "@/components/ui/link-preview/native-navigation";
import { useChat } from "./use-chat";
import { ChatShell } from "./chat-shell";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("./use-chat", () => ({ useChat: jest.fn() }));
jest.mock("./use-chat-maya-runs", () => ({
  useChatMayaRuns: () => ({ data: { enabled: false } }),
}));
jest.mock("./audio-access-guard", () => ({ AudioAccessGuard: () => null }));
jest.mock("./audio-playback-context", () => ({
  AudioPlaybackProvider: ({ children }: { children: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("./message-edit-drafts", () => ({
  MessageEditDraftProvider: ({ children }: { children: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("./attachment-preview-context", () => ({
  ChatAttachmentPreviewProvider: ({ children }: { children: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("./chat-context", () => ({
  ChatIdentityProvider: ({ children }: { children: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("./conversation-list", () => ({ ConversationList: () => null }));
jest.mock("./thread-panel", () => ({ ThreadPanel: () => null }));
jest.mock("./new-message", () => ({ NewMessage: () => null }));
jest.mock("./story-panel", () => ({
  ChatStoryPanel: ({ onClose }: { onClose: () => void }) => (
    <div aria-label="Story details" role="dialog">
      <button onClick={onClose} type="button">
        Close story
      </button>
    </div>
  ),
}));
jest.mock("./conversation-view", () => ({
  ConversationView: ({
    hasThread,
    onBack,
  }: {
    hasThread: boolean;
    onBack: () => void;
  }) => {
    const open = useNativeLinkNavigation();
    return (
      <div data-panel-open={hasThread} data-testid="conversation">
        <button
          onClick={(event) => {
            open?.(
              {
                kind: "story",
                id: "TM-276",
                workspaceSlug: "acme",
                href: "https://acme.fortyone.app/work/TM-276",
              },
              event.currentTarget,
            );
          }}
          type="button"
        >
          View story
        </button>
        <button onClick={onBack} type="button">
          Back to conversations
        </button>
      </div>
    );
  },
}));

const state = {
  activeId: "conversation",
  activeTab: "messages",
  showConversation: false,
  isComposing: false,
  conversation: { id: "conversation", kind: "team", members: [] },
  conversationList: {},
  detail: {},
  history: {},
  sendingKeys: new Set(),
  closeThread: jest.fn(),
  conversationHeadingRef: { current: null },
  showList: jest.fn(),
} as unknown as ReturnType<typeof useChat>;

let queryClient: QueryClient;
const TestProviders = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

beforeEach(() => {
  queryClient = new QueryClient();
  jest.mocked(useChat).mockReturnValue(state);
});
afterEach(() => {
  cleanup();
  queryClient.clear();
  jest.clearAllMocks();
});

it("opens a story from the desktop conversation even before mobile conversation selection", () => {
  render(<ChatShell />, { wrapper: TestProviders });
  fireEvent.click(screen.getByRole("button", { name: "View story" }));
  expect(screen.getByRole("dialog", { name: "Story details" })).toBeVisible();
  expect(screen.getByTestId("conversation")).toHaveAttribute(
    "data-panel-open",
    "true",
  );
  fireEvent.click(screen.getByRole("button", { name: "Close story" }));
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
});

it("closes the story when returning to the conversation list or switching conversation scope", () => {
  const { rerender } = render(<ChatShell />, { wrapper: TestProviders });
  fireEvent.click(screen.getByRole("button", { name: "View story" }));
  fireEvent.click(
    screen.getByRole("button", { name: "Back to conversations" }),
  );
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  expect(state.showList).toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "View story" }));
  jest
    .mocked(useChat)
    .mockReturnValue({ ...state, activeId: "another-conversation" });
  rerender(<ChatShell />);
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
});
