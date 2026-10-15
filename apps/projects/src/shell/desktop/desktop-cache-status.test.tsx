import { act, fireEvent, render, screen } from "@testing-library/react";
import { DesktopCacheStatus } from "./desktop-cache-status";
import {
  DESKTOP_CACHE_RETRY_EVENT,
  updateDesktopCacheStatus,
} from "./cache-status";

jest.mock("./desktop-provider", () => ({
  useDesktop: () => ({ isDesktop: true }),
}));
jest.mock("./connectivity", () => ({ useDesktopOnline: () => true }));

beforeEach(() => {
  updateDesktopCacheStatus({
    workspaceId: "workspace",
    phase: "ready",
    storageError: undefined,
    downloadError: undefined,
    shellError: undefined,
  });
});

it.each(["restoring", "settings", "my-work", "ready"] as const)(
  "does not render ordinary cache activity in settings: %s",
  (phase) => {
    updateDesktopCacheStatus({ workspaceId: "workspace", phase });
    const view = render(<DesktopCacheStatus />);
    expect(view.container).toBeEmptyDOMElement();
  },
);

it("shows a previous storage failure when settings opens and supports retry", () => {
  updateDesktopCacheStatus({
    workspaceId: "workspace",
    shellError: "Not enough space to save app pages",
  });
  const retry = jest.fn();
  window.addEventListener(DESKTOP_CACHE_RETRY_EVENT, retry);
  render(<DesktopCacheStatus />);
  expect(screen.getByRole("alert")).toHaveTextContent("Not enough space");
  fireEvent.click(
    screen.getByRole("button", { name: "Retry offline storage" }),
  );
  expect(retry).toHaveBeenCalledTimes(1);
  act(() => {
    updateDesktopCacheStatus({
      workspaceId: "workspace",
      shellError: undefined,
    });
  });
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  window.removeEventListener(DESKTOP_CACHE_RETRY_EVENT, retry);
});
