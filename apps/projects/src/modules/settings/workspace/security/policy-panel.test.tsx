import { fireEvent, render, screen } from "@testing-library/react";
import type { SecurityPolicy } from "./types";
import { SecurityPolicies } from "./policy-panel";

const DEFAULT_POLICY: SecurityPolicy = {
  allowedDomains: [],
  allowGuests: true,
  maxSessionAgeHours: 0,
  version: 3,
  updatedAt: null,
};
let mockPolicy = DEFAULT_POLICY;
const mockSave = jest.fn();
const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
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
});

jest.mock("./hooks", () => ({
  useSecurityPolicy: () => ({
    data: mockPolicy,
    isPending: false,
    isError: false,
  }),
  useSaveSecurityPolicy: () => ({
    mutate: mockSave,
    isPending: false,
    isError: false,
  }),
}));

beforeEach(() => {
  jest.clearAllMocks();
  mockPolicy = DEFAULT_POLICY;
});

it("keeps an unlimited policy unchanged until a login time limit is enabled", () => {
  render(<SecurityPolicies />);
  const toggle = screen.getByRole("checkbox", {
    name: "Require a new login after a set time",
  });
  const save = screen.getByRole("button", { name: "Save policy" });
  expect(toggle).not.toBeChecked();
  expect(
    screen.queryByRole("spinbutton", { name: "Hours" }),
  ).not.toBeInTheDocument();
  expect(save).toBeDisabled();

  fireEvent.click(toggle);
  const hours = screen.getByRole("spinbutton", { name: "Hours" });
  expect(hours).toHaveValue(24);
  fireEvent.change(hours, { target: { value: "12" } });
  expect(save).toBeEnabled();
  fireEvent.click(toggle);
  expect(
    screen.queryByRole("spinbutton", { name: "Hours" }),
  ).not.toBeInTheDocument();
  expect(save).toBeDisabled();
  expect(mockSave).not.toHaveBeenCalled();

  fireEvent.click(toggle);
  expect(screen.getByRole("spinbutton", { name: "Hours" })).toHaveValue(12);
  fireEvent.click(save);
  expect(mockSave.mock.calls[0][0]).toEqual({
    allowedDomains: [],
    allowGuests: true,
    maxSessionAgeHours: 12,
    expectedVersion: 3,
  });
});

it("loads an existing time limit and sends zero when it is disabled", () => {
  mockPolicy = { ...DEFAULT_POLICY, maxSessionAgeHours: 48 };
  render(<SecurityPolicies />);
  const toggle = screen.getByRole("checkbox", {
    name: "Require a new login after a set time",
  });
  const save = screen.getByRole("button", { name: "Save policy" });
  expect(toggle).toBeChecked();
  expect(screen.getByRole("spinbutton", { name: "Hours" })).toHaveValue(48);
  expect(save).toBeDisabled();

  fireEvent.change(screen.getByRole("spinbutton", { name: "Hours" }), {
    target: { value: "" },
  });
  expect(save).toBeDisabled();
  fireEvent.click(toggle);
  expect(
    screen.queryByRole("spinbutton", { name: "Hours" }),
  ).not.toBeInTheDocument();
  expect(save).toBeEnabled();
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  fireEvent.click(save);
  expect(mockSave.mock.calls[0][0]).toMatchObject({ maxSessionAgeHours: 0 });
});

it("validates enabled hours as a positive whole number within the existing limit", () => {
  mockPolicy = { ...DEFAULT_POLICY, maxSessionAgeHours: 48 };
  render(<SecurityPolicies />);
  const hours = screen.getByRole("spinbutton", { name: "Hours" });
  const save = screen.getByRole("button", { name: "Save policy" });

  for (const value of ["0", "1.5", "721"]) {
    fireEvent.change(hours, { target: { value } });
    expect(save).toBeDisabled();
    expect(screen.getByRole("alert")).toHaveTextContent("between 1 and 720");
  }
  fireEvent.change(hours, { target: { value: "720" } });
  expect(save).toBeEnabled();
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  expect(mockSave).not.toHaveBeenCalled();
});
