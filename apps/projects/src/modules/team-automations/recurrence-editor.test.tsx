import type { Editor } from "@tiptap/core";
import type { ReactNode, useState as useReactState } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import type { useNewStoryDialogEditors } from "@/components/ui/use-new-story-dialog-editors";
import type { TaskTemplateConfiguration } from "@/modules/work-presets/public/types";
import type { AutomationInput } from "./types";
import { RecurrenceEditor } from "./recurrence-editor";

let mockTitleEditor: Editor | null = null;
let mockDescriptionEditor: Editor | null = null;
const mockTemplate: TaskTemplateConfiguration = {
  version: 1,
  title: "Review <release>",
  description: "Review the release",
  descriptionHTML:
    '<h2>Release review</h2><p><strong>Keep this formatting</strong></p><ul data-type="taskList"><li data-type="taskItem" data-checked="true"><p>Existing item</p></li></ul>',
  priority: "High",
  statusId: "status-id",
  assigneeId: "member-id",
  labelIds: ["label-id"],
  estimateValue: 3,
  estimatedDurationMinutes: 45,
  minimumFocusBlockMinutes: 15,
  checklist: ["Check <logs>"],
  customFieldValues: [
    { fieldId: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", value: "5" },
  ],
};

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
      mockDescriptionEditor = editors.descriptionEditor;
      return editors;
    },
  };
});
jest.mock("./recurrence-properties", () => ({
  RecurrenceProperties: ({ children }: { children?: ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("@/modules/custom-fields/public/creation", () => {
  const { useState } = jest.requireActual<{ useState: typeof useReactState }>(
    "react",
  );
  return {
    useCreateCustomFields: () => {
      const [values, setValues] = useState<Record<string, unknown>>({});
      return {
        values,
        setValues,
        prepareValues: () =>
          Object.keys(values).length
            ? Object.entries(values).map(([fieldId, value]) => ({
                fieldId,
                value,
              }))
            : [
                {
                  fieldId: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
                  value: null,
                },
              ],
        isPending: false,
        isError: false,
      };
    },
    CreateCustomFields: () => null,
  };
});
jest.mock("@/modules/work-presets/public/template-picker", () => ({
  CreationTemplatePicker: ({
    onSelect,
  }: {
    onSelect: (template: TaskTemplateConfiguration) => void;
  }) => (
    <button
      onClick={() => {
        onSelect(mockTemplate);
      }}
      type="button"
    >
      Use task template
    </button>
  ),
}));

const readyEditors = async () => {
  await waitFor(() => {
    expect(mockTitleEditor).not.toBeNull();
    expect(mockDescriptionEditor).not.toBeNull();
  });
  return {
    title: mockTitleEditor!,
    description: mockDescriptionEditor!,
  };
};

beforeEach(() => {
  mockTitleEditor = null;
  mockDescriptionEditor = null;
});

it("saves rich description content with calendar timezone and atomic custom fields", async () => {
  const save = jest
    .fn<Promise<unknown>, [AutomationInput]>()
    .mockResolvedValue({});
  const close = jest.fn();
  render(<RecurrenceEditor onClose={close} onSave={save} teamId="team-id" />);
  const editors = await readyEditors();
  act(() => {
    editors.title.commands.setContent("Weekly review");
    editors.description.commands.setContent(
      "<p><strong>Review &lt;logs&gt;</strong></p><p>Then follow up</p>",
    );
  });
  expect(screen.getByRole("textbox", { name: "Task title" })).toBeVisible();
  expect(screen.getByRole("textbox", { name: "Description" })).toBeVisible();
  fireEvent.change(screen.getByLabelText("Timezone"), {
    target: { value: "Africa/Harare" },
  });
  fireEvent.change(screen.getByLabelText("Time"), {
    target: { value: "11:30" },
  });
  fireEvent.click(
    screen.getByRole("button", { name: "Create recurring task" }),
  );
  await waitFor(() => {
    expect(save).toHaveBeenCalledTimes(1);
  });
  expect(save.mock.calls[0][0]).toMatchObject({
    teamId: "team-id",
    kind: "recurrence",
    name: "Weekly review",
    configuration: {
      version: 1,
      schedule: {
        frequency: "weekly",
        timezone: "Africa/Harare",
        localTime: "11:30",
      },
      draft: {
        title: "Weekly review",
        description: "Review <logs>\n\nThen follow up",
        descriptionHTML:
          "<p><strong>Review &lt;logs&gt;</strong></p><p>Then follow up</p>",
        customFieldValues: [
          { fieldId: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", value: null },
        ],
      },
    },
  });
  await waitFor(() => {
    expect(close).toHaveBeenCalled();
  });
});

it("keeps template formatting, unchecked checklist items, native properties and custom values", async () => {
  const save = jest
    .fn<Promise<unknown>, [AutomationInput]>()
    .mockResolvedValue({});
  render(
    <RecurrenceEditor onClose={jest.fn()} onSave={save} teamId="team-id" />,
  );
  await readyEditors();
  fireEvent.click(screen.getByRole("button", { name: "Use task template" }));
  expect(screen.getByRole("textbox", { name: "Task title" })).toHaveTextContent(
    "Review <release>",
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Create recurring task" }),
  );
  await waitFor(() => {
    expect(save).toHaveBeenCalledTimes(1);
  });
  const configuration = save.mock.calls[0][0].configuration;
  if (!("draft" in configuration))
    throw new Error("Expected a recurring draft");
  const { draft } = configuration;
  expect(draft).toMatchObject({
    title: mockTemplate.title,
    priority: "High",
    statusId: "status-id",
    assigneeId: "member-id",
    labelIds: ["label-id"],
    estimateValue: 3,
    estimatedDurationMinutes: 45,
    minimumFocusBlockMinutes: 15,
    customFieldValues: mockTemplate.customFieldValues,
    checklist: [],
  });
  const parsed = new DOMParser().parseFromString(
    draft.descriptionHTML,
    "text/html",
  );
  expect(parsed.querySelector("h2")?.textContent).toBe("Release review");
  expect(parsed.querySelector("strong")?.textContent).toBe(
    "Keep this formatting",
  );
  expect(parsed.querySelectorAll('[data-type="taskItem"]')).toHaveLength(2);
  expect(parsed.querySelectorAll('[data-checked="true"]')).toHaveLength(0);
  expect(parsed.body.textContent.match(/Check <logs>/g)).toHaveLength(1);
});

it("shows a validation error and prevents saving an oversized editor title", async () => {
  const save = jest.fn();
  render(
    <RecurrenceEditor onClose={jest.fn()} onSave={save} teamId="team-id" />,
  );
  const { title } = await readyEditors();
  act(() => {
    title.commands.setContent("x".repeat(256));
  });
  expect(screen.getByRole("alert")).toHaveTextContent("255 characters");
  expect(
    screen.getByRole("button", { name: "Create recurring task" }),
  ).toBeDisabled();
  expect(save).not.toHaveBeenCalled();
});

it("requires a valid timezone and preserves an empty description and optional schedule name", async () => {
  const save = jest
    .fn<Promise<unknown>, [AutomationInput]>()
    .mockResolvedValue({});
  render(
    <RecurrenceEditor onClose={jest.fn()} onSave={save} teamId="team-id" />,
  );
  const { title } = await readyEditors();
  act(() => {
    title.commands.setContent("Weekly review");
  });
  fireEvent.change(screen.getByLabelText("Timezone"), {
    target: { value: "Unknown/Timezone" },
  });
  expect(screen.getByRole("alert")).toHaveTextContent("Enter a valid timezone");
  expect(
    screen.getByRole("button", { name: "Create recurring task" }),
  ).toBeDisabled();
  fireEvent.change(screen.getByLabelText("Timezone"), {
    target: { value: "+02:00" },
  });
  expect(screen.getByRole("alert")).toHaveTextContent("Enter a valid timezone");
  expect(
    screen.getByRole("button", { name: "Create recurring task" }),
  ).toBeDisabled();
  expect(screen.getByLabelText("Timezone")).toBeVisible();
  fireEvent.change(screen.getByLabelText("Timezone"), {
    target: { value: " africa/harare " },
  });
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  fireEvent.change(screen.getByLabelText("Schedule name (optional)"), {
    target: { value: "Release routine" },
  });
  fireEvent.click(
    screen.getByRole("button", { name: "Create recurring task" }),
  );
  await waitFor(() => {
    expect(save).toHaveBeenCalledTimes(1);
  });
  expect(save.mock.calls[0][0]).toMatchObject({
    name: "Release routine",
    configuration: {
      schedule: { timezone: "Africa/Harare" },
      draft: { description: "", descriptionHTML: "" },
    },
  });
});
