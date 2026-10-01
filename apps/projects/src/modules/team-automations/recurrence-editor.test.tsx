import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { RecurrenceEditor } from "./recurrence-editor";

jest.mock("@/modules/custom-fields/public/creation", () => ({
  useCreateCustomFields: () => ({
    values: {},
    setValues: jest.fn(),
    prepareValues: () => [
      { fieldId: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", value: null },
    ],
    isPending: false,
    isError: false,
  }),
  CreateCustomFields: () => null,
}));
jest.mock("@/modules/work-presets/public/template-picker", () => ({
  PresetPicker: () => null,
}));

it("saves a recurring draft with calendar timezone and atomic custom fields", async () => {
  const save = jest.fn().mockResolvedValue({});
  const close = jest.fn();
  render(<RecurrenceEditor onClose={close} onSave={save} teamId="team-id" />);
  fireEvent.change(screen.getByLabelText("Task title"), {
    target: { value: "Weekly review" },
  });
  fireEvent.change(screen.getByLabelText("Description"), {
    target: { value: "Review <logs>\nThen follow up" },
  });
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
        descriptionHTML: "<p>Review &lt;logs&gt;</p><p>Then follow up</p>",
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
