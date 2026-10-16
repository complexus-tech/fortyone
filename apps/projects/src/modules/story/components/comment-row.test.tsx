/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentProps } from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import type { Comment } from "@/types/comment";
import { FORMER_USER_ID } from "@/lib/former-user";
import { CommentRow } from "./comment-row";

jest.mock("ui", () => ({ ...jest.requireActual("ui"), TimeAgo: () => null }));
jest.mock("@/lib/hooks/delete-comment-mutation", () => ({
  useDeleteCommentMutation: () => ({ mutate: jest.fn(), isPending: false }),
}));
jest.mock("../hooks/comment-collaboration", () => ({
  useCommentCollaboration: () => ({
    reaction: { mutate: jest.fn(), isPending: false },
    resolution: { mutate: jest.fn(), isPending: false },
  }),
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

describe("collapsed comment replies", () => {
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
    const expanded = screen.getByRole("button", {
      name: "3 replies. Hide replies",
    });
    expect(expanded).toHaveAttribute("aria-expanded", "true");
    fireEvent.click(expanded);
    expect(screen.queryByText("first content")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "3 replies. Show replies" }),
    ).toHaveAttribute("aria-expanded", "false");
  });

  it("opens the thread for a linked reply and permits explicit collapse afterwards", () => {
    render(<CommentRow {...propsFor({ targetId: "second" })} />);
    expect(screen.getByText("second content")).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "3 replies. Hide replies" }),
    );
    expect(screen.queryByText("second content")).not.toBeInTheDocument();
  });

  it("opens a newly targeted reply when the link changes on a mounted thread", () => {
    const props = propsFor();
    const { rerender } = render(<CommentRow {...props} />);
    rerender(<CommentRow {...props} targetId="third" />);
    expect(screen.getByText("third content")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "3 replies. Hide replies" }),
    ).toHaveAttribute("aria-expanded", "true");
  });

  it("reveals the same latest reply again after its thread was manually collapsed", () => {
    const props = propsFor({ targetId: "second" });
    const { rerender } = render(<CommentRow {...props} revealKey="second:1" />);
    fireEvent.click(
      screen.getByRole("button", { name: "3 replies. Hide replies" }),
    );
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
      screen.getByRole("button", { name: "3 replies. Hide replies" }),
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
