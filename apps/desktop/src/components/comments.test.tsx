import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { WorkspaceData } from "../types";
import { Comments } from "./comments";
vi.mock("../hooks/use-draft", () => ({
  useDraft: () => ({
    value: { comment: "" },
    ready: true,
    update: vi.fn(),
    status: "",
    error: null,
  }),
}));
const data: WorkspaceData = {
  workspaceId: "workspace",
  stories: [],
  documents: [],
  teams: [],
  statuses: [],
  comments: [
    {
      id: "root",
      storyId: "story",
      comment: "Root comment",
      createdAt: "2026-10-14",
      subComments: [
        {
          id: "reply",
          storyId: "story",
          comment: "Nested reply",
          createdAt: "2026-10-14",
        },
      ],
    },
  ],
  pending: [],
  lastSyncedAt: null,
};

describe("downloaded activity", () => {
  it("shows the partial-history notice until this specific story is fully downloaded and includes replies", () => {
    const { rerender } = render(
      <Comments storyId="story" data={data} save={vi.fn()} />,
    );
    expect(
      screen.getByText(/Only downloaded comments are shown/),
    ).toBeInTheDocument();
    expect(screen.getByText("Nested reply")).toBeInTheDocument();
    rerender(
      <Comments
        storyId="story"
        data={{ ...data, commentsCompleteFor: ["another-story"] }}
        save={vi.fn()}
      />,
    );
    expect(
      screen.getByText(/Only downloaded comments are shown/),
    ).toBeInTheDocument();
    rerender(
      <Comments
        storyId="story"
        data={{ ...data, commentsCompleteFor: ["story"] }}
        save={vi.fn()}
      />,
    );
    expect(
      screen.queryByText(/Only downloaded comments are shown/),
    ).not.toBeInTheDocument();
  });
});
