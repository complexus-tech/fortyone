import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import type { CustomField } from "./types";
import { TeamCustomFieldSettings } from "./team-settings";

const FIELD: CustomField = {
  id: "d350e64b-06cd-43e5-8bc8-66d756d54378",
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Impact",
  type: "text",
  icon: null,
  currency: null,
  options: [],
  showOnCreate: false,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};
let mockFields: CustomField[] = [];
let mockPending = false;
const mockCreate = jest.fn();
const mockUpdate = jest.fn();

jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "admin" }),
}));
jest.mock("./hooks", () => ({
  useTeamCustomFields: () => ({
    data: mockFields,
    isPending: false,
    isError: false,
    refetch: jest.fn(),
  }),
  useCustomFieldMutations: () => ({
    create: { mutateAsync: mockCreate, isPending: mockPending },
    update: { mutateAsync: mockUpdate, isPending: mockPending },
    archive: { mutateAsync: jest.fn(), isPending: false, reset: jest.fn() },
  }),
}));

const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
beforeAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  });
});
afterAll(() => {
  if (resizeObserverDescriptor)
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
});
beforeEach(() => {
  jest.clearAllMocks();
  mockFields = [FIELD];
  mockPending = false;
  mockCreate.mockResolvedValue(FIELD);
  mockUpdate.mockResolvedValue(FIELD);
});

const openEditor = async (editing: boolean) => {
  const trigger = screen.getByRole("button", {
    name: editing ? "Actions for Impact" : "Create field",
  });
  trigger.focus();
  if (editing) {
    fireEvent.keyDown(trigger, { key: "Enter", code: "Enter" });
    const item = await screen.findByRole("menuitem", { name: "Edit field" });
    item.focus();
    fireEvent.keyDown(item, { key: "Enter", code: "Enter" });
  } else fireEvent.click(trigger);
  const dialog = await screen.findByRole("dialog", {
    name: editing ? "Edit field" : "Create field",
  });
  await waitFor(() => {
    expect(within(dialog).getByLabelText("Field name")).toHaveFocus();
  });
  return { trigger, dialog };
};

describe.each([false, true])("%s editing mode", (editing) => {
  it.each(["Cancel", "Escape", "Save"])(
    "returns focus to the initiating button after %s",
    async (dismissal) => {
      render(<TeamCustomFieldSettings teamId={FIELD.teamId} />);
      const { trigger, dialog } = await openEditor(editing);
      if (dismissal === "Escape")
        fireEvent.keyDown(dialog, { key: "Escape", code: "Escape" });
      else if (dismissal === "Cancel")
        fireEvent.click(within(dialog).getByRole("button", { name: "Cancel" }));
      else {
        fireEvent.change(within(dialog).getByLabelText("Field name"), {
          target: { value: "Impact" },
        });
        fireEvent.click(
          within(dialog).getByRole("button", {
            name: editing ? "Save field" : "Create field",
          }),
        );
      }
      await waitFor(() => {
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
        expect(screen.queryByRole("menu")).not.toBeInTheDocument();
        expect(trigger).toHaveFocus();
      });
      if (dismissal === "Save") {
        expect(editing ? mockUpdate : mockCreate).toHaveBeenCalledTimes(1);
        if (editing)
          expect(mockUpdate).toHaveBeenCalledWith({
            fieldId: FIELD.id,
            input: {
              name: "Impact",
              icon: null,
              options: [],
              showOnCreate: false,
            },
          });
        else
          expect(mockCreate).toHaveBeenCalledWith({
            name: "Impact",
            type: "text",
            icon: null,
            currency: null,
            options: [],
            showOnCreate: false,
          });
      } else {
        expect(mockCreate).not.toHaveBeenCalled();
        expect(mockUpdate).not.toHaveBeenCalled();
      }
    },
  );
});

it("keeps the editor open while pending and restores focus once it can close", async () => {
  const { rerender } = render(
    <TeamCustomFieldSettings teamId={FIELD.teamId} />,
  );
  const { trigger, dialog } = await openEditor(true);
  mockPending = true;
  rerender(<TeamCustomFieldSettings teamId={FIELD.teamId} />);
  fireEvent.keyDown(dialog, { key: "Escape", code: "Escape" });
  expect(dialog).toBeInTheDocument();
  expect(within(dialog).getByRole("button", { name: "Cancel" })).toBeDisabled();
  expect(within(dialog).getByLabelText("Field name")).toBeDisabled();
  mockPending = false;
  rerender(<TeamCustomFieldSettings teamId={FIELD.teamId} />);
  fireEvent.keyDown(dialog, { key: "Escape", code: "Escape" });
  await waitFor(() => {
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });
});

it("returns focus to the settings section when reaching the limit disables Create", async () => {
  mockFields = Array.from({ length: 49 }, (_, index) => ({
    ...FIELD,
    id: `00000000-0000-4000-8000-${String(index).padStart(12, "0")}`,
    name: `Field ${index}`,
  }));
  mockCreate.mockImplementation(async () => {
    mockFields = [...mockFields, FIELD];
    return FIELD;
  });
  render(<TeamCustomFieldSettings teamId={FIELD.teamId} />);
  const { trigger, dialog } = await openEditor(false);
  fireEvent.change(within(dialog).getByLabelText("Field name"), {
    target: { value: "Impact" },
  });
  fireEvent.click(within(dialog).getByRole("button", { name: "Create field" }));
  await waitFor(() => {
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toBeDisabled();
    expect(screen.getByRole("region", { name: "Custom fields" })).toHaveFocus();
  });
});
