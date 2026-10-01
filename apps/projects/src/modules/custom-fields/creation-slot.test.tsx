import { act, render, screen } from "@testing-library/react";
import { createRef } from "react";
import type { CreationPropertiesController } from "@/shared/story/creation-property-slots";
import type { CustomField } from "./types";
import { CustomFieldsCreationSlot } from "./creation-slot";

const mockFields = jest.fn();
jest.mock("./hooks", () => ({
  useTeamCustomFields: (...args: unknown[]) => mockFields(...args),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [] }),
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

describe("creation property capability", () => {
  beforeEach(() => {
    mockFields.mockReturnValue({
      data: [amount],
      isPending: false,
      isError: false,
    });
  });

  it("prepares the latest exact draft and resets without duplicating creation state", () => {
    const controller = createRef<CreationPropertiesController>();
    render(
      <CustomFieldsCreationSlot
        controllerRef={controller}
        teamId={amount.teamId}
      />,
    );
    act(() => {
      controller.current!.setValues({ [amount.id]: "9007199254740993.29" });
    });
    expect(controller.current!.prepareValues()).toEqual([
      { fieldId: amount.id, value: "9007199254740993.29" },
    ]);
    const chip = screen.getByRole("button", {
      name: "Contract value: USD 9,007,199,254,740,993.29",
    });
    expect(chip).toHaveTextContent("USD 9,007,199,254,740,993.29");
    expect(chip).not.toHaveTextContent("Contract value");
    expect(chip).toHaveAttribute(
      "title",
      "Contract value: USD 9,007,199,254,740,993.29",
    );
    act(() => {
      controller.current!.reset();
    });
    expect(controller.current!.prepareValues()).toEqual([]);
    expect(
      screen.getByRole("button", { name: "Contract value: Not set" }),
    ).toHaveTextContent("Not set");
  });

  it("observes changed team and query state before submission", () => {
    const controller = createRef<CreationPropertiesController>();
    const { rerender } = render(
      <CustomFieldsCreationSlot
        controllerRef={controller}
        teamId={amount.teamId}
      />,
    );
    act(() => {
      controller.current!.setValues({ [amount.id]: "1250.50" });
    });
    mockFields.mockReturnValue({
      data: undefined,
      isPending: true,
      isError: false,
    });
    rerender(
      <CustomFieldsCreationSlot
        controllerRef={controller}
        teamId={amount.teamId}
      />,
    );
    expect(controller.current!.isPending).toBe(true);
    expect(() => controller.current!.prepareValues()).toThrow(
      "Custom fields have not loaded",
    );
    mockFields.mockReturnValue({ data: [], isPending: false, isError: false });
    rerender(
      <CustomFieldsCreationSlot
        controllerRef={controller}
        teamId="different-team"
      />,
    );
    expect(controller.current!.prepareValues()).toEqual([]);
  });
});
