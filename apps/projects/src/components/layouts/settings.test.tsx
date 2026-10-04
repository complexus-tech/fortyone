import type { ReactNode } from "react";
import { useEffect } from "react";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { Select } from "ui";
import { SettingsLayout } from "./settings";

const mockPush = jest.fn();
const mockSetPrevPage = jest.fn();
let mockPrevPage = "/acme/my-work";

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => "/acme/settings/account/preferences",
}));
jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ href, children }: { href: string; children: ReactNode }) => (
    <a href={href}>{children}</a>
  ),
}));
jest.mock("@/hooks", () => ({
  useLocalStorage: () => [mockPrevPage, mockSetPrevPage],
  useUserRole: () => ({ userRole: "admin" }),
  useTerminology: () => ({ getTermDisplay: () => "Objectives" }),
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/modules/invitations/hooks/my-invitations", () => ({
  useMyInvitations: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({ hasFeature: () => false }),
}));
jest.mock("@/shell/commands/commands", () => ({ Commands: () => null }));
jest.mock("../shared/mobile-menu", () => ({ MobileMenuButton: () => null }));
jest.mock("../ui", () => ({
  NavLink: ({ href, children }: { href: string; children: ReactNode }) => (
    <a href={href}>{children}</a>
  ),
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

beforeEach(() => {
  jest.clearAllMocks();
  mockPrevPage = "/acme/my-work";
});

afterEach(() => {
  fireEvent.keyUp(document, { key: "Escape", code: "Escape" });
});

const LabeledSettingsContent = ({ onMount }: { onMount: () => void }) => {
  useEffect(onMount, [onMount]);
  return (
    <>
      <label htmlFor="settings-test-domain">Allowed email domains</label>
      <input id="settings-test-domain" />
    </>
  );
};

it("mounts settings content once and associates labels with its visible fields", () => {
  const onMount = jest.fn();
  const { container } = render(
    <SettingsLayout>
      <LabeledSettingsContent onMount={onMount} />
    </SettingsLayout>,
  );
  const canvas = container.querySelector<HTMLElement>(
    "[data-settings-content-canvas]",
  );
  if (!canvas) throw new Error("Settings content was not rendered");

  expect(onMount).toHaveBeenCalledTimes(1);
  expect(screen.getAllByRole("textbox")).toHaveLength(1);
  expect(container.querySelectorAll("#settings-test-domain")).toHaveLength(1);
  expect(screen.getByLabelText("Allowed email domains")).toBe(
    within(canvas).getByRole("textbox", { name: "Allowed email domains" }),
  );
});

it("lets a child handle bubbling Escape without leaving Settings", () => {
  const { container } = render(
    <SettingsLayout>
      <button
        onKeyDown={(event) => {
          if (event.key === "Escape") event.preventDefault();
        }}
        type="button"
      >
        Handle Escape
      </button>
    </SettingsLayout>,
  );
  const shell = container.querySelector<HTMLElement>("[data-settings-shell]");
  if (!shell) throw new Error("Settings shell was not rendered");
  const event = new KeyboardEvent("keydown", {
    key: "Escape",
    code: "Escape",
    bubbles: true,
    cancelable: true,
  });
  fireEvent(
    within(shell).getByRole("button", { name: "Handle Escape" }),
    event,
  );

  expect(event.defaultPrevented).toBe(true);
  expect(mockPush).not.toHaveBeenCalled();
  expect(mockSetPrevPage).not.toHaveBeenCalled();
});

it("closes a real settings select on Escape and navigates back on the next Escape", async () => {
  const { container } = render(
    <SettingsLayout>
      <Select defaultValue="dark">
        <Select.Trigger aria-label="Appearance">
          <Select.Input />
        </Select.Trigger>
        <Select.Content>
          <Select.Option value="light">Day Mode</Select.Option>
          <Select.Option value="dark">Night Mode</Select.Option>
        </Select.Content>
      </Select>
    </SettingsLayout>,
  );
  const shell = container.querySelector<HTMLElement>("[data-settings-shell]");
  if (!shell) throw new Error("Settings shell was not rendered");
  const trigger = within(shell).getByRole("combobox", { name: "Appearance" });
  fireEvent.keyDown(trigger, { key: "ArrowDown", code: "ArrowDown" });
  fireEvent.keyUp(trigger, { key: "ArrowDown", code: "ArrowDown" });
  const menu = await screen.findByRole("listbox");
  fireEvent.keyDown(menu, { key: "Escape", code: "Escape" });
  fireEvent.keyUp(menu, { key: "Escape", code: "Escape" });

  await waitFor(() => {
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });
  expect(mockPush).not.toHaveBeenCalled();
  expect(mockSetPrevPage).not.toHaveBeenCalled();

  fireEvent.keyDown(trigger, { key: "Escape", code: "Escape" });
  expect(mockPush).toHaveBeenCalledTimes(1);
  expect(mockPush).toHaveBeenCalledWith("/acme/my-work");
  expect(mockSetPrevPage).toHaveBeenCalledWith("");
});

it("uses the workspace home when a normal Escape has no previous page", () => {
  mockPrevPage = "";
  render(<SettingsLayout>Preferences</SettingsLayout>);

  fireEvent.keyDown(document, { key: "Escape", code: "Escape" });

  expect(mockPush).toHaveBeenCalledTimes(1);
  expect(mockPush).toHaveBeenCalledWith("/acme/maya");
  expect(mockSetPrevPage).toHaveBeenCalledWith("");
});
