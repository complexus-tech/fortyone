import type { ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { TeamAutomations } from "./index";

let mockRole: "admin" | "member" | "guest" | undefined = "member";
let mockCreationPending = false;
let mockLoadError = false;
const mockCreate = jest.fn();
const mockRefetch = jest.fn();

jest.mock("./recurrence-properties", () => ({
  RecurrenceProperties: ({ children }: { children?: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockRole }),
}));
jest.mock("./hooks", () => ({
  useAutomations: () => ({
    data: [],
    isPending: false,
    isError: mockLoadError,
    refetch: mockRefetch,
    create: { mutateAsync: mockCreate, isPending: mockCreationPending },
    pause: { mutateAsync: jest.fn(), isPending: false },
    archive: { mutateAsync: jest.fn(), isPending: false },
  }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => ({ data: [], isPending: false, isError: false }),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [], isPending: false, isError: false }),
}));
jest.mock("@/modules/custom-fields/public/creation", () => ({
  useCreateCustomFields: () => ({
    values: {},
    setValues: jest.fn(),
    prepareValues: () => [],
    isPending: false,
    isError: false,
  }),
  CreateCustomFields: () => null,
}));
jest.mock("@/modules/work-presets/public/template-picker", () => ({
  CreationTemplatePicker: () => null,
}));
jest.mock("./select-field", () => ({
  AutomationSelect: ({
    label,
    value,
    onChange,
    options,
  }: {
    label: string;
    value: string;
    onChange: (value: string) => void;
    options: { value: string; label: string }[];
  }) => (
    <label>
      {label}
      <select
        onChange={(event) => {
          onChange(event.target.value);
        }}
        value={value}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  ),
}));

const openCreateMenu = async () => {
  const trigger = screen.getByRole("button", { name: "Create automation" });
  fireEvent.keyDown(trigger, { key: "Enter" });
  return screen.findByRole("menu");
};

describe("automation header creation menu", () => {
  beforeEach(() => {
    mockRole = "member";
    mockCreationPending = false;
    mockLoadError = false;
    mockCreate.mockReset().mockResolvedValue({});
    mockRefetch.mockReset();
  });

  it("opens both real editors by keyboard and creates the chosen rule for its team", async () => {
    render(<TeamAutomations teamId="team-id" />);
    expect(
      screen.queryByRole("button", { name: "Create workflow rule" }),
    ).toBeNull();
    expect(
      screen.queryByRole("button", { name: "Create recurring task" }),
    ).toBeNull();

    await openCreateMenu();
    fireEvent.keyDown(screen.getByRole("menuitem", { name: "Workflow rule" }), {
      key: "Enter",
    });
    const rule = await screen.findByRole("dialog", {
      name: "Create workflow rule",
    });
    expect(screen.queryByRole("menu")).toBeNull();
    expect(rule).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole("textbox", { name: "Rule name" })).toHaveFocus();
    });
    fireEvent.change(screen.getByRole("textbox", { name: "Rule name" }), {
      target: { value: "Route urgent work" },
    });
    fireEvent.change(screen.getByLabelText("Set priority"), {
      target: { value: "High" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Create workflow rule" }),
    );
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          teamId: "team-id",
          kind: "rule",
          name: "Route urgent work",
          configuration: expect.objectContaining({
            actions: { priority: "High" },
          }),
        }),
      );
      expect(screen.queryByRole("dialog")).toBeNull();
      expect(
        screen.getByRole("button", { name: "Create automation" }),
      ).toHaveFocus();
      expect(document.body.style.pointerEvents).not.toBe("none");
    });

    await openCreateMenu();
    fireEvent.keyDown(
      screen.getByRole("menuitem", { name: "Recurring task" }),
      { key: "Enter" },
    );
    expect(
      await screen.findByRole("dialog", { name: "Create recurring task" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Task title" })).toBeVisible();
    await waitFor(() => {
      expect(screen.getByRole("textbox", { name: "Task title" })).toHaveFocus();
    });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
      expect(screen.queryByRole("menu")).toBeNull();
      expect(
        screen.getByRole("button", { name: "Create automation" }),
      ).toHaveFocus();
      expect(document.body.style.pointerEvents).not.toBe("none");
    });
    expect(mockCreate).toHaveBeenCalledTimes(1);
  });

  it.each(["guest", undefined] as const)(
    "prevents creation for a %s role",
    (role) => {
      mockRole = role;
      render(<TeamAutomations teamId="team-id" />);
      const trigger = screen.getByRole("button", { name: "Create automation" });
      expect(trigger).toBeDisabled();
      fireEvent.keyDown(trigger, { key: "Enter" });
      expect(screen.queryByRole("menu")).toBeNull();
      expect(screen.queryByRole("dialog")).toBeNull();
      expect(mockCreate).not.toHaveBeenCalled();
    },
  );

  it("keeps creation disabled while a previous save is pending", () => {
    mockRole = "admin";
    mockCreationPending = true;
    render(<TeamAutomations teamId="team-id" />);
    expect(
      screen.getByRole("button", { name: "Create automation" }),
    ).toBeDisabled();
  });

  it("moves mouse focus into the editor and restores it after dismissal", async () => {
    render(<TeamAutomations teamId="team-id" />);
    const trigger = screen.getByRole("button", { name: "Create automation" });
    trigger.focus();
    fireEvent.pointerDown(trigger, {
      button: 0,
      ctrlKey: false,
      pointerType: "mouse",
    });
    fireEvent.click(
      await screen.findByRole("menuitem", { name: "Workflow rule" }),
    );

    await screen.findByRole("dialog", { name: "Create workflow rule" });
    await waitFor(() => {
      expect(screen.getByRole("textbox", { name: "Rule name" })).toHaveFocus();
    });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
      expect(trigger).toHaveFocus();
      expect(document.body.style.pointerEvents).not.toBe("none");
    });
  });

  it("restores trigger focus when the menu closes without opening an editor", async () => {
    render(<TeamAutomations teamId="team-id" />);
    const trigger = screen.getByRole("button", { name: "Create automation" });
    trigger.focus();
    const menu = await openCreateMenu();
    fireEvent.keyDown(menu, { key: "Escape" });

    await waitFor(() => {
      expect(screen.queryByRole("menu")).toBeNull();
      expect(trigger).toHaveFocus();
      expect(screen.queryByRole("dialog")).toBeNull();
    });
  });

  it("preserves the list error and retry action beside the header menu", () => {
    mockLoadError = true;
    render(<TeamAutomations teamId="team-id" />);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Could not load automations.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockRefetch).toHaveBeenCalledTimes(1);
    expect(
      screen.getByRole("button", { name: "Create automation" }),
    ).toBeEnabled();
  });
});
