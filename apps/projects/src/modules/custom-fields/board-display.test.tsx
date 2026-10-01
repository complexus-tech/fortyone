import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ApiError } from "api-client";
import type {
  CustomField,
  CustomFieldStoryValues,
  CustomFieldValue,
} from "./types";
import {
  CustomFieldsBoardProvider,
  StoryCustomFieldBadges,
} from "./board-display";
import { customFieldKeys } from "./hooks";

const mockReadFields = jest.fn();
const mockReadBatch = jest.fn();
const mockReadStory = jest.fn();
const mockWrite = jest.fn();
const mockReadPeople = jest.fn();
let mockRole: string | undefined = "member";

jest.mock("./api", () => ({
  getTeamCustomFields: (...args: unknown[]) => mockReadFields(...args),
  getCustomFieldStoryValues: (...args: unknown[]) => mockReadBatch(...args),
  getStoryCustomFields: (...args: unknown[]) => mockReadStory(...args),
  updateStoryCustomFields: (...args: unknown[]) => mockWrite(...args),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { token: "test" } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "First" }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockRole }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({ data: [{ id: "member-1", fullName: "Sam" }] }),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: (...args: unknown[]) => mockReadPeople(...args),
}));

const storyId = "0a11fba0-fde7-4669-9950-8b1bfb0b0bf5";
const siblingId = "15999f70-276c-41c1-9d4c-71b70b55b451";
const amount: CustomField = {
  id: "d350e64b-06cd-43e5-8bc8-66d756d54378",
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Deal amount",
  type: "money",
  icon: "work",
  currency: "USD",
  options: [],
  showOnCreate: true,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};
const score: CustomField = {
  ...amount,
  id: "979e97de-75f6-4b90-8945-a2bc348aa83c",
  name: "Score",
  type: "number",
  currency: null,
};
const note: CustomField = {
  ...score,
  id: "a7f782bc-5c7b-4af8-ad08-b596b5a6a7f4",
  name: "Note",
  type: "text",
};
let mockFields = [amount, score, note];
let mockItems: CustomFieldStoryValues["items"] = [];

const taskSnapshot = (id: string) => {
  const item = mockItems.find((candidate) => candidate.storyId === id)!;
  return { fields: mockFields, values: item.values, version: item.version };
};

const View = ({ disabled = false }: { disabled?: boolean }) => (
  <CustomFieldsBoardProvider
    selectedIds={mockFields.map((field) => field.id)}
    stories={[
      {
        id: storyId,
        teamId: amount.teamId,
        subStories: [{ id: siblingId, teamId: amount.teamId }],
      },
    ]}
  >
    <StoryCustomFieldBadges
      asList
      disabled={disabled}
      storyId={storyId}
      teamId={amount.teamId}
    />
    <StoryCustomFieldBadges storyId={siblingId} teamId={amount.teamId} />
  </CustomFieldsBoardProvider>
);

const renderView = (disabled = false) => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return {
    client,
    ...render(
      <QueryClientProvider client={client}>
        <View disabled={disabled} />
      </QueryClientProvider>,
    ),
  };
};

