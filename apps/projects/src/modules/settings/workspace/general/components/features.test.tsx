import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { WorkspaceFeatures } from "./features";

const mockUpdateSettings = jest.fn();
let mockSaving = false;
const mockFeatures = {
  objectiveEnabled: true,
  keyResultEnabled: true,
  chatEnabled: true,
  isPending: false,
  isError: false,
};
jest.mock("@/hooks", () => ({
  useFeatures: () => mockFeatures,
  useTerminology: () => ({
    getTermDisplay: (term: string) =>
      term === "objectiveTerm" ? "Goals" : "Milestones",
  }),
}));
jest.mock("@/lib/hooks/workspace/update-settings", () => ({
  useUpdateWorkspaceSettingsMutation: () => ({
    mutate: mockUpdateSettings,
    isPending: mockSaving,
  }),
}));
jest.mock("@/modules/settings/components", () => ({
  SectionHeader: () => <h2>Features</h2>,
}));
jest.mock("@/components/ui", () => ({
  RowWrapper: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  Switch: ({
    checked,
    disabled,
    onCheckedChange,
    "aria-label": label,
  }: {
    checked: boolean;
    disabled: boolean;
    onCheckedChange: () => void;
    "aria-label": string;
  }) => (
    <button
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={onCheckedChange}
      role="switch"
      type="button"
    />
  ),
}));

beforeEach(() => {
  jest.clearAllMocks();
  mockSaving = false;
  Object.assign(mockFeatures, {
    chatEnabled: true,
    isPending: false,
    isError: false,
  });
});

it("shows Chat enabled alongside the workspace's existing feature switches", () => {
  const view = render(<WorkspaceFeatures />);
  expect(screen.getByRole("switch", { name: "Goals" })).toBeInTheDocument();
  expect(
    screen.getByRole("switch", { name: "Milestones" }),
  ).toBeInTheDocument();
  const chat = screen.getByRole("switch", { name: "Chat" });
  expect(chat).toHaveAttribute("aria-checked", "true");
  fireEvent.click(chat);
  expect(mockUpdateSettings).toHaveBeenLastCalledWith({ chatEnabled: false });
  mockFeatures.chatEnabled = false;
  view.rerender(<WorkspaceFeatures />);
  expect(chat).toHaveAttribute("aria-checked", "false");
  fireEvent.click(chat);
  expect(mockUpdateSettings).toHaveBeenLastCalledWith({ chatEnabled: true });
});

it.each(["loading", "error", "saving"])(
  "prevents toggles while settings are %s",
  (state) => {
    mockFeatures.isPending = state === "loading";
    mockFeatures.isError = state === "error";
    mockSaving = state === "saving";
    render(<WorkspaceFeatures />);
    for (const toggle of screen.getAllByRole("switch")) {
      expect(toggle).toBeDisabled();
      fireEvent.click(toggle);
    }
    expect(mockUpdateSettings).not.toHaveBeenCalled();
  },
);
