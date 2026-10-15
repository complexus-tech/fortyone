import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";
import { DesktopUpdateSettings } from "./desktop-update-settings";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./desktop-provider", () => ({ useDesktop: jest.fn() }));
const getUpdateStatus = jest.fn();
const checkForUpdate = jest.fn();
const installUpdate = jest.fn();
beforeEach(() => {
  jest.clearAllMocks();
  jest
    .mocked(useDesktop)
    .mockReturnValue({ isDesktop: true } as ReturnType<typeof useDesktop>);
  jest.mocked(getDesktopBridge).mockReturnValue({
    getUpdateStatus,
    checkForUpdate,
    installUpdate,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
  getUpdateStatus.mockResolvedValue({
    configured: true,
    currentVersion: "1.0.0",
  });
  checkForUpdate.mockResolvedValue({ available: true, version: "1.1.0" });
});

it("stays absent from ordinary browsers and makes no updater calls", () => {
  jest
    .mocked(useDesktop)
    .mockReturnValue({ isDesktop: false } as ReturnType<typeof useDesktop>);
  expect(render(<DesktopUpdateSettings />).container).toBeEmptyDOMElement();
  expect(getUpdateStatus).not.toHaveBeenCalled();
});

it("explains unconfigured development builds without presenting a fake update check", async () => {
  getUpdateStatus.mockResolvedValue({
    configured: false,
    currentVersion: "0.1.0",
  });
  render(<DesktopUpdateSettings />);
  expect(await screen.findByText(/signed release build/)).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Check for updates" }),
  ).not.toBeInTheDocument();
  expect(checkForUpdate).not.toHaveBeenCalled();
});

it("requires review and explicit install, then blocks dismissal until failure or restart", async () => {
  let fail!: (reason: Error) => void;
  installUpdate.mockImplementation(
    () =>
      new Promise<void>((_resolve, reject) => {
        fail = reject;
      }),
  );
  render(<DesktopUpdateSettings />);
  fireEvent.click(
    await screen.findByRole("button", { name: "Check for updates" }),
  );
  fireEvent.click(
    await screen.findByRole("button", { name: "Review version 1.1.0" }),
  );
  expect(installUpdate).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Install and restart" }));
  expect(installUpdate).toHaveBeenCalledWith({ version: "1.1.0" });
  expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  act(() => {
    window.dispatchEvent(
      new CustomEvent("fortyone:desktop-update-progress", {
        detail: { phase: "downloading", received: 25, total: 100 },
      }),
    );
  });
  expect(screen.getByRole("status")).toHaveTextContent("25%");
  act(() => {
    window.dispatchEvent(
      new CustomEvent("fortyone:desktop-update-progress", {
        detail: { phase: "saving", received: 100, total: 100 },
      }),
    );
  });
  expect(screen.getByRole("status")).toHaveTextContent(
    "Saving your local work",
  );
  fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
  expect(screen.getByRole("dialog")).toBeInTheDocument();
  await act(async () => {
    fail(new Error("Local drafts could not be saved."));
  });
  await waitFor(() => {
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Local drafts could not be saved",
    );
  });
  expect(screen.getByRole("button", { name: "Cancel" })).toBeEnabled();
});
