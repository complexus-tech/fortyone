/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import { render, screen } from "@testing-library/react";
import { useDraggable } from "@dnd-kit/core";
import type { Story } from "@/modules/stories/types";
import { StoryCardPreview } from "./card";

jest.mock("@dnd-kit/core", () => ({
  useDraggable: jest.fn(),
}));
jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactModule>("react");
  const Primitive = ({ children }: { children?: ReactModule.ReactNode }) =>
    React.createElement("div", null, children);

  return {
    Avatar: Primitive,
    Badge: Primitive,
    Box: Primitive,
    Button: Primitive,
    Flex: Primitive,
    Text: Primitive,
  };
});
jest.mock("lib", () => ({
  cn: (...values: unknown[]) =>
    values.filter((value) => typeof value === "string").join(" "),
}));
jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children }: { children?: ReactModule.ReactNode }) => children,
}));
jest.mock("../priority-icon", () => ({
  PriorityIcon: () => null,
}));
jest.mock("../board-context", () => ({
  useBoard: jest.fn(),
}));
jest.mock("./assignees-menu", () => ({ AssigneesMenu: () => null }));
jest.mock("./context-menu", () => ({ StoryContextMenu: () => null }));
jest.mock("./properties", () => ({ StoryProperties: () => null }));
jest.mock("@/modules/settings/workspace/integrations/figma/icon", () => ({
  FigmaIcon: () => null,
}));
jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("@/hooks", () => ({
  useMediaQuery: jest.fn(),
  useUserRole: jest.fn(),
  useWorkspacePath: jest.fn(),
}));
jest.mock("@tanstack/react-query", () => ({ useQueryClient: jest.fn() }));
jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: jest.fn(),
}));
jest.mock("@/lib/hooks/figma", () => ({
  useFigmaHandoffStatuses: jest.fn(),
}));
jest.mock("@/modules/story/queries/get-story", () => ({
  getStory: jest.fn(),
}));
jest.mock("@/modules/story/queries/get-attachments", () => ({
  getStoryAttachments: jest.fn(),
}));
jest.mock("@/lib/queries/links/get-links", () => ({
  getLinks: jest.fn(),
}));

const story = {
  assignee: {
    avatarUrl: null,
    fullName: "Joseph Mukorivo",
    id: "user-1",
    isActive: true,
    isSystem: false,
    username: "joseph",
  },
  estimateValue: 5,
  id: "story-1",
  priority: "High",
  sequenceId: 41,
  team: {
    code: "PRO",
    id: "team-1",
    name: "Product",
  },
  title: "Make Kanban interactions smooth",
} as Story;

describe("StoryCardPreview", () => {
  it("renders a lightweight snapshot without registering another draggable", () => {
    render(<StoryCardPreview story={story} />);

    expect(
      screen.getByText("Make Kanban interactions smooth"),
    ).toBeInTheDocument();
    expect(screen.getByText("PRO-41")).toBeInTheDocument();
    expect(screen.getByText("High")).toBeInTheDocument();
    expect(useDraggable).not.toHaveBeenCalled();
  });
});
