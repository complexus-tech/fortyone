/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentProps } from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import type { Comment } from "@/types/comment";
import { FORMER_USER_ID } from "@/lib/former-user";
import type { CommentMayaRun } from "../types/comment-maya";
import { normalizeCommentThreads } from "../lib/comment-tree";
import { useCommentThreadReadTracking } from "../hooks/comment-read-tracking";
import { CommentRow } from "./comment-row";

jest.mock("ui", () => ({ ...jest.requireActual("ui"), TimeAgo: () => null }));
jest.mock("@/lib/hooks/delete-comment-mutation", () => ({
  useDeleteCommentMutation: () => ({ mutate: jest.fn(), isPending: false }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: () => ({ data: null }),
}));
jest.mock("../hooks/comment-collaboration", () => ({
  useCommentCollaboration: () => ({
    reaction: { mutate: jest.fn(), isPending: false },
    resolution: { mutate: jest.fn(), isPending: false },
  }),
}));
jest.mock("../hooks/comment-read-tracking", () => ({
  useCommentThreadReadTracking: jest.fn(() => ({ current: null })),
}));
jest.mock("../hooks/use-comment-seen", () => ({
  useCommentSeen: () => ({ current: null }),
}));
jest.mock("../hooks/comment-maya", () => ({
  useCommentMayaAction: () => ({ mutate: jest.fn(), isPending: false }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: "acme",
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("./comment-attachment-list", () => ({
  CommentAttachmentList: () => null,
}));
jest.mock("./comment-input", () => ({
  CommentInput: ({
    parentId,
    onCancel,
  }: {
    parentId?: string;
    onCancel?: () => void;
  }) => (
    <div aria-label={`Reply composer for ${parentId}`} role="region">
      <button onClick={onCancel} type="button">
        Cancel reply
      </button>
    </div>
  ),
}));

const createComment = (
  id: string,
  author: string,
  overrides: Partial<Comment> = {},
): Comment => ({
  id,
  storyId: "story",
  parentId: id === "root" ? null : "root",
  userId: author,
  user: {
    id: author,
    fullName: author,
    username: author.toLowerCase(),
    avatarUrl: `/avatars/${author.toLowerCase()}.png`,
    isActive: true,
    isSystem: false,
  },
  comment: `<p>${id} content</p>`,
  createdAt: "2026-10-16T12:00:00Z",
  updatedAt: "2026-10-16T12:00:00Z",
  subComments: [],
  ...overrides,
});

const thread = () =>
  createComment("root", "Root", {
    subComments: [
      createComment("first", "Alice"),
      createComment("second", "Bob"),
      createComment("third", "Alice"),
    ],
  });
const propsFor = (
  overrides: Partial<ComponentProps<typeof CommentRow>> = {},
): ComponentProps<typeof CommentRow> => ({
  comment: thread(),
  storyId: "story",
  teamId: "team",
  currentUserId: "viewer",
  stateById: new Map(),
  targetId: null,
  lastReadAt: undefined,
  ...overrides,
});

const pendingRun: CommentMayaRun = {
  id: "pending-maya",
  sourceId: "root",
  requesterId: "viewer",
  status: "running",
  summary: "Internal progress must stay private",
  updatedAt: "2026-10-16T12:00:00Z",
  version: 1,
};
const openReplies = () => {
  fireEvent.click(screen.getByRole("button", { name: /Show replies/ }));
};

describe("Maya pending replies", () => {
  it("keeps a newly published Maya reply collapsed until the reader opens it", () => {
    const props = propsFor({ comment: createComment("root", "Root") });
    const { rerender } = render(
      <CommentRow
        {...props}
        mayaPending={[{ run: pendingRun, status: "thinking" }]}
      />,
    );
    rerender(
      <CommentRow
        {...props}
        comment={{
          ...props.comment,
          subComments: [createComment("maya-reply", "Maya")],
        }}
      />,
    );
    expect(screen.queryByText("maya-reply content")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "1 reply. Show replies" }),
    ).toBeInTheDocument();
    openReplies();
    expect(screen.getByText("maya-reply content")).toBeInTheDocument();
  });
  it("shows pending feedback in the collapsed summary without counting it as a reply", () => {
    const props = propsFor({
      comment: createComment("root", "Root"),
      mayaPending: [{ run: pendingRun, status: "thinking" }],
    });
    const { rerender } = render(<CommentRow {...props} />);
    expect(screen.getByRole("status")).toHaveTextContent("Maya is thinking");
    expect(
      screen.queryByText(/0 replies|1 reply|Internal progress/),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Hide replies" }),
    ).not.toBeInTheDocument();
    openReplies();
    fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
    expect(
      screen.getByRole("button", { name: "Show replies and Maya’s response" }),
    ).toBeInTheDocument();
    rerender(
      <CommentRow
        {...props}
        mayaPending={[
          { run: { ...pendingRun, version: 2 }, status: "working" },
        ]}
      />,
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      "Maya is still working on this",
    );
    expect(
      screen.queryByRole("button", { name: "Hide replies" }),
    ).not.toBeInTheDocument();
  });

  it("replaces pending feedback with the actual reply in the same expanded thread", () => {
    const props = propsFor({
      comment: createComment("root", "Root"),
      mayaPending: [{ run: pendingRun, status: "thinking" }],
    });
    const { rerender } = render(<CommentRow {...props} />);
    openReplies();
    const reply = createComment("maya-reply", "Maya", {
      user: { ...createComment("maya", "Maya").user, isSystem: true },
      comment: "<p>Updated this story’s priority.</p>",
    });
    rerender(
      <CommentRow
        {...props}
        comment={{ ...props.comment, subComments: [reply] }}
        mayaPending={[]}
      />,
    );
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(
      screen.getByText("Updated this story’s priority."),
    ).toBeInTheDocument();
    expect(screen.queryByText("1 reply")).not.toBeInTheDocument();
  });

  it("keeps another member's collapsed thread compact and puts pending feedback after its real replies when opened", () => {
    const props = propsFor({
      mayaPending: [
        {
          run: { ...pendingRun, requesterId: "another", sourceId: "first" },
          status: "thinking",
        },
      ],
    });
    const { container } = render(<CommentRow {...props} />);
    expect(screen.queryByText("first content")).not.toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Show replies and Maya’s response" }),
    );
    const replies = container.querySelector<HTMLElement>(
      "#comment-replies-root",
    )!;
    expect(within(replies).getByText("third content")).toBeInTheDocument();
    expect(within(replies).getAllByRole("status")).toHaveLength(1);
    expect(within(replies).queryByText("4 replies")).not.toBeInTheDocument();
  });

  it("checks status without resubmitting and limits retry to the requester", () => {
    const onCheckMaya = jest.fn();
    const onRetryMaya = jest.fn();
    const props = propsFor({
      onCheckMaya,
      onRetryMaya,
      mayaPending: [{ run: pendingRun, status: "delayed" }],
    });
    const { rerender } = render(<CommentRow {...props} />);
    openReplies();
    fireEvent.click(screen.getByRole("button", { name: "Check again" }));
    expect(onCheckMaya).toHaveBeenCalledTimes(1);
    expect(onRetryMaya).not.toHaveBeenCalled();
    rerender(
      <CommentRow
        {...props}
        mayaPending={[
          { run: { ...pendingRun, status: "failed" }, status: "failed" },
        ]}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetryMaya).toHaveBeenCalledWith(
      expect.objectContaining({ id: pendingRun.id }),
    );
    rerender(
      <CommentRow
        {...props}
        mayaPending={[
          {
            run: { ...pendingRun, requesterId: "another", status: "failed" },
            status: "failed",
          },
        ]}
      />,
    );
    expect(
      screen.queryByRole("button", { name: "Retry" }),
    ).not.toBeInTheDocument();
  });
});

describe("collapsed comment replies", () => {
  it("keeps Reply before the reaction control without unused emoji buttons", () => {
    const { container } = render(<CommentRow {...propsFor()} />);
    const footer = container.querySelector(".comment-footer")!;
    expect(
      within(footer as HTMLElement)
        .getAllByRole("button")
        .map(
          (button) => button.getAttribute("aria-label") ?? button.textContent,
        ),
    ).toEqual(["Reply", "Add reaction"]);
  });

  it("keeps single replies and newly active threads collapsed until opened", () => {
    const comment = thread();
    const { rerender } = render(
      <CommentRow
        {...propsFor({
          comment: { ...comment, subComments: comment.subComments.slice(0, 1) },
        })}
      />,
    );
    expect(screen.queryByText("first content")).not.toBeInTheDocument();
    expect(screen.getByText("1 reply")).toBeInTheDocument();
    rerender(<CommentRow {...propsFor({ comment })} />);
    expect(screen.queryByText("third content")).not.toBeInTheDocument();
    openReplies();
    rerender(<CommentRow {...propsFor({ comment })} />);
    expect(screen.getByText("third content")).toBeInTheDocument();
  });

  it("keeps an explicit collapse through new replies and restores the dot for a later unread reply", () => {
    const comment = thread();
    const { rerender } = render(
      <CommentRow {...propsFor({ comment, lastReadAt: null })} />,
    );
    openReplies();
    rerender(
      <CommentRow
        {...propsFor({
          comment,
          lastReadAt: "2026-10-16T12:00:00Z",
        })}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
    expect(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    ).not.toHaveAccessibleDescription();
    const newReply = createComment("new-reply", "Alice", {
      createdAt: "2026-10-16T13:00:00Z",
    });
    rerender(
      <CommentRow
        {...propsFor({
          comment: {
            ...comment,
            subComments: [...comment.subComments, newReply],
          },
          lastReadAt: "2026-10-16T12:00:00Z",
        })}
      />,
    );
    expect(screen.queryByText("new-reply content")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "4 replies. Show replies" }),
    ).toHaveAccessibleDescription("New replies");
  });
  it("uses only the unread dot for new replies, with no heading or divider when expanded", () => {
    const { rerender } = render(
      <CommentRow {...propsFor({ lastReadAt: "2026-10-15T12:00:00Z" })} />,
    );
    const summary = screen.getByRole("button", {
      name: "3 replies. Show replies",
    });
    expect(summary).toHaveAccessibleDescription("New replies");
    expect(screen.queryByText("New comments")).not.toBeInTheDocument();
    fireEvent.click(summary);
    expect(screen.queryByText("New comments")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
    rerender(
      <CommentRow {...propsFor({ lastReadAt: "2026-10-16T12:00:00Z" })} />,
    );
    expect(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    ).not.toHaveAccessibleDescription();
  });
  it("keeps internal Maya states out of the conversation and offers the ordinary Reply on Maya's response", () => {
    const comment = thread();
    comment.subComments[1] = {
      ...comment.subComments[1],
      user: {
        ...comment.subComments[1].user,
        isSystem: true,
        fullName: "Maya",
        username: "maya",
      },
    };
    const run: CommentMayaRun = {
      id: "run",
      sourceId: "first",
      requesterId: "viewer",
      status: "running",
      summary: "Internal progress",
      updatedAt: "2026-10-16T12:00:00Z",
      version: 1,
    };
    const { rerender, container } = render(
      <CommentRow {...propsFor({ comment, mayaRuns: [run] })} />,
    );
    expect(
      screen.queryByText(/Maya is working|Internal progress/),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", {
        name: /Cancel request|Reply to Maya|View reply/,
      }),
    ).not.toBeInTheDocument();
    rerender(
      <CommentRow
        {...propsFor({
          comment,
          mayaRuns: [
            {
              ...run,
              status: "succeeded",
              replyCommentId: "second",
              version: 2,
            },
          ],
        })}
      />,
    );
    expect(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    ).toHaveAttribute("aria-expanded", "false");
    fireEvent.click(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    );
    fireEvent.click(
      within(
        container.querySelector<HTMLElement>("#comment-second")!,
      ).getByRole("button", {
        name: /^Reply$/,
      }),
    );
    expect(
      screen.getByRole("region", { name: "Reply composer for second" }),
    ).toBeInTheDocument();
    expect(screen.queryByText("Maya finished")).not.toBeInTheDocument();
  });

  it("places approval only inside the actual Maya reply, never on its source or collapsed parent", () => {
    const comment = thread();
    comment.subComments[1] = {
      ...comment.subComments[1],
      user: { ...comment.subComments[1].user, isSystem: true },
    };
    const run: CommentMayaRun = {
      id: "run",
      sourceId: "first",
      requesterId: "viewer",
      status: "awaiting_approval",
      summary: "Update the story?",
      proposalReview: "Status\nBefore: In progress\nAfter: Done",
      replyCommentId: "second",
      updatedAt: "2026-10-16T12:00:00Z",
      version: 2,
    };
    const { container } = render(
      <CommentRow {...propsFor({ comment, mayaRuns: [run] })} />,
    );
    expect(
      screen.queryByText("Review proposed changes"),
    ).not.toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    );
    const reply = container.querySelector<HTMLElement>("#comment-second")!;
    expect(
      within(reply).getByText("Review proposed changes"),
    ).toBeInTheDocument();
    expect(
      within(reply).getByRole("button", { name: "Approve and apply" }),
    ).toBeInTheDocument();
    expect(
      within(
        container.querySelector<HTMLElement>("#comment-first")!,
      ).queryByRole("button", {
        name: "Approve and apply",
      }),
    ).not.toBeInTheDocument();
  });

  it("starts collapsed, shows unique real reply authors and total reply count, and toggles on summary click", () => {
    render(<CommentRow {...propsFor()} />);
    expect(screen.getByText("root content")).toBeInTheDocument();
    expect(screen.queryByText("first content")).not.toBeInTheDocument();
    const summary = screen.getByRole("button", {
      name: "3 replies. Show replies",
    });
    expect(summary).toHaveAttribute("aria-expanded", "false");
    expect(within(summary).getAllByRole("img", { name: "Alice" })).toHaveLength(
      1,
    );
    expect(within(summary).getAllByRole("img", { name: "Bob" })).toHaveLength(
      1,
    );
    expect(within(summary).getByRole("img", { name: "Alice" })).toHaveAttribute(
      "src",
      expect.stringContaining("alice.png"),
    );
    expect(
      within(summary).queryByRole("img", { name: "Root" }),
    ).not.toBeInTheDocument();

    fireEvent.click(summary);
    expect(screen.getByText("first content")).toBeInTheDocument();
    expect(screen.getByText("second content")).toBeInTheDocument();
    expect(screen.getByText("third content")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "3 replies. Show replies" }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("3 replies")).not.toBeInTheDocument();
    const expanded = screen.getByRole("button", {
      name: "Hide replies",
    });
    expect(expanded).toHaveAttribute("aria-expanded", "true");
    fireEvent.click(expanded);
    expect(screen.queryByText("first content")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    ).toHaveAttribute("aria-expanded", "false");
  });

  it("caps a long thread at four distinct replying avatars without including the parent", () => {
    const authors = ["Alice", "Bob", "Charlie", "Dana", "Evan"];
    const comment = createComment("root", "Root", {
      subComments: Array.from({ length: 20 }, (_, i) =>
        createComment(`reply-${i}`, authors[i % authors.length]),
      ),
    });
    render(<CommentRow {...propsFor({ comment })} />);
    const summary = screen.getByRole("button", {
      name: "20 replies. Show replies",
    });
    expect(within(summary).getAllByRole("img")).toHaveLength(4);
    for (const author of authors.slice(0, 4))
      expect(
        within(summary).getByRole("img", { name: author }),
      ).toBeInTheDocument();
    expect(
      within(summary).queryByRole("img", { name: "Evan" }),
    ).not.toBeInTheDocument();
    expect(
      within(summary).queryByRole("img", { name: "Root" }),
    ).not.toBeInTheDocument();
  });

  it("opens the thread for a linked reply and permits explicit collapse afterwards", () => {
    render(<CommentRow {...propsFor({ targetId: "second" })} />);
    expect(screen.getByText("second content")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
    expect(screen.queryByText("second content")).not.toBeInTheDocument();
  });

  it("removes an expired highlight while keeping the linked reply expanded", () => {
    const props = propsFor({ targetId: "second", highlightedId: "second" });
    const { container, rerender } = render(<CommentRow {...props} />);
    const target = container.querySelector("#comment-second");
    expect(target).toHaveClass("bg-primary/5", "ring-1");
    expect(target).toHaveClass("motion-reduce:transition-none");

    rerender(<CommentRow {...props} highlightedId={null} />);
    expect(target).not.toHaveClass("bg-primary/5", "ring-1");
    expect(screen.getByText("second content")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Hide replies" }),
    ).toHaveAttribute("aria-expanded", "true");
  });

  it("opens a newly targeted reply when the link changes on a mounted thread", () => {
    const props = propsFor();
    const { rerender } = render(<CommentRow {...props} />);
    rerender(<CommentRow {...props} targetId="third" />);
    expect(screen.getByText("third content")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Hide replies" }),
    ).toHaveAttribute("aria-expanded", "true");
  });

  it("reveals the same latest reply again after its thread was manually collapsed", () => {
    const props = propsFor({ targetId: "second" });
    const { rerender } = render(<CommentRow {...props} revealKey="second:1" />);
    fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
    expect(screen.queryByText("second content")).not.toBeInTheDocument();
    rerender(<CommentRow {...props} revealKey="second:2" />);
    expect(screen.getByText("second content")).toBeInTheDocument();
  });

  it("opens existing reply context when starting a reply", () => {
    render(<CommentRow {...propsFor()} />);
    fireEvent.click(screen.getByRole("button", { name: /^Reply$/ }));
    expect(screen.getByText("first content")).toBeInTheDocument();
    expect(
      screen.getByRole("region", { name: "Reply composer for root" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Hide replies" }),
    ).toHaveAttribute("aria-expanded", "true");
  });

  it("uses a neutral former-user identity in the summary and expanded reply", () => {
    const former = createComment("former", FORMER_USER_ID, {
      user: {
        id: FORMER_USER_ID,
        fullName: "Stale name",
        username: "stale-name",
        avatarUrl: "/avatars/stale.png",
        isActive: false,
        isSystem: true,
      },
    });
    const { container } = render(
      <CommentRow
        {...propsFor({
          comment: createComment("root", "Root", { subComments: [former] }),
        })}
      />,
    );
    const summary = screen.getByRole("button", {
      name: "1 reply. Show replies",
    });
    expect(summary).toHaveTextContent("1 reply");
    expect(container.innerHTML).not.toContain("Stale name");
    expect(container.innerHTML).not.toContain("stale.png");
    fireEvent.click(summary);
    expect(screen.getByText("former content")).toBeInTheDocument();
    expect(screen.getAllByText("Former user").length).toBeGreaterThan(0);
    expect(container.innerHTML).not.toContain("stale-name");
    expect(container.innerHTML).not.toContain("stale.png");
  });
});

describe("three-level comment threads", () => {
  const nestedThread = () => {
    const second = createComment("second-level", "Alice");
    const third = createComment("third-level", "Bob", {
      parentId: second.id,
    });
    const maya = createComment("third-level-maya", "Maya", {
      parentId: second.id,
      user: { ...createComment("maya", "Maya").user, isSystem: true },
    });
    return createComment("root", "Root", {
      subComments: [second, third, maya],
    });
  };
  const footerFor = (container: HTMLElement, id: string) =>
    container.querySelector<HTMLElement>(`#comment-${id} > .comment-footer`)!;

  it("allows replies at levels one and two, but not level three for humans or Maya", () => {
    const comment = normalizeCommentThreads([nestedThread()])[0];
    const { container } = render(<CommentRow {...propsFor({ comment })} />);
    openReplies();
    for (const id of ["root", "second-level"]) {
      expect(
        within(footerFor(container, id)).getByRole("button", { name: "Reply" }),
      ).toBeInTheDocument();
    }
    for (const id of ["third-level", "third-level-maya"]) {
      expect(container.querySelector(`#comment-${id}`)).toHaveAttribute(
        "data-comment-depth",
        "3",
      );
      expect(
        within(footerFor(container, id)).queryByRole("button", {
          name: "Reply",
        }),
      ).not.toBeInTheDocument();
      expect(
        within(footerFor(container, id)).getByRole("button", {
          name: "Add reaction",
        }),
      ).toBeInTheDocument();
    }
    expect(
      screen.getAllByRole("button", { name: "Hide replies" }),
    ).toHaveLength(1);
  });

  it("does not offer replies anywhere in a resolved thread", () => {
    const comment = normalizeCommentThreads([nestedThread()])[0];
    render(
      <CommentRow
        {...propsFor({
          comment,
          targetId: "third-level",
          stateById: new Map([
            [
              "root",
              {
                commentId: "root",
                reactions: [],
                resolvedAt: "2026-10-16T12:01:00Z",
                resolvedBy: "viewer",
              },
            ],
          ]),
        })}
      />,
    );
    expect(
      screen.queryByRole("button", { name: "Reply" }),
    ).not.toBeInTheDocument();
  });

  it("keeps the second-level composer and exact reply target through a flat server refresh", () => {
    const flat = nestedThread();
    const optimistic = normalizeCommentThreads([flat])[0];
    const { container, rerender } = render(
      <CommentRow {...propsFor({ comment: optimistic })} />,
    );
    openReplies();
    fireEvent.click(
      within(footerFor(container, "second-level")).getByRole("button", {
        name: "Reply",
      }),
    );
    const composer = screen.getByRole("region", {
      name: "Reply composer for second-level",
    });
    rerender(
      <CommentRow
        {...propsFor({ comment: normalizeCommentThreads([flat])[0] })}
      />,
    );
    expect(
      screen.getByRole("region", { name: "Reply composer for second-level" }),
    ).toBe(composer);
    expect(
      container.querySelector("#comment-second-level #comment-third-level"),
    ).toBeInTheDocument();
  });

  it("reveals all nested replies from a third-level link with only the root tracking reads", () => {
    jest.mocked(useCommentThreadReadTracking).mockClear();
    const comment = normalizeCommentThreads([nestedThread()])[0];
    render(<CommentRow {...propsFor({ comment, targetId: "third-level" })} />);
    expect(screen.getByText("third-level content")).toBeInTheDocument();
    expect(screen.getByText("third-level-maya content")).toBeInTheDocument();
    const trackingCalls = jest.mocked(useCommentThreadReadTracking).mock.calls;
    expect(
      trackingCalls
        .filter(([, , enabled]) => enabled)
        .every(([, id]) => id === "root"),
    ).toBe(true);
    expect(
      trackingCalls.some(([, id, enabled]) => id === "root" && enabled),
    ).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
    expect(screen.queryByText("third-level content")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    ).toBeInTheDocument();
  });

  it.each(["second-level", "third-level"])(
    "keeps Maya's pending and published reply beneath the same displayed parent for %s",
    (sourceId) => {
      const flat = nestedThread();
      const props = propsFor({
        comment: normalizeCommentThreads([flat])[0],
        mayaPending: [{ run: { ...pendingRun, sourceId }, status: "thinking" }],
      });
      const { container, rerender } = render(<CommentRow {...props} />);
      openReplies();
      const replyList = container.querySelector<HTMLElement>(
        "#comment-replies-second-level",
      )!;
      expect(within(replyList).getAllByRole("status")).toHaveLength(1);
      expect(screen.getAllByRole("status")).toHaveLength(1);
      const published = createComment("published", "Maya", {
        parentId: sourceId,
      });
      rerender(
        <CommentRow
          {...props}
          comment={
            normalizeCommentThreads([
              { ...flat, subComments: [...flat.subComments, published] },
            ])[0]
          }
          mayaPending={[]}
        />,
      );
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
      expect(
        within(replyList).getByText("published content"),
      ).toBeInTheDocument();
      expect(container.querySelector("#comment-published")).toHaveAttribute(
        "data-comment-depth",
        "3",
      );
    },
  );
});
