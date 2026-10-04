import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import type { BrowserSession } from "./types";
import { SecuritySessions } from "./sessions-panel";

const SESSION: BrowserSession = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  userId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  name: "Ada Lovelace",
  username: "ada",
  browserName: "Chrome",
  email: "ada@example.com",
  role: "admin",
  authenticatedAt: "2026-10-03T08:00:00Z",
  lastSeenAt: "2026-10-04T09:00:00Z",
  expiresAt: "2099-10-04T09:00:00Z",
  revokedAt: null,
  revokedBefore: null,
  current: true,
};
const mockRevoke = jest.fn();
let mockPending = false;
let mockError: Error | null = null;
let mockSessionQueryError = false;
const sessions = [
  SESSION,
  {
    ...SESSION,
    id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
    current: false,
    browserName: null,
    username: undefined,
  },
  {
    ...SESSION,
    id: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
    name: "Grace Hopper",
    email: "grace@example.com",
    userId: "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee",
    revokedAt: "2026-10-04T10:00:00Z",
    current: false,
  },
  {
    ...SESSION,
    id: "ffffffff-ffff-4fff-8fff-ffffffffffff",
    expiresAt: "2020-10-04T09:00:00Z",
    current: false,
  },
  {
    ...SESSION,
    id: "77777777-7777-4777-8777-777777777777",
    revokedBefore: "2026-10-04T10:00:00Z",
    current: false,
  },
];
let mockSessionItems = sessions;

jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({
    data: [
      {
        id: SESSION.userId,
        username: "ada",
        fullName: SESSION.name,
        email: SESSION.email,
        isSystem: false,
      },
    ],
  }),
}));
jest.mock("./hooks", () => ({
  useSecuritySessions: () => ({
    data: { items: mockSessionItems, hasMore: false },
    isPending: false,
    isFetching: false,
    isError: mockSessionQueryError,
    refetch: jest.fn(),
  }),
  useRevokeSessions: () => ({
    mutate: mockRevoke,
    reset: jest.fn(),
    isPending: mockPending,
    isError: Boolean(mockError),
    error: mockError,
  }),
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
  if (resizeObserverDescriptor)
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
});
beforeEach(() => {
  jest.clearAllMocks();
  mockPending = false;
  mockError = null;
  mockSessionQueryError = false;
  mockSessionItems = sessions;
});

const openActions = async (shortId: string, name = "ada") => {
  const trigger = screen.getByRole("button", {
    name: `Session actions for ${name} (${shortId})`,
  });
  trigger.focus();
  fireEvent.keyDown(trigger, { key: "Enter", code: "Enter" });
  await screen.findByRole("menu");
  return trigger;
};

it("shows only active sessions with usernames, browsers and compact UTC timestamps", () => {
  render(<SecuritySessions />);
  const table = screen.getByRole("table", { name: /Member browser sessions/ });
  expect(
    within(table)
      .getAllByRole("columnheader")
      .map((cell) => cell.textContent),
  ).toEqual([
    "Member",
    "Browser",
    "Last active",
    "Signed in",
    "Expires",
    "Actions",
  ]);
  const rows = within(table).getAllByRole("row").slice(1);
  expect(rows).toHaveLength(2);
  expect(within(rows[0]).getByText("ada")).toBeInTheDocument();
  expect(within(rows[0]).getByText("Chrome")).toBeInTheDocument();
  expect(within(rows[1]).getByText("ada")).toBeInTheDocument();
  expect(within(rows[1]).getByText("Unknown browser")).toBeInTheDocument();
  expect(rows[0]).toHaveTextContent("This browser");
  expect(rows[0]).not.toHaveTextContent(SESSION.email);
  expect(rows[0]).not.toHaveTextContent(SESSION.name);
  expect(rows[0]).not.toHaveTextContent(SESSION.id);
  expect(
    Array.from(rows[0].querySelectorAll("time"), (time) => time.dateTime),
  ).toEqual([SESSION.lastSeenAt, SESSION.authenticatedAt, SESSION.expiresAt]);
  expect(
    within(rows[0]).getByLabelText("04 Oct 2026 at 09:00 UTC"),
  ).toHaveTextContent("04 Oct 2026 · 09:00");
  expect(screen.queryByText("Grace Hopper")).not.toBeInTheDocument();
  expect(
    screen.queryByRole("columnheader", { name: "Status" }),
  ).not.toBeInTheDocument();
  expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: /Revoke session/ }),
  ).not.toBeInTheDocument();
});

