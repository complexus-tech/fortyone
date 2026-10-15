import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";
import { DesktopSettings } from "./desktop-settings";
import { useDesktopOutbox } from "./desktop-outbox-coordinator";
import { updateDesktopCacheStatus } from "./cache-status";

jest.mock("./desktop-outbox-coordinator", () => ({
  useDesktopOutbox: jest.fn(),
  resolveDesktopOperation: jest.fn(),
  refreshDesktopOutbox: jest.fn(),
}));
jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./desktop-provider", () => ({ useDesktop: jest.fn() }));

const initial = { notifications: false, keepRunning: true, autostart: false };
const getPreferences = jest.fn();
const updatePreferences = jest.fn();
const requestNotificationPermission = jest.fn();
const show = (section: "notifications" | "behavior" = "notifications") =>
  render(
    <QueryClientProvider
      client={
        new QueryClient({
          defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
          },
        })
      }
    >
      <DesktopSettings section={section} />
    </QueryClientProvider>,
  );

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useDesktopOutbox).mockReturnValue(null);
  updateDesktopCacheStatus({
    workspaceId: "workspace",
    phase: "ready",
    storageError: undefined,
    shellError: undefined,
    downloadError: undefined,
  });
  jest
    .mocked(useDesktop)
    .mockReturnValue({ isDesktop: true } as ReturnType<typeof useDesktop>);
  jest.mocked(getDesktopBridge).mockReturnValue({
    getPreferences,
    updatePreferences,
    requestNotificationPermission,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
  getPreferences.mockResolvedValue(initial);
  updatePreferences.mockImplementation(async (patch) => ({
    ...initial,
    ...patch,
  }));
  requestNotificationPermission.mockResolvedValue("granted");
});

it("keeps desktop controls out of ordinary browser settings", () => {
  jest
    .mocked(useDesktop)
    .mockReturnValue({ isDesktop: false } as ReturnType<typeof useDesktop>);
  const view = show();
  expect(view.container).toBeEmptyDOMElement();
  expect(getPreferences).not.toHaveBeenCalled();
});

it("asks for notification permission only after the user enables the switch", async () => {
  show();
  const toggle = await screen.findByRole("switch", {
    name: "Desktop notifications",
  });
  expect(requestNotificationPermission).not.toHaveBeenCalled();
  fireEvent.click(toggle);
  await waitFor(() => {
    expect(toggle).toBeChecked();
  });
  expect(requestNotificationPermission).toHaveBeenCalledTimes(1);
  expect(updatePreferences).toHaveBeenCalledWith({ notifications: true });
});

it("keeps the preference disabled and explains denied system permission", async () => {
  requestNotificationPermission.mockResolvedValue("denied");
  show();
  const toggle = await screen.findByRole("switch", {
    name: "Desktop notifications",
  });
  fireEvent.click(toggle);
  expect(await screen.findByRole("alert")).toHaveTextContent("system settings");
  expect(toggle).not.toBeChecked();
  expect(updatePreferences).not.toHaveBeenCalled();
});

it("disables duplicate writes and displays the saved launch preference", async () => {
  let finish!: (value: typeof initial) => void;
  updatePreferences.mockReturnValue(
    new Promise((resolve) => {
      finish = resolve;
    }),
  );
  show("behavior");
  const toggle = await screen.findByRole("switch", { name: "Launch at login" });
  fireEvent.click(toggle);
  await waitFor(() => {
    expect(toggle).toBeDisabled();
  });
  fireEvent.click(toggle);
  expect(updatePreferences).toHaveBeenCalledTimes(1);
  await act(async () => {
    finish({ ...initial, autostart: true });
  });
  await waitFor(() => {
    expect(toggle).toBeChecked();
    expect(toggle).toBeEnabled();
  });
  expect(requestNotificationPermission).not.toHaveBeenCalled();
});

it("surfaces native save failures without changing the displayed value", async () => {
  updatePreferences.mockRejectedValue("Login items could not be updated");
  show("behavior");
  const toggle = await screen.findByRole("switch", { name: "Launch at login" });
  fireEvent.click(toggle);
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Login items could not be updated",
  );
  expect(toggle).not.toBeChecked();
});

it("keeps conflict review and storage recovery in Desktop App settings", async () => {
  updateDesktopCacheStatus({
    workspaceId: "workspace",
    storageError: "Disk is full",
  });
  jest.mocked(useDesktopOutbox).mockReturnValue({
    scope: {
      accountId: "account",
      workspaceId: "workspace",
      workspaceSlug: "acme",
    },
    snapshot: { operations: [], completed: [] },
    busy: false,
    online: true,
    error: "A change needs review",
  });
  show("behavior");
  expect(
    await screen.findByRole("switch", { name: "Launch at login" }),
  ).toBeInTheDocument();
  expect(screen.getByRole("alert")).toHaveTextContent("Disk is full");
  expect(
    screen.getByRole("button", { name: "Retry offline storage" }),
  ).toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Review saved changes" }));
  expect(screen.getByText("A change needs review")).toBeVisible();
});
