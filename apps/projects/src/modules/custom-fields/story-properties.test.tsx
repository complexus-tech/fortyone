import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { CustomField } from "./types";
import { StoryCustomFieldProperties } from "./story-properties";

const mockSnapshot = jest.fn();
const mockSave = jest.fn();
const mockReset = jest.fn();
jest.mock("./hooks", () => ({
  useStoryCustomFields: (...args: unknown[]) => mockSnapshot(...args),
  useUpdateStoryCustomFields: () => ({
    mutateAsync: mockSave,
    reset: mockReset,
    isPending: false,
    error: null,
  }),
}));
jest.mock("@/hooks/media", () => ({ useMediaQuery: () => false }));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [], isPending: false, isError: false }),
}));

const amount: CustomField = {
  id: "d350e64b-06cd-43e5-8bc8-66d756d54378",
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Contract value",
  type: "money",
  icon: "star",
  currency: "USD",
  options: [],
  showOnCreate: true,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};
const snapshot = (field = amount, value: string | null = "0") => ({
  data: { fields: [field], values: [{ fieldId: field.id, value }], version: 7 },
  isPending: false,
  isError: false,
  refetch: jest.fn(),
});

describe("task custom field properties", () => {
  beforeEach(() => {
    mockSnapshot.mockReturnValue(snapshot());
    mockSave.mockReset().mockResolvedValue(undefined);
  });
  it("uses the native label layout and saves an exact amount with its version", async () => {
    render(<StoryCustomFieldProperties storyId="story" />);
    expect(screen.getByText("Contract value (USD)").parentElement).toHaveClass(
      "grid-cols-[7.875rem_auto]",
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Contract value (USD): USD 0.00" }),
    );
    const input = await screen.findByRole("textbox", {
      name: "Contract value (USD)",
    });
    fireEvent.change(input, { target: { value: "1,250.50" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    const error = await screen.findByRole("alert");
    expect(input).toHaveAttribute("aria-describedby", error.id);
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(mockSave).not.toHaveBeenCalled();
    fireEvent.change(input, { target: { value: "9007199254740993.29" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(mockSave).toHaveBeenCalledWith({
        values: [{ fieldId: amount.id, value: "9007199254740993.29" }],
        expectedVersion: 7,
      });
    });
  });
  it("uses value-only compact chips and keeps empty values behind disclosure", () => {
    mockSnapshot.mockReturnValue(snapshot(amount, null));
    render(<StoryCustomFieldProperties isCompact storyId="story" />);
    expect(
      screen.queryByRole("button", { name: "Contract value (USD): Not set" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Add details (1)" }));
    const chip = screen.getByRole("button", {
      name: "Contract value (USD): Not set",
    });
    expect(chip).toHaveTextContent(/^Not set$/);
    expect(chip).toHaveAttribute("title", "Contract value (USD): Not set");
    expect(chip.closest(".grid")).toBeNull();
    expect(
      screen.getByRole("button", { name: "Hide empty fields" }),
    ).toBeInTheDocument();
  });
  it("keeps an unsaved draft mounted when cached fields fail to refresh", async () => {
    const cachedSnapshot = snapshot();
    mockSnapshot.mockReturnValue(cachedSnapshot);
    const { rerender } = render(<StoryCustomFieldProperties storyId="story" />);
    fireEvent.click(
      screen.getByRole("button", { name: "Contract value (USD): USD 0.00" }),
    );
    const input = await screen.findByRole("textbox", {
      name: "Contract value (USD)",
    });
    fireEvent.change(input, { target: { value: "1250.50" } });

    mockSnapshot.mockReturnValue({ ...cachedSnapshot, isError: true });
    rerender(<StoryCustomFieldProperties storyId="story" />);

    expect(screen.getByRole("textbox")).toBe(input);
    expect(input).toHaveValue("1250.50");
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Fields could not be loaded.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(cachedSnapshot.refetch).toHaveBeenCalledTimes(1);
  });
  it("resets the editor and opening version when switching cached tasks", async () => {
    const { rerender } = render(
      <StoryCustomFieldProperties storyId="first-story" />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Contract value (USD): USD 0.00" }),
    );
    fireEvent.change(
      await screen.findByRole("textbox", { name: "Contract value (USD)" }),
      { target: { value: "1250.50" } },
    );

    const nextSnapshot = snapshot(amount, "249.50");
    nextSnapshot.data.version = 12;
    mockSnapshot.mockReturnValue(nextSnapshot);
    rerender(<StoryCustomFieldProperties storyId="second-story" />);
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Contract value (USD): USD 249.50" }),
    );
    const input = await screen.findByRole("textbox", {
      name: "Contract value (USD)",
    });
    expect(input).toHaveValue("249.50");
    fireEvent.change(input, { target: { value: "250.50" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(mockSave).toHaveBeenCalledWith({
        values: [{ fieldId: amount.id, value: "250.50" }],
        expectedVersion: 12,
      });
    });
  });
  it.each([true, false])(
    "retains readable values without editing archived or disabled work (%s)",
    (archived) => {
      mockSnapshot.mockReturnValue(
        snapshot({ ...amount, archivedAt: archived ? amount.createdAt : null }),
      );
      render(
        <StoryCustomFieldProperties
          disabled={!archived}
          isNotifications
          storyId="story"
        />,
      );
      const chip = screen.getByRole("button", {
        name: /Contract value \(USD\).*USD 0.00/,
      });
      expect(chip).toBeDisabled();
      expect(chip).toHaveTextContent(/^USD 0.00$/);
      fireEvent.click(chip);
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    },
  );
});