describe("custom field board/list editing", () => {
  beforeEach(() => {
    mockRole = "member";
    mockFields = [amount, score, note];
    mockItems = [
      {
        storyId,
        version: 7,
        values: [
          { fieldId: amount.id, value: "1250.50" },
          { fieldId: score.id, value: "0" },
          { fieldId: note.id, value: null },
        ],
      },
      {
        storyId: siblingId,
        version: 3,
        values: [
          { fieldId: amount.id, value: "249.50" },
          { fieldId: note.id, value: "" },
        ],
      },
    ];
    mockReadFields.mockReset().mockImplementation(async () => mockFields);
    mockReadBatch
      .mockReset()
      .mockImplementation(async () => ({ items: mockItems }));
    mockReadStory
      .mockReset()
      .mockImplementation(async (id: string) => taskSnapshot(id));
    mockReadPeople
      .mockReset()
      .mockReturnValue({ data: [], isPending: false, isError: false });
    mockWrite
      .mockReset()
      .mockImplementation(
        async (id: string, values: CustomFieldValue[], version: number) => {
          const current = mockItems.find((item) => item.storyId === id)!;
          if (version !== current.version)
            throw new ApiError("Value changed", 409, null);
          mockItems = mockItems.map((item) =>
            item.storyId === id
              ? {
                  ...item,
                  version: item.version + 1,
                  values: item.values.map(
                    (existing) =>
                      values.find(
                        (next) => next.fieldId === existing.fieldId,
                      ) ?? existing,
                  ),
                }
              : item,
          );
          return taskSnapshot(id);
        },
      );
  });

  it("renders only set icon/value chips, including zero, with one batch read for nested tasks", async () => {
    renderView();
    const chip = await screen.findByRole("button", {
      name: "Deal amount (USD): USD 1,250.50",
    });
    expect(chip).toHaveTextContent(/^USD 1,250.50$/);
    expect(chip).toHaveAttribute("title", "Deal amount (USD): USD 1,250.50");
    expect(chip).toHaveClass("h-[1.85rem]", "gap-1", "px-2", "rounded-xl");
    expect(chip.querySelector("svg")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Score: 0" })).toHaveTextContent(
      /^0$/,
    );
    expect(
      screen.getByRole("button", { name: "Deal amount (USD): USD 249.50" }),
    ).toHaveTextContent(/^USD 249.50$/);
    expect(screen.queryByText("Not set")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /Note/ }),
    ).not.toBeInTheDocument();
    expect(mockReadBatch).toHaveBeenCalledTimes(1);
    expect(mockReadBatch.mock.calls[0][0]).toEqual([storyId, siblingId]);
    expect(mockReadStory).not.toHaveBeenCalled();
    expect(mockReadPeople).not.toHaveBeenCalled();
  });

  it("saves one exact field/version, preserves siblings, and updates cards even when refresh fails", async () => {
    const { client } = renderView();
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Deal amount (USD): USD 1,250.50",
      }),
    );
    fireEvent.change(
      screen.getByRole("textbox", { name: "Deal amount (USD)" }),
      { target: { value: "9007199254740993.29" } },
    );
    mockReadBatch.mockRejectedValue(new Error("Refresh unavailable"));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    const saved = await screen.findByRole("button", {
      name: "Deal amount (USD): USD 9,007,199,254,740,993.29",
    });
    await waitFor(() => {
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    });
    expect(saved).toHaveTextContent(/^USD 9,007,199,254,740,993.29$/);
    expect(mockWrite).toHaveBeenCalledWith(
      storyId,
      [{ fieldId: amount.id, value: "9007199254740993.29" }],
      7,
      expect.anything(),
    );
    expect(
      screen.getByRole("button", { name: "Score: 0" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Deal amount (USD): USD 249.50" }),
    ).toBeInTheDocument();
    const batch = client.getQueryData<CustomFieldStoryValues>([
      ...customFieldKeys.batches("First"),
      [storyId, siblingId],
    ])!;
    expect(batch.items.find((item) => item.storyId === storyId)?.version).toBe(
      8,
    );
    expect(
      batch.items.find((item) => item.storyId === siblingId)?.version,
    ).toBe(3);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Custom fields could not be loaded.",
    );
  });

  it("keeps the opening version through a background refresh and reviews a cleared conflict without losing the draft", async () => {
    const { client } = renderView();
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Deal amount (USD): USD 1,250.50",
      }),
    );
    const input = screen.getByRole("textbox", { name: "Deal amount (USD)" });
    fireEvent.change(input, { target: { value: "1500.00" } });
    mockItems = mockItems.map((item) =>
      item.storyId === storyId
        ? {
            ...item,
            version: 8,
            values: item.values.map((value) =>
              value.fieldId === amount.id ? { ...value, value: null } : value,
            ),
          }
        : item,
    );
    act(() => {
      client.setQueryData(
        [...customFieldKeys.batches("First"), [storyId, siblingId]],
        { items: mockItems },
      );
    });
    client.setQueryData(customFieldKeys.story("First", storyId), {
      ...taskSnapshot(storyId),
      version: 7,
    });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await screen.findByRole("button", { name: "Review latest value" });
    expect(mockWrite.mock.calls[0][2]).toBe(7);
    expect(input).toHaveValue("1500.00");
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    fireEvent.click(
      screen.getByRole("button", { name: "Review latest value" }),
    );
    await screen.findByRole("status");
    expect(input).toHaveValue("1500.00");
    expect(screen.getByRole("status")).toHaveTextContent(
      "Latest saved value: Not set",
    );
    expect(
      client.getQueryState(customFieldKeys.story("First", storyId))
        ?.isInvalidated,
    ).toBe(true);
    expect(mockWrite).toHaveBeenCalledTimes(1);
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => {
      expect(mockWrite).toHaveBeenCalledTimes(2);
    });
    expect(mockWrite.mock.calls[1][2]).toBe(8);
    await screen.findByRole("button", {
      name: "Deal amount (USD): USD 1,500.00",
    });
  });

  it.each(["guest", undefined])(
    "keeps readable chips read-only for role %s",
    async (role) => {
      mockRole = role;
      renderView();
      const chip = await screen.findByRole("button", {
        name: "Deal amount (USD): USD 1,250.50",
      });
      expect(chip).toBeDisabled();
      fireEvent.click(chip);
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
      expect(mockWrite).not.toHaveBeenCalled();
    },
  );

  it("retains lifecycle and archived-schema read-only guards", async () => {
    mockFields = [{ ...amount, archivedAt: amount.createdAt }, score, note];
    renderView(true);
    const chip = await screen.findByRole("button", {
      name: /Deal amount.*USD 1,250.50/,
    });
    expect(chip).toBeDisabled();
    expect(screen.getByRole("button", { name: "Score: 0" })).toBeDisabled();
  });
});
