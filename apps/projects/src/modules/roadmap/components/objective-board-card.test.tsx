import { fireEvent, render, screen } from "@testing-library/react";
import type { Objective } from "@/modules/objectives/types";
import { ObjectiveBoardCard } from "./objective-board-card";

let mockDragging = false;
const mockPointerDown = jest.fn();
const mockDraggable = jest.fn();
jest.mock("@dnd-kit/core", () => ({
  useDraggable: (options: unknown) => {
    mockDraggable(options);
    return {
      isDragging: mockDragging,
      listeners: { onPointerDown: mockPointerDown },
      setNodeRef: jest.fn(),
    };
  },
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
const OBJECTIVE = {
  id: "objective-1",
  teamId: "team-1",
  sequenceId: 1,
  name: "Improve reliability",
} as Objective;
describe("objective board title navigation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockDragging = false;
  });
  it("provides a native full-objective link while leaving property controls outside its drag area", () => {
    render(
      <ObjectiveBoardCard objective={OBJECTIVE} teamCode="ENG">
        <button type="button">Change status</button>
      </ObjectiveBoardCard>,
    );
    const link = screen.getByRole("link", {
      name: "Improve reliability ENG-1",
    });
    expect(link).toHaveAttribute(
      "href",
      "/acme/teams/team-1/objectives/objective-1",
    );
    fireEvent.pointerDown(link);
    expect(mockPointerDown).toHaveBeenCalledTimes(1);
    fireEvent.pointerDown(
      screen.getByRole("button", { name: "Change status" }),
    );
    expect(mockPointerDown).toHaveBeenCalledTimes(1);
  });
  it("suppresses clicks while its objective is dragging", () => {
    mockDragging = true;
    render(<ObjectiveBoardCard objective={OBJECTIVE} />);
    expect(fireEvent.click(screen.getByRole("link"))).toBe(false);
  });
  it("keeps read-only titles navigable and drag overlays noninteractive", () => {
    const { rerender } = render(
      <ObjectiveBoardCard canDrag={false} objective={OBJECTIVE} />,
    );
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/acme/teams/team-1/objectives/objective-1",
    );
    expect(mockDraggable).toHaveBeenLastCalledWith({
      id: OBJECTIVE.id,
      disabled: true,
    });
    rerender(<ObjectiveBoardCard isOverlay objective={OBJECTIVE} />);
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
