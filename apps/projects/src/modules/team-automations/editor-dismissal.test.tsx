import type { Editor } from "@tiptap/core";
import type { ReactNode } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import type { useNewStoryDialogEditors } from "@/components/ui/use-new-story-dialog-editors";
import { RecurrenceEditor } from "./recurrence-editor";
import { RuleEditor } from "./rule-editor";

let mockTitleEditor: Editor | null = null;

jest.mock("./recurrence-properties", () => ({
  RecurrenceProperties: ({ children }: { children?: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("@/components/ui/use-new-story-dialog-editors", () => {
  const actual = jest.requireActual<{
    useNewStoryDialogEditors: typeof useNewStoryDialogEditors;
  }>("@/components/ui/use-new-story-dialog-editors");
  return {
    useNewStoryDialogEditors: (
      options: Parameters<typeof actual.useNewStoryDialogEditors>[0],
    ) => {
      const editors = actual.useNewStoryDialogEditors(options);
      mockTitleEditor = editors.titleEditor;
      return editors;
    },
  };
});
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
  CreationTemplatePicker: jest.fn(() => null),
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

const editors = [
  {
    name: "rule",
    Component: RuleEditor,
    createLabel: "Create workflow rule",
    prepare: async () => {
      fireEvent.change(screen.getByLabelText("Rule name"), {
        target: { value: "Route urgent work" },
      });
      fireEvent.change(screen.getByLabelText("Set priority"), {
        target: { value: "High" },
      });
    },
  },
  {
    name: "recurrence",
    Component: RecurrenceEditor,
    createLabel: "Create recurring task",
    prepare: async () => {
      await waitFor(() => {
        expect(mockTitleEditor).not.toBeNull();
        expect(
          screen.getByRole("textbox", { name: "Task title" }),
        ).toHaveFocus();
      });
      act(() => {
        mockTitleEditor!.commands.setContent("Weekly review");
      });
    },
  },
];

describe.each(editors)(
  "$name editor dismissal",
  ({ Component, createLabel, prepare }) => {
    it("keeps the dialog open and prevents duplicate creation while saving", async () => {
      let finish!: () => void;
      const save = jest.fn(
        () =>
          new Promise<void>((resolve) => {
            finish = resolve;
          }),
      );
      const close = jest.fn();
      render(<Component onClose={close} onSave={save} teamId="team-id" />);
      await prepare();
      fireEvent.click(screen.getByRole("button", { name: createLabel }));

      const dialog = screen.getByRole("dialog");
      expect(dialog).toHaveAttribute("aria-busy", "true");
      expect(
        screen.queryByRole("button", { name: "Close" }),
      ).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
      const creating = screen.getByRole("button", { name: "Creating..." });
      expect(creating).toBeDisabled();
      if (createLabel === "Create recurring task") {
        expect(
          screen.getByRole("textbox", { name: "Task title" }),
        ).toHaveAttribute("contenteditable", "false");
        expect(
          screen.getByRole("textbox", { name: "Description" }),
        ).toHaveAttribute("contenteditable", "false");
        expect(screen.getByLabelText("Time")).toBeDisabled();
      }
      fireEvent.click(creating);
      fireEvent.keyDown(dialog, { key: "Escape" });
      fireEvent.pointerDown(document.body, { button: 0, pointerType: "mouse" });
      expect(save).toHaveBeenCalledTimes(1);
      expect(close).not.toHaveBeenCalled();

      await act(async () => {
        finish();
      });
      expect(close).toHaveBeenCalledTimes(1);
    });

    it("restores dismissal and preserves the draft after a failed save", async () => {
      const save = jest.fn().mockRejectedValue(new Error("Connection lost"));
      const close = jest.fn();
      render(<Component onClose={close} onSave={save} teamId="team-id" />);
      await prepare();
      fireEvent.click(screen.getByRole("button", { name: createLabel }));

      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith(expect.any(String), {
          description: "Connection lost",
        });
        expect(screen.getByRole("dialog")).toHaveAttribute(
          "aria-busy",
          "false",
        );
      });
      expect(close).not.toHaveBeenCalled();
      expect(screen.getByRole("button", { name: createLabel })).toBeEnabled();
      fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
      expect(close).toHaveBeenCalledTimes(1);
    });
  },
);
