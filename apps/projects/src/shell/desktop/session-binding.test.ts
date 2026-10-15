import { getDesktopBridge } from "./bridge";
import { getDesktopOnline } from "./connectivity";
import { bindDesktopSession, clearDesktopBindings } from "./session-binding";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./connectivity", () => ({ getDesktopOnline: jest.fn() }));

const scope = {
  accountId: "alice",
  workspaceId: "one",
  workspaceSlug: "first",
};
const bindSession = jest.fn();
const restoreSession = jest.fn();
beforeEach(() => {
  jest.clearAllMocks();
  clearDesktopBindings();
  jest.mocked(getDesktopOnline).mockReturnValue(false);
  bindSession.mockResolvedValue(undefined);
  restoreSession.mockResolvedValue({
    ...scope,
    expiresAt: Date.now() + 60_000,
  });
  jest.mocked(getDesktopBridge).mockReturnValue({
    bindSession,
    restoreSession,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
});

it("uses the native current-cookie lease without starting an HTTP verification while offline", async () => {
  bindSession.mockImplementation(
    () =>
      new Promise(() => {
        /* Simulate a paused server that never responds. */
      }),
  );
  await expect(bindDesktopSession(scope)).resolves.toBe(true);
  expect(restoreSession).toHaveBeenCalledWith({ workspaceSlug: "first" });
  expect(bindSession).not.toHaveBeenCalled();
});

it.each(["accountId", "workspaceId", "workspaceSlug"])(
  "rejects a native restored identity with a different %s",
  async (field) => {
    restoreSession.mockResolvedValue({ ...scope, [field]: "different" });
    await expect(bindDesktopSession(scope)).rejects.toThrow(
      "account or workspace changed",
    );
    expect(bindSession).not.toHaveBeenCalled();
  },
);

it("preserves native denial and expiry failures without falling back to HTTP", async () => {
  restoreSession.mockRejectedValueOnce(
    new Error("Reconnect to renew offline access"),
  );
  await expect(bindDesktopSession(scope)).rejects.toThrow("Reconnect to renew");
  expect(bindSession).not.toHaveBeenCalled();
  await expect(bindDesktopSession(scope)).resolves.toBe(true);
  expect(restoreSession).toHaveBeenCalledTimes(2);
});

it("does not fall back to an offline lease after an authoritative online denial", async () => {
  jest.mocked(getDesktopOnline).mockReturnValue(true);
  bindSession.mockRejectedValueOnce(new Error("Sign in and verify access"));
  await expect(bindDesktopSession(scope)).rejects.toThrow(
    "Sign in and verify access",
  );
  expect(restoreSession).not.toHaveBeenCalled();
});

it("reverifies online after an offline lease and does not share a hung online check with offline reads", async () => {
  await bindDesktopSession(scope);
  jest.mocked(getDesktopOnline).mockReturnValue(true);
  let finish!: () => void;
  bindSession.mockImplementationOnce(
    () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
  );
  const online = bindDesktopSession(scope);
  expect(bindSession).toHaveBeenCalledTimes(1);
  jest.mocked(getDesktopOnline).mockReturnValue(false);
  await expect(bindDesktopSession(scope)).resolves.toBe(true);
  expect(restoreSession).toHaveBeenCalledTimes(2);
  finish();
  await online;
});
