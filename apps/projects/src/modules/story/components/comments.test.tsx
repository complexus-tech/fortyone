import { fireEvent, render, screen } from "@testing-library/react";
import { FORMER_USER_ID } from "@/lib/former-user";
import { useStoryCommentsInfinite } from "@/modules/story/hooks/story-comments";
import { Comments } from "./comments";

jest.mock("../hooks/comment-maya", () => ({
  useStoryMayaRuns: () => ({ data: { enabled: false, runs: [] } }),
  useMayaRunRefresh: () => undefined,
  useCommentMayaAction: () => ({ mutate: jest.fn(), isPending: false }),
}));

jest.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: "acme",
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("../hooks/comment-collaboration", () => ({
  useStoryCommentState: () => ({ data: { comments: [], unreadCount: 0 } }),
  useCommentThread: () => ({}),
  useCommentCollaboration: () => ({
    reaction: { mutate: jest.fn() },
    resolution: { mutate: jest.fn() },
  }),
}));
jest.mock("../hooks/comment-read-tracking", () => ({
  useCommentReadTracking: () => ({
    containerRef: { current: null },
    lastReadAt: undefined,
    hadUnread: false,
  }),
  useCommentThreadReadTracking: () => ({ current: null }),
}));
jest.mock("../hooks/use-comment-seen", () => ({
  useCommentSeen: () => ({ current: null }),
}));

jest.mock("ui", () => ({ ...jest.requireActual("ui"), TimeAgo: () => null }));

jest.mock("@/modules/story/hooks/story-comments", () => ({
  useStoryCommentsInfinite: jest.fn(),
}));
jest.mock("@/lib/auth/client", () => ({ useSession: () => ({ data: null }) }));
jest.mock("@/lib/hooks/delete-comment-mutation", () => ({
  useDeleteCommentMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/modules/story/components/comment-input", () => ({
  CommentInput: () => null,
}));
jest.mock("./comment-attachment-list", () => ({
  CommentAttachmentList: () => null,
}));
jest.mock("./comment-latest-control", () => ({
  CommentLatestControl: () => null,
}));

it("keeps retained comments and reply threads visible without a former-user profile or stale identity", () => {
  const comment = {
    id: "comment-1",
    storyId: "story-1",
    userId: FORMER_USER_ID,
    user: {
      id: FORMER_USER_ID,
      fullName: "Stale name",
      username: "stale-name",
      avatarUrl: "https://example.com/stale.png",
      isSystem: true,
    },
    createdAt: "2026-09-15T12:00:00Z",
    comment: "<p>Shared decision remains.</p>",
    subComments: [],
  };
  jest.mocked(useStoryCommentsInfinite).mockReturnValue({
    data: {
      pages: [
        {
          comments: [
            {
              ...comment,
              subComments: [
                {
                  ...comment,
                  id: "reply-1",
                  comment: "<p>Shared reply remains.</p>",
                },
              ],
            },
          ],
        },
      ],
    },
    hasNextPage: false,
  } as unknown as ReturnType<typeof useStoryCommentsInfinite>);
  const { container } = render(<Comments storyId="story-1" teamId="team-1" />);
  expect(screen.getByText("Shared decision remains.")).toBeInTheDocument();
  expect(screen.queryByText("Shared reply remains.")).not.toBeInTheDocument();
  fireEvent.click(
    screen.getByRole("button", { name: "1 reply. Show replies" }),
  );
  expect(screen.getByText("Shared reply remains.")).toBeInTheDocument();
  expect(screen.getAllByText("Former user").length).toBeGreaterThan(0);
  expect(screen.queryByRole("link")).not.toBeInTheDocument();
  expect(container.innerHTML).not.toContain("Stale name");
  expect(container.innerHTML).not.toContain("stale.png");
  expect(screen.queryByText("(System Account)")).not.toBeInTheDocument();
});