it("hides retained session rows when refreshing current access fails", () => {
  const { rerender } = render(<SecuritySessions />);
  expect(screen.getByRole("table")).toBeInTheDocument();
  mockSessionQueryError = true;
  rerender(<SecuritySessions />);
  expect(screen.getByRole("alert")).toHaveTextContent(
    "Member sessions could not be loaded.",
  );
  expect(screen.queryByRole("table")).not.toBeInTheDocument();
  expect(
    screen.queryByText("No active sessions match this member filter."),
  ).not.toBeInTheDocument();
});

it("identifies the current session and keeps errors and pending protection in its confirmation", async () => {
  const { rerender } = render(<SecuritySessions />);
  await openActions("aaaaaaaa");
  fireEvent.click(screen.getByRole("menuitem", { name: "Revoke session" }));
  await screen.findByRole("dialog", { name: "Revoke session" });
  const dialog = () => screen.getByRole("dialog", { name: "Revoke session" });
  expect(
    within(dialog()).getByText(/Your current session will lose access/),
  ).toBeInTheDocument();
  expect(within(dialog()).queryByText(SESSION.id)).not.toBeInTheDocument();
  expect(within(dialog()).getByText("Chrome")).toBeInTheDocument();
  expect(
    within(dialog()).getByRole("button", { name: "Revoke access" }),
  ).toBeDisabled();
  fireEvent.change(within(dialog()).getByLabelText("Reason for revocation"), {
    target: { value: "  Lost device  " },
  });
  fireEvent.click(
    within(dialog()).getByRole("button", { name: "Revoke access" }),
  );
  expect(mockRevoke).toHaveBeenCalledWith(
    { id: SESSION.id, member: false, reason: "Lost device" },
    expect.anything(),
  );

  mockPending = true;
  rerender(<SecuritySessions />);
  expect(
    within(dialog()).getByRole("button", { name: "Cancel" }),
  ).toBeDisabled();
  fireEvent.keyDown(dialog(), { key: "Escape" });
  expect(dialog()).toBeInTheDocument();
  expect(mockRevoke).toHaveBeenCalledTimes(1);
  mockPending = false;
  mockError = new Error("Access could not be revoked.");
  rerender(<SecuritySessions />);
  expect(within(dialog()).getByRole("alert")).toHaveTextContent(
    "Access could not be revoked.",
  );
});

it("does not open a revocation dialog when an active session expires while its menu is open", async () => {
  render(<SecuritySessions />);
  await openActions("aaaaaaaa");
  const now = jest
    .spyOn(Date, "now")
    .mockReturnValue(Date.parse(SESSION.expiresAt));
  try {
    fireEvent.click(screen.getByRole("menuitem", { name: "Revoke session" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(mockRevoke).not.toHaveBeenCalled();
  } finally {
    now.mockRestore();
  }
});

it.each(["Cancel", "Escape"])(
  "returns keyboard focus to the row menu after %s",
  async (dismissal) => {
    render(<SecuritySessions />);
    const trigger = await openActions("cccccccc");
    fireEvent.keyDown(
      screen.getByRole("menuitem", { name: "Revoke session" }),
      { key: "Enter", code: "Enter" },
    );
    const dialog = await screen.findByRole("dialog", {
      name: "Revoke session",
    });
    await waitFor(() => {
      expect(
        within(dialog).getByLabelText("Reason for revocation"),
      ).toHaveFocus();
    });
    if (dismissal === "Cancel")
      fireEvent.click(within(dialog).getByRole("button", { name: "Cancel" }));
    else fireEvent.keyDown(dialog, { key: "Escape", code: "Escape" });
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      expect(screen.queryByRole("menu")).not.toBeInTheDocument();
      expect(trigger).toHaveFocus();
    });
    expect(mockRevoke).not.toHaveBeenCalled();
  },
);

it("returns focus to the member filter when a revoked session row disappears", async () => {
  const { rerender } = render(<SecuritySessions />);
  const trigger = await openActions("cccccccc");
  fireEvent.click(screen.getByRole("menuitem", { name: "Revoke session" }));
  const dialog = await screen.findByRole("dialog", { name: "Revoke session" });
  fireEvent.change(within(dialog).getByLabelText("Reason for revocation"), {
    target: { value: "Lost device" },
  });
  fireEvent.click(
    within(dialog).getByRole("button", { name: "Revoke access" }),
  );
  mockSessionItems = sessions.filter(
    (session) => !session.id.startsWith("cccccccc"),
  );
  rerender(<SecuritySessions />);
  expect(trigger.isConnected).toBe(false);
  act(() => {
    mockRevoke.mock.calls[0][1].onSuccess();
  });
  await waitFor(() => {
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Member" })).toHaveFocus();
  });
});
