import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { CustomField } from "./types";
import { CustomFieldEditor } from "./field-editor";
import { CustomFieldIconPicker } from "./icon-picker";
import {
  customFieldIconGroups,
  customFieldIconOptions,
  getCustomFieldIconKey,
} from "./icons";
import {
  customFieldIconKeys,
  customFieldReportSchema,
  customFieldSchema,
} from "./types";

const mockCreate = jest.fn();
const mockUpdate = jest.fn();
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
  if (resizeObserverDescriptor)
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
  if (scrollIntoViewDescriptor)
    Object.defineProperty(
      HTMLElement.prototype,
      "scrollIntoView",
      scrollIntoViewDescriptor,
    );
  else Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
});

jest.mock("./hooks", () => ({
  useCustomFieldMutations: () => ({
    create: { mutateAsync: mockCreate, isPending: false },
    update: { mutateAsync: mockUpdate, isPending: false },
  }),
}));

const field: CustomField = {
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

describe("field icon configuration", () => {
  beforeEach(() => {
    mockCreate.mockReset().mockResolvedValue(field);
    mockUpdate.mockReset().mockResolvedValue(field);
  });

  it("accepts legacy definitions and validates the persisted icon catalogue", () => {
    const { icon, ...legacy } = field;
    expect(customFieldSchema.parse(legacy).icon).toBeUndefined();
    expect(customFieldSchema.parse({ ...field, icon: null }).icon).toBeNull();
    expect(customFieldSchema.parse({ ...field, icon: "workspace" }).icon).toBe(
      "workspace",
    );
    expect(
      customFieldSchema.safeParse({ ...field, icon: "untrusted-icon" }).success,
    ).toBe(false);
    expect(icon).toBe("star");
  });

  it("keeps every persisted key selectable and supports expanded icons in exact-value reports", () => {
    const selectableKeys = customFieldIconOptions.map((option) => option.value);
    expect(new Set(selectableKeys).size).toBe(selectableKeys.length);
    expect([...selectableKeys].sort()).toEqual([...customFieldIconKeys].sort());
    for (const icon of customFieldIconKeys) {
      expect(customFieldSchema.parse({ ...field, icon }).icon).toBe(icon);
    }
    expect(
      customFieldReportSchema.parse({
        field: { ...field, icon: "analytics" },
        aggregation: "sum",
        groupBy: "none",
        currency: "USD",
        totalCount: 1,
        valuedCount: 1,
        missingCount: 0,
        rows: [{ key: "all", label: "All", value: "1250.50", count: 1 }],
      }),
    ).toEqual(
      expect.objectContaining({
        field: expect.objectContaining({ icon: "analytics" }),
        rows: [{ key: "all", label: "All", value: "1250.50", count: 1 }],
      }),
    );
    expect(getCustomFieldIconKey({ type: "money", icon: null })).toBe("money");
  });

  it("searches existing icons and persists the chosen key on an edit", async () => {
    const onClose = jest.fn();
    render(
      <CustomFieldEditor
        field={field}
        onClose={onClose}
        teamId={field.teamId}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Field icon: Star" }));
    const search = await screen.findByRole("combobox", {
      name: "Search field icons",
    });
    fireEvent.change(search, { target: { value: "calendar" } });
    await screen.findByRole("option", { name: "Calendar" });
    fireEvent.keyDown(search, { key: "Enter" });
    expect(
      screen.getByRole("button", { name: "Field icon: Calendar" }),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Save field" }));
    await waitFor(() => {
      expect(mockUpdate).toHaveBeenCalledWith({
        fieldId: field.id,
        input: {
          name: field.name,
          icon: "calendar",
          options: [],
          showOnCreate: true,
        },
      });
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  it("resets a previously chosen icon to Automatic with an explicit null", async () => {
    render(
      <CustomFieldEditor
        field={field}
        onClose={jest.fn()}
        teamId={field.teamId}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Field icon: Star" }));
    fireEvent.click(await screen.findByRole("option", { name: "Automatic" }));
    fireEvent.click(screen.getByRole("button", { name: "Save field" }));
    await waitFor(() => {
      expect(mockUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          input: expect.objectContaining({ icon: null }),
        }),
      );
    });
  });

  it("exposes named icon choices without category headers and searches business synonyms", async () => {
    render(
      <CustomFieldEditor
        field={field}
        onClose={jest.fn()}
        teamId={field.teamId}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Field icon: Star" }));
    const search = await screen.findByRole("combobox", {
      name: "Search field icons",
    });
    for (const group of customFieldIconGroups) {
      expect(screen.queryByText(group)).not.toBeInTheDocument();
    }
    for (const option of customFieldIconOptions) {
      expect(
        screen.getByRole("option", { name: option.label }),
      ).toHaveAccessibleName(option.label);
    }
    fireEvent.change(search, { target: { value: "referral" } });
    expect(await screen.findByRole("option", { name: "Share" })).toBeVisible();
    expect(screen.queryByRole("option", { name: "Calendar" })).toBeNull();
    fireEvent.keyDown(search, { key: "Enter" });
    expect(
      screen.getByRole("button", { name: "Field icon: Share" }),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Save field" }));
    await waitFor(() => {
      expect(mockUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          input: expect.objectContaining({ icon: "share" }),
        }),
      );
    });
  });

  it("navigates rows and columns without changing the saved icon until Enter", async () => {
    const onChange = jest.fn();
    render(
      <CustomFieldIconPicker onChange={onChange} type="text" value={null} />,
    );
    const trigger = screen.getByRole("button", {
      name: "Field icon: Automatic",
    });
    fireEvent.click(trigger);
    const search = await screen.findByRole("combobox", {
      name: "Search field icons",
    });
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "Automatic" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    for (const [key, label] of [
      ["ArrowRight", "Text"],
      ["ArrowDown", "Star"],
      ["ArrowUp", "Text"],
      ["ArrowLeft", "Automatic"],
      ["ArrowRight", "Text"],
      ["ArrowDown", "Star"],
      ["Home", "Automatic"],
      ["End", customFieldIconOptions[customFieldIconOptions.length - 1].label],
      ["Home", "Automatic"],
      ["ArrowRight", "Text"],
      ["ArrowDown", "Star"],
    ]) {
      fireEvent.keyDown(search, { key });
      const option = screen.getByRole("option", { name: label });
      expect(option).toHaveAttribute("aria-selected", "true");
      expect(search).toHaveAttribute("aria-activedescendant", option.id);
    }
    expect(onChange).not.toHaveBeenCalled();
    expect(trigger).toHaveAccessibleName("Field icon: Automatic");

    fireEvent.keyDown(search, { key: "Enter" });
    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith("star");
      expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
      expect(trigger).toHaveFocus();
    });
  });

  it("leaves horizontal arrows available for editing a search query", async () => {
    const onChange = jest.fn();
    render(
      <CustomFieldIconPicker onChange={onChange} type="text" value={null} />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Field icon: Automatic" }),
    );
    const search = await screen.findByRole("combobox", {
      name: "Search field icons",
    });
    fireEvent.change(search, { target: { value: "money" } });
    const money = await screen.findByRole("option", { name: "Money" });
    await waitFor(() => {
      expect(money).toHaveAttribute("aria-selected", "true");
    });

    fireEvent.keyDown(search, { key: "ArrowLeft" });
    fireEvent.keyDown(search, { key: "ArrowRight" });

    expect(money).toHaveAttribute("aria-selected", "true");
    expect(search).toHaveValue("money");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("closes only the picker on Escape while preserving the field draft and restoring focus", async () => {
    const onClose = jest.fn();
    render(
      <CustomFieldEditor
        field={field}
        onClose={onClose}
        teamId={field.teamId}
      />,
    );
    const name = screen.getByRole("textbox", { name: /Field name/ });
    fireEvent.change(name, { target: { value: "Marketing pipeline" } });
    const trigger = screen.getByRole("button", { name: "Field icon: Star" });
    fireEvent.click(trigger);
    const search = await screen.findByRole("combobox", {
      name: "Search field icons",
    });
    fireEvent.change(search, { target: { value: "calendar" } });
    fireEvent.keyDown(search, { key: "Escape" });

    await waitFor(() => {
      expect(
        screen.queryByRole("combobox", { name: "Search field icons" }),
      ).not.toBeInTheDocument();
      expect(
        screen.getByRole("dialog", { name: "Edit field" }),
      ).toBeInTheDocument();
      expect(trigger).toHaveFocus();
    });
    expect(name).toHaveValue("Marketing pipeline");
    expect(onClose).not.toHaveBeenCalled();
    expect(mockUpdate).not.toHaveBeenCalled();

    fireEvent.click(trigger);
    expect(
      await screen.findByRole("combobox", { name: "Search field icons" }),
    ).toHaveValue("");
    expect(trigger).toHaveAccessibleName("Field icon: Star");
    expect(name).toHaveValue("Marketing pipeline");
  });

  it("does not submit the field while pressing Enter in an empty icon search", async () => {
    const onClose = jest.fn();
    render(
      <CustomFieldEditor
        field={field}
        onClose={onClose}
        teamId={field.teamId}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Field icon: Star" }));
    const search = await screen.findByRole("combobox", {
      name: "Search field icons",
    });
    fireEvent.change(search, { target: { value: "not-an-icon-zzzz" } });
    expect(await screen.findByText("No icons found.")).toBeVisible();
    const allowsFormSubmission = fireEvent.keyDown(search, { key: "Enter" });

    expect(allowsFormSubmission).toBe(false);
    expect(mockUpdate).not.toHaveBeenCalled();
    expect(mockCreate).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
    expect(
      screen.getByRole("dialog", { name: "Edit field" }),
    ).toBeInTheDocument();
    expect(search).toHaveValue("not-an-icon-zzzz");
  });

  it("prevents changing an icon while the control is disabled", () => {
    const onChange = jest.fn();
    render(
      <CustomFieldIconPicker
        disabled
        onChange={onChange}
        type="text"
        value="automation"
      />,
    );
    const trigger = screen.getByRole("button", {
      name: "Field icon: Automation",
    });
    expect(trigger).toBeDisabled();
    fireEvent.click(trigger);
    expect(screen.queryByRole("combobox")).toBeNull();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("persists a chosen icon when creating a field without changing its type", async () => {
    render(
      <CustomFieldEditor
        field={null}
        onClose={jest.fn()}
        teamId={field.teamId}
      />,
    );
    fireEvent.change(screen.getByRole("textbox", { name: /Field name/ }), {
      target: { value: "Company" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Field icon: Automatic" }),
    );
    fireEvent.click(await screen.findByRole("option", { name: "Workspace" }));
    fireEvent.click(screen.getByRole("button", { name: "Create field" }));
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith({
        name: "Company",
        type: "text",
        icon: "workspace",
        currency: null,
        options: [],
        showOnCreate: false,
      });
    });
  });
});
