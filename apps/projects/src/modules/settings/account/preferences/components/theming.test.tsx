import { act, within } from "@testing-library/react";
import type { Root } from "react-dom/client";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import { Theming } from "./theming";

let mockTheme: string | undefined;
const mockSetTheme = jest.fn();

jest.mock("next-themes", () => ({
  useTheme: () => ({ theme: mockTheme, setTheme: mockSetTheme }),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "task" }),
}));
jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: () => ({
    data: { openStoryInDialog: true },
  }),
}));
jest.mock("@/lib/hooks/users/update-auto-preferences", () => ({
  useUpdateAutomationPreferencesMutation: () => ({ mutate: jest.fn() }),
}));

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

it.each([
  ["system", "Sync with system"],
  ["light", "Day Mode"],
  ["dark", "Night Mode"],
])("hydrates the theme selector before restoring %s", async (theme, label) => {
  mockSetTheme.mockClear();
  mockTheme = undefined;
  const container = document.createElement("div");
  container.innerHTML = renderToString(<Theming />);
  document.body.appendChild(container);
  const appearance = () =>
    within(container).getByRole("combobox", { name: "Appearance" });
  const originalTrigger = appearance();

  // The server has no browser theme, but next-themes has already read it when
  // the first browser render begins. Hydration must still reuse the server DOM.
  mockTheme = theme;
  const onRecoverableError = jest.fn();
  const consoleError = jest
    .spyOn(console, "error")
    .mockImplementation(() => {});
  let root: Root | undefined;
  try {
    await act(async () => {
      root = hydrateRoot(container, <Theming />, { onRecoverableError });
    });

    expect(onRecoverableError).not.toHaveBeenCalled();
    expect(consoleError).not.toHaveBeenCalled();
    expect(appearance()).toBe(originalTrigger);
    expect(appearance()).toHaveTextContent(label);
    expect(appearance()).not.toHaveAttribute("data-placeholder");
    expect(mockSetTheme).not.toHaveBeenCalled();
  } finally {
    await act(async () => {
      root?.unmount();
    });
    container.remove();
    consoleError.mockRestore();
  }
});
