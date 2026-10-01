import { fireEvent, render, screen } from "@testing-library/react";
import { PresetPicker } from "./preset-picker";
import type { Preset } from "./types";

const mockQuery = jest.fn();
jest.mock("./hooks", () => ({
  useWorkPresets: (...args: unknown[]) => mockQuery(...args),
  usePresetMutations: () => ({ rename: {}, archive: {} }),
}));
jest.mock("./name-dialog", () => ({ PresetNameDialog: () => null }));
const preset: Preset = {
  id: "template",
  teamId: "team",
  ownerId: "owner",
  kind: "template",
  name: "Customer handoff",
  visibility: "team",
  canEdit: false,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
  configuration: {
    version: 1,
    title: "Call customer",
    description: "",
    descriptionHTML: "",
    priority: "No Priority",
    checklist: [],
  },
};
const loaded = (items: Preset[]) => ({
  data: { pages: [{ items }] },
  isPending: false,
  isError: false,
});
describe("creation template availability", () => {
  it.each([
    { isPending: true, isError: false },
    { isPending: false, isError: true },
    loaded([]),
  ])(
    "keeps the header action hidden until templates are available (%j)",
    (query) => {
      mockQuery.mockReturnValue(query);
      render(
        <PresetPicker
          hideWhenEmpty
          kind="template"
          label="Use template"
          onSelect={jest.fn()}
          teamId="team"
        />,
      );
      expect(
        screen.queryByRole("button", { name: "Use template" }),
      ).not.toBeInTheDocument();
      expect(mockQuery).toHaveBeenCalledWith("team", "template", true);
    },
  );
  it("opens the real chooser and selects an available team template", async () => {
    const onSelect = jest.fn();
    mockQuery.mockReturnValue(loaded([preset]));
    render(
      <PresetPicker
        hideWhenEmpty
        kind="template"
        label="Use template"
        onSelect={onSelect}
        teamId="team"
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Use template" }));
    fireEvent.click(
      await screen.findByRole("button", { name: "Customer handoff Team" }),
    );
    expect(onSelect).toHaveBeenCalledWith(preset);
  });
});
