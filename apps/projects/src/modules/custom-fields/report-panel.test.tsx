import { fireEvent, render, screen, within } from "@testing-library/react";
import { CustomFieldReportPanel } from "./report-panel";
import type { CustomField, CustomFieldReport } from "./types";

const mockTeams = jest.fn();
const mockStatuses = jest.fn();
const mockPeople = jest.fn();
const mockFields = jest.fn();
const mockReport = jest.fn();
const mockMutate = jest.fn();
const mockReset = jest.fn();
jest.mock("@/modules/teams/public/client", () => ({
  useJoinedTeams: () => mockTeams(),
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => mockStatuses(),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => mockPeople(),
}));
jest.mock("./hooks", () => ({
  useTeamCustomFields: () => mockFields(),
  useCustomFieldReport: () => mockReport(),
}));

const amount: CustomField = {
  id: "d350e64b-06cd-43e5-8bc8-66d756d54378",
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Contract value",
  type: "money",
  currency: "USD",
  options: [],
  showOnCreate: true,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};
const initialInput = {
  fieldId: amount.id,
  aggregation: "sum",
  groupBy: "status",
  dateBasis: "created",
  statusIds: [],
  assigneeIds: [],
};
const report: CustomFieldReport = {
  field: amount,
  aggregation: "sum",
  groupBy: "status",
  currency: "USD",
  totalCount: 3,
  valuedCount: 2,
  missingCount: 1,
  rows: [{ key: "won", label: "Won", value: "9007199254740993.29", count: 2 }],
};
const query = (data: unknown[]) => ({
  data,
  isPending: false,
  isError: false,
  refetch: jest.fn(),
});

describe("custom field reports", () => {
  beforeEach(() => {
    mockTeams.mockReturnValue(query([{ id: amount.teamId, name: "Sales" }]));
    mockStatuses.mockReturnValue(query([{ id: "won", name: "Won" }]));
    mockPeople.mockReturnValue(query([]));
    mockFields.mockReturnValue(query([amount]));
    mockMutate.mockReset();
    mockReport.mockReturnValue({
      mutate: mockMutate,
      reset: mockReset,
      isPending: false,
      error: null,
    });
  });
  it("keeps field selection controlled while definitions load", () => {
    const warn = jest.spyOn(console, "warn").mockImplementation(() => {});
    try {
      mockFields.mockReturnValue({ ...query([]), isPending: true });
      const { rerender } = render(<CustomFieldReportPanel />);
      expect(
        screen.getByRole("combobox", { name: "Report field" }),
      ).toBeDisabled();

      mockFields.mockReturnValue(query([amount]));
      rerender(<CustomFieldReportPanel />);
      expect(
        screen.getByRole("combobox", { name: "Report field" }),
      ).toHaveTextContent("Contract value");
      expect(warn).not.toHaveBeenCalled();
    } finally {
      warn.mockRestore();
    }
  });
  it("keeps dates and people collapsed, validates dates and sends selected configured statuses", () => {
    render(<CustomFieldReportPanel />);
    expect(
      screen.getByRole("combobox", { name: "Report field" }),
    ).toBeInTheDocument();
    expect(screen.queryByLabelText("From")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Filters" }));
    const from = screen.getByLabelText("From");
    const through = screen.getByLabelText("Through");
    expect(from).toHaveClass("h-11");
    fireEvent.change(from, { target: { value: "2026-10-02" } });
    fireEvent.change(through, { target: { value: "2026-10-01" } });
    fireEvent.click(screen.getByRole("button", { name: "Build report" }));
    expect(screen.getByRole("alert")).toHaveTextContent("end date");
    expect(mockMutate).not.toHaveBeenCalled();
    fireEvent.change(through, { target: { value: "2026-10-03" } });
    fireEvent.click(screen.getByRole("checkbox", { name: "Won" }));
    fireEvent.click(screen.getByRole("button", { name: "Build report" }));
    expect(mockMutate).toHaveBeenCalledWith({
      ...initialInput,
      statusIds: ["won"],
      startDate: "2026-10-02",
      endDate: "2026-10-03",
    });
    fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(from).toHaveValue("");
    expect(screen.getByRole("checkbox", { name: "Won" })).not.toBeChecked();
  });
  it("displays exact money and missing coverage, then hides stale results when filters change", () => {
    mockReport.mockReturnValue({
      mutate: mockMutate,
      reset: mockReset,
      isPending: false,
      error: null,
      data: report,
      variables: initialInput,
    });
    render(<CustomFieldReportPanel />);
    expect(
      screen.getByText("USD 9,007,199,254,740,993.29"),
    ).toBeInTheDocument();
    expect(screen.getByText("1 missing a value")).toBeInTheDocument();
    expect(screen.getByText("67% value coverage")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Filters" }));
    fireEvent.change(screen.getByLabelText("From"), {
      target: { value: "2026-10-01" },
    });
    expect(
      screen.queryByText("USD 9,007,199,254,740,993.29"),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText("Build the report to apply these changes."),
    ).toBeInTheDocument();
  });
  it("shows loading and retries filter failures without presenting an empty successful list", () => {
    const retry = jest.fn();
    mockStatuses.mockReturnValue({
      data: [],
      isPending: false,
      isError: true,
      refetch: retry,
    });
    mockPeople.mockReturnValue({ data: [], isPending: true, isError: false });
    render(<CustomFieldReportPanel />);
    fireEvent.click(screen.getByRole("button", { name: "Filters" }));
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Statuses could not be loaded.",
    );
    expect(screen.getByRole("status")).toHaveTextContent("Loading people…");
    expect(
      screen.queryByText("This team has no statuses."),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry statuses" }));
    expect(retry).toHaveBeenCalledTimes(1);
  });
  it("shows absent averages as No value while retaining a valued zero", () => {
    const scrollIntoView = Object.getOwnPropertyDescriptor(
      HTMLElement.prototype,
      "scrollIntoView",
    );
    Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
      configurable: true,
      value: jest.fn(),
    });
    try {
      mockReport.mockReturnValue({
        mutate: mockMutate,
        reset: mockReset,
        isPending: false,
        error: null,
        data: {
          ...report,
          aggregation: "average",
          rows: [
            { key: "won", label: "Won", value: "0.0000000000000000", count: 2 },
            { key: "started", label: "Started", value: "", count: 0 },
          ],
        },
        variables: { ...initialInput, aggregation: "average" },
      });
      render(<CustomFieldReportPanel />);
      fireEvent.keyDown(
        screen.getByRole("combobox", { name: "Report measure" }),
        {
          key: "ArrowDown",
        },
      );
      fireEvent.click(screen.getByRole("option", { name: "Average" }));
      expect(
        within(screen.getByRole("row", { name: "Won USD 0.00 2" })).getByText(
          "USD 0.00",
        ),
      ).toBeInTheDocument();
      expect(
        within(
          screen.getByRole("row", { name: "Started No value 0" }),
        ).getByText("No value"),
      ).toBeInTheDocument();
    } finally {
      if (scrollIntoView) {
        Object.defineProperty(
          HTMLElement.prototype,
          "scrollIntoView",
          scrollIntoView,
        );
      } else {
        Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
      }
    }
  });
});
