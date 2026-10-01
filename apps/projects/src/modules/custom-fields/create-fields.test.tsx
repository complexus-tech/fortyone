import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { CustomField } from "./types";
import { CreateCustomFields } from "./create-fields";

const mockFields = jest.fn();
jest.mock("./hooks", () => ({
  useTeamCustomFields: (...args: unknown[]) => mockFields(...args),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [], isPending: false, isError: false }),
}));

const money: CustomField = {
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

describe("creation property controls", () => {
  beforeEach(() => {
    mockFields.mockReturnValue({
      data: [money],
      isPending: false,
      isError: false,
    });
  });
  it("uses a native property chip with currency and applies validated exact amounts", async () => {
    const onChange = jest.fn();
    render(
      <CreateCustomFields
        onChange={onChange}
        teamId={money.teamId}
        values={{}}
      />,
    );
    const chip = screen.getByRole("button", {
      name: "Contract value: Not set",
    });
    expect(chip).toHaveTextContent("Not set");
    expect(chip).not.toHaveTextContent("Contract value");
    expect(chip).toHaveAttribute("title", "Contract value: Not set");
    expect(chip.querySelector("svg")).not.toBeNull();
    expect(chip).toHaveClass("dark:bg-surface-elevated/90");
    expect(screen.queryByText("Custom fields")).not.toBeInTheDocument();
    fireEvent.click(chip);
    const input = await screen.findByRole("textbox", {
      name: "Contract value (USD)",
    });
    fireEvent.change(input, { target: { value: "1,250.50" } });
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));
    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
    fireEvent.change(input, { target: { value: "9007199254740993.29" } });
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));
    expect(onChange).toHaveBeenCalledWith({
      [money.id]: "9007199254740993.29",
    });
    await waitFor(() => {
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    });
  });
  it("shows a real zero as the value without a visible field-name prefix", () => {
    render(
      <CreateCustomFields
        onChange={jest.fn()}
        teamId={money.teamId}
        values={{ [money.id]: "0" }}
      />,
    );

    const chip = screen.getByRole("button", {
      name: "Contract value: USD 0.00",
    });
    expect(chip).toHaveTextContent("USD 0.00");
    expect(chip).not.toHaveTextContent("Contract value");
    expect(chip).not.toHaveTextContent("Not set");
    expect(chip).toHaveAttribute("title", "Contract value: USD 0.00");
  });
  it("keeps optional fields behind the property picker and excludes archived fields", async () => {
    const hidden = {
      ...money,
      id: "d0f7eece-94cc-4cc7-83ba-33d632a09c89",
      name: "Follow-up",
      type: "text" as const,
      currency: null,
      showOnCreate: false,
    };
    mockFields.mockReturnValue({
      data: [
        money,
        hidden,
        {
          ...hidden,
          id: "archived",
          name: "Archived field",
          archivedAt: money.createdAt,
        },
      ],
      isPending: false,
      isError: false,
    });
    render(
      <CreateCustomFields
        onChange={jest.fn()}
        teamId={money.teamId}
        values={{}}
      />,
    );
    expect(
      screen.queryByRole("button", { name: "Follow-up: Not set" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Properties" }));
    const option = await screen.findByRole("option", { name: "Follow-up" });
    expect(screen.queryByText("Archived field")).not.toBeInTheDocument();
    fireEvent.click(option);
    expect(
      await screen.findByRole("button", { name: "Follow-up: Not set" }),
    ).toBeInTheDocument();
  });
});
