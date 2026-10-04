import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { Preset, TaskTemplateConfiguration } from "./types";
import { CreationTemplatePicker } from "./creation-template-picker";

const mockQuery = jest.fn();
const mockFetchNextPage = jest.fn();
const mockRefetch = jest.fn();

jest.mock("./hooks", () => ({
  useWorkPresets: (...args: unknown[]) => mockQuery(...args),
}));

const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
const scrollIntoViewDescriptor = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollIntoView",
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
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: jest.fn(),
  });
});

afterAll(() => {
  if (resizeObserverDescriptor) {
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  } else {
    Reflect.deleteProperty(globalThis, "ResizeObserver");
  }
  if (scrollIntoViewDescriptor) {
    Object.defineProperty(
      HTMLElement.prototype,
      "scrollIntoView",
      scrollIntoViewDescriptor,
    );
  } else {
    Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
  }
});

const template: Preset<TaskTemplateConfiguration> = {
  id: "handoff-template",
  teamId: "team",
  ownerId: "owner",
  kind: "template",
  name: "Customer handoff",
  visibility: "team",
  canEdit: true,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
  configuration: {
    version: 1,
    title: "Call customer",
    description: "Keep the original configuration",
    descriptionHTML: "<p><strong>Keep the original configuration</strong></p>",
    priority: "High",
    checklist: ["Confirm delivery"],
  },
};
const secondTemplate: Preset<TaskTemplateConfiguration> = {
  ...template,
  id: "bug-template",
  name: "Bug report",
  visibility: "personal",
  configuration: { ...template.configuration, title: "Investigate the bug" },
};

const loaded = (pages: Preset[][]) => ({
  data: { pages: pages.map((items) => ({ items, nextCursor: "" })) },
  isPending: false,
  isError: false,
  isFetching: false,
  isFetchingNextPage: false,
  hasNextPage: false,
  fetchNextPage: mockFetchNextPage,
  refetch: mockRefetch,
});

beforeEach(() => {
  mockQuery.mockReset().mockReturnValue(loaded([[template, secondTemplate]]));
  mockFetchNextPage.mockReset().mockResolvedValue({});
  mockRefetch.mockReset().mockResolvedValue({});
});

describe("creation template picker", () => {
  it.each([
    { isPending: true, isError: false },
    { isPending: false, isError: true },
    loaded([[]]),
  ])(
    "prefetches and hides the chip before templates are available (%j)",
    (query) => {
      mockQuery.mockReturnValue(query);
      render(<CreationTemplatePicker onSelect={jest.fn()} teamId="team" />);
      expect(
        screen.queryByRole("button", { name: "Template" }),
      ).not.toBeInTheDocument();
      expect(mockQuery).toHaveBeenCalledWith("team", "template", true);
    },
  );

  it("searches without applying content, selects by keyboard, and shows the applied template name", async () => {
    const onSelect = jest.fn();
    render(<CreationTemplatePicker onSelect={onSelect} teamId="team" />);
    fireEvent.click(screen.getByRole("button", { name: "Template" }));
    const search = await screen.findByLabelText("Search templates");
    expect(search).toHaveFocus();
    expect(screen.queryByText("Team", { exact: true })).not.toBeInTheDocument();
    expect(
      screen.queryByText("Private", { exact: true }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /Manage/ }),
    ).not.toBeInTheDocument();
    fireEvent.change(search, { target: { value: "handoff" } });
    await waitFor(() => {
      expect(
        screen.queryByRole("option", { name: "Bug report" }),
      ).not.toBeInTheDocument();
    });
    expect(onSelect).not.toHaveBeenCalled();
    fireEvent.keyDown(search, { key: "Enter" });
    expect(onSelect).toHaveBeenCalledWith(template.configuration);
    expect(screen.queryByLabelText("Search templates")).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Customer handoff" }),
    ).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Customer handoff" }));
    expect(await screen.findByLabelText("Search templates")).toHaveValue("");
    expect(
      await screen.findByRole("option", { name: "Bug report" }),
    ).toBeVisible();
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it("resets selection and closes the menu when changing teams", async () => {
    const onSelect = jest.fn();
    const { rerender } = render(
      <CreationTemplatePicker onSelect={onSelect} teamId="team" />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Template" }));
    fireEvent.click(
      await screen.findByRole("option", { name: "Customer handoff" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Customer handoff" }));
    rerender(
      <CreationTemplatePicker onSelect={onSelect} teamId="other-team" />,
    );
    expect(screen.getByRole("button", { name: "Template" })).toBeVisible();
    expect(screen.queryByLabelText("Search templates")).not.toBeInTheDocument();
    rerender(<CreationTemplatePicker onSelect={onSelect} teamId="team" />);
    expect(screen.getByRole("button", { name: "Template" })).toBeVisible();
  });

  it("retains loaded templates and retries a later query failure", async () => {
    mockQuery.mockReturnValue({ ...loaded([[template]]), isError: true });
    render(<CreationTemplatePicker onSelect={jest.fn()} teamId="team" />);
    fireEvent.click(screen.getByRole("button", { name: "Template" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Could not load templates.",
    );
    expect(
      screen.getByRole("option", { name: "Customer handoff" }),
    ).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockRefetch).toHaveBeenCalledTimes(1);
  });

  it("loads additional pages and disables repeat fetches while they are pending", async () => {
    const onSelect = jest.fn();
    mockQuery.mockReturnValue({ ...loaded([[template]]), hasNextPage: true });
    const { rerender } = render(
      <CreationTemplatePicker onSelect={onSelect} teamId="team" />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Template" }));
    fireEvent.click(await screen.findByRole("button", { name: "Load more" }));
    expect(mockFetchNextPage).toHaveBeenCalledTimes(1);
    mockQuery.mockReturnValue({
      ...loaded([[template]]),
      hasNextPage: true,
      isFetchingNextPage: true,
    });
    rerender(<CreationTemplatePicker onSelect={onSelect} teamId="team" />);
    const loading = screen.getByRole("button", { name: "Loading..." });
    expect(loading).toBeDisabled();
    fireEvent.click(loading);
    expect(mockFetchNextPage).toHaveBeenCalledTimes(1);
    mockQuery.mockReturnValue(loaded([[template], [secondTemplate]]));
    rerender(<CreationTemplatePicker onSelect={onSelect} teamId="team" />);
    fireEvent.click(await screen.findByRole("option", { name: "Bug report" }));
    expect(onSelect).toHaveBeenCalledWith(secondTemplate.configuration);
  });

  it("prevents opening or choosing a template while disabled", async () => {
    const onSelect = jest.fn();
    const { rerender } = render(
      <CreationTemplatePicker disabled onSelect={onSelect} teamId="team" />,
    );
    const trigger = screen.getByRole("button", { name: "Template" });
    expect(trigger).toBeDisabled();
    fireEvent.click(trigger);
    expect(screen.queryByLabelText("Search templates")).not.toBeInTheDocument();
    rerender(<CreationTemplatePicker onSelect={onSelect} teamId="team" />);
    fireEvent.click(trigger);
    const option = await screen.findByRole("option", {
      name: "Customer handoff",
    });
    rerender(
      <CreationTemplatePicker disabled onSelect={onSelect} teamId="team" />,
    );
    expect(screen.getByLabelText("Search templates")).toBeDisabled();
    expect(option).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(option);
    expect(onSelect).not.toHaveBeenCalled();
  });
});
