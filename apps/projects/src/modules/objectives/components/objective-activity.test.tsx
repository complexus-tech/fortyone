import { render, screen } from "@testing-library/react";
import { FORMER_USER_ID } from "@/lib/former-user";
import type { ObjectiveActivity } from "../types";
import { ObjectiveActivityComponent } from "./objective-activity";

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
