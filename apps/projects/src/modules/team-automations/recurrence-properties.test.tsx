import { fireEvent, render, screen } from "@testing-library/react";
import type { AutomationDraft } from "./types";
import { RecurrenceProperties } from "./recurrence-properties";

jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => ({
    data: [{ id: "status-backlog", name: "Backlog", teamId: "team-id" }],
  }),
  useStatuses: () => ({
    data: [{ id: "status-backlog", name: "Backlog", teamId: "team-id" }],
  }),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({
    data: [{ id: "member-id", fullName: "Alex", username: "alex" }],
  }),
}));
jest.mock("@/lib/hooks/labels", () => ({
  useLabels: () => ({
    data: [{ id: "label-id", name: "Operations", color: "#123456" }],
  }),
  useLabelsInfinite: () => ({
    data: { pages: [] },
    isPending: false,
    isFetching: false,
    isFetchingNextPage: false,
    hasNextPage: false,
  }),
}));
jest.mock("@/lib/hooks/create-label-mutation", () => ({
  useCreateLabelMutation: () => ({ mutateAsync: jest.fn() }),
}));
jest.mock("@/modules/teams/public/client", () => ({
  useTeamSettings: () => ({
    data: { estimationSettings: { scheme: "points" } },
  }),
}));

const DRAFT: AutomationDraft = {
  title: "Review operations",
  description: "Keep this description",
  descriptionHTML: "<p>Keep this description</p>",
  priority: "No Priority",
  checklist: ["Review open work"],
};

describe("recurring task properties", () => {
  it("leaves status unset until explicitly choosing even the first team status", () => {
    const onChange = jest.fn();
    render(
      <RecurrenceProperties
        disabled={false}
        draft={DRAFT}
        onChange={onChange}
        teamId="team-id"
      />,
    );
    expect(screen.getByRole("button", { name: "Status" })).toBeVisible();
    expect(onChange).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Status" }));
    fireEvent.click(screen.getByRole("option", { name: /Backlog/ }));

    expect(onChange).toHaveBeenCalledWith({ statusId: "status-backlog" });
    expect(DRAFT.statusId).toBeUndefined();
  });

  it("displays a template's native values without rewriting its draft", () => {
    const onChange = jest.fn();
    render(
      <RecurrenceProperties
        disabled={false}
        draft={{
          ...DRAFT,
          statusId: "status-backlog",
          assigneeId: "member-id",
          labelIds: ["label-id"],
          priority: "High",
          estimateValue: 3,
          estimatedDurationMinutes: 120,
          minimumFocusBlockMinutes: 60,
        }}
        onChange={onChange}
        teamId="team-id"
      />,
    );

    for (const name of [
      "Backlog",
      "High",
      "Operations",
      "alex",
      "3 points",
      "2 hours",
    ]) {
      expect(screen.getByRole("button", { name })).toBeVisible();
    }
    expect(onChange).not.toHaveBeenCalled();
  });

  it("updates duration and removes a focus block that no longer fits", () => {
    const onChange = jest.fn();
    render(
      <RecurrenceProperties
        disabled={false}
        draft={{
          ...DRAFT,
          estimatedDurationMinutes: 120,
          minimumFocusBlockMinutes: 60,
        }}
        onChange={onChange}
        teamId="team-id"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "2 hours" }));
    fireEvent.change(
      screen.getByRole("spinbutton", { name: "Custom time needed" }),
      {
        target: { value: "0.25" },
      },
    );
    fireEvent.click(screen.getByRole("button", { name: "Set" }));

    expect(onChange).toHaveBeenCalledWith({
      estimatedDurationMinutes: 15,
      minimumFocusBlockMinutes: undefined,
    });
  });

  it("prevents an already-open property menu from changing the draft while saving", () => {
    const onChange = jest.fn();
    const { rerender } = render(
      <RecurrenceProperties
        disabled={false}
        draft={DRAFT}
        onChange={onChange}
        teamId="team-id"
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Status" }));
    rerender(
      <RecurrenceProperties
        disabled
        draft={DRAFT}
        onChange={onChange}
        teamId="team-id"
      />,
    );
    for (const name of [
      "Status",
      "No Priority",
      "Labels",
      "Assignee",
      "Complexity",
      "Time needed",
    ]) {
      expect(screen.getByRole("button", { name })).toBeDisabled();
    }
    fireEvent.click(screen.getByRole("option", { name: /Backlog/ }));

    expect(onChange).not.toHaveBeenCalled();
  });
});
