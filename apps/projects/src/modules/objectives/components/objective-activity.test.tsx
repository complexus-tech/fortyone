import { render, screen } from "@testing-library/react";
import { FORMER_USER_ID } from "@/lib/former-user";
import type { ObjectiveActivity } from "../types";
import { ObjectiveActivityComponent } from "./objective-activity";

jest.mock("react-markdown", () => ({
  __esModule: true,
  default: ({ children }: { children: string }) => (
    <div data-testid="markdown-comment">{children}</div>
  ),
}));
jest.mock("remark-gfm", () => ({ __esModule: true, default: () => undefined }));

jest.mock("ui", () => ({ ...jest.requireActual("ui"), TimeAgo: () => null }));

jest.mock("@/lib/hooks/members", () => ({ useMembers: () => ({ data: [] }) }));
jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({ data: [] }),
}));
jest.mock("@/modules/objectives/hooks", () => ({
  useKeyResults: () => ({ data: [] }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/components/ui", () => ({ ObjectiveHealthIcon: () => null }));

it("retains OKR updates with Former user attribution when active members exclude the deleted account", () => {
  const activity = {
    userId: FORMER_USER_ID,
    field: "name",
    currentValue: "Launch",
    type: "create",
    updateType: "objective",
    objectiveId: "objective-1",
    createdAt: "2026-09-15T12:00:00Z",
    comment: "Shared progress explanation.",
  } as ObjectiveActivity;
  render(<ObjectiveActivityComponent {...activity} />);
  expect(screen.getAllByText("Former user").length).toBeGreaterThan(0);
  expect(screen.getByText("Shared progress explanation.")).toBeInTheDocument();
  expect(screen.queryByRole("link")).not.toBeInTheDocument();
});

it("renders standalone comments without treating them as a property change", () => {
  render(
    <ObjectiveActivityComponent
      {...({
        userId: FORMER_USER_ID,
        field: "comment",
        currentValue: "",
        type: "update",
        updateType: "objective",
        objectiveId: "objective-1",
        createdAt: "2026-09-15T12:00:00Z",
        comment: "**Ready** for review",
      } as ObjectiveActivity)}
    />,
  );
  expect(screen.getByText("commented")).toBeInTheDocument();
  expect(screen.getByTestId("markdown-comment")).toHaveTextContent(
    "**Ready** for review",
  );
  expect(screen.queryByText("changed the")).not.toBeInTheDocument();
});

it("retains an activity for an unknown field without crashing", () => {
  render(
    <ObjectiveActivityComponent
      {...({
        userId: FORMER_USER_ID,
        field: "future_field",
        currentValue: "Updated value",
        type: "update",
        updateType: "objective",
        objectiveId: "objective-1",
        createdAt: "2026-09-15T12:00:00Z",
      } as ObjectiveActivity)}
    />,
  );
  expect(screen.getByText("future_field")).toBeInTheDocument();
  expect(screen.getByText("Updated value")).toBeInTheDocument();
});
