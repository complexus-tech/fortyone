import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { post } from "api-client";
import { DesktopAuthorization } from "./client";

jest.mock("api-client", () => ({ post: jest.fn() }));
jest.mock("@/components/ui", () => ({ Logo: () => null }));
jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children, role }: HTMLAttributes<HTMLParagraphElement>) => (
    <p role={role}>{children}</p>
  ),
  Button: ({
    children,
    onClick,
    loading,
  }: ButtonHTMLAttributes<HTMLButtonElement> & { loading: boolean }) => (
    <button disabled={loading} onClick={onClick} type="button">
      {children}
    </button>
  ),
}));

const transaction = { state: "s".repeat(43), codeChallenge: "c".repeat(43) };
const mockAssign = jest.fn();
const originalLocation = window.location;

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(post).mockReset();
  Object.defineProperty(window, "location", {
    configurable: true,
    value: { ...originalLocation, assign: mockAssign },
  });
});
afterEach(() => {
  jest.restoreAllMocks();
  Object.defineProperty(window, "location", {
    configurable: true,
    value: originalLocation,
  });
});

it("issues and returns the code through the validated development callback", async () => {
  jest.replaceProperty(process, "env", {
    ...process.env,
    NODE_ENV: "development",
  });
  const redirectUri = "fortyone-desktop-dev://login";
  jest
    .mocked(post)
    .mockResolvedValue({
      data: { code: "k".repeat(43), state: transaction.state },
    });
  render(
    <DesktopAuthorization
      email="person@example.com"
      transaction={{ ...transaction, redirectUri }}
    />,
  );
  await act(async () => {
    fireEvent.click(screen.getByRole("button"));
  });
  expect(post).toHaveBeenCalledWith("auth/desktop/authorize", {
    ...transaction,
    redirectUri,
    codeChallengeMethod: "S256",
  });
  expect(mockAssign).toHaveBeenCalledWith(
    `${redirectUri}?code=${"k".repeat(43)}&state=${transaction.state}`,
  );
});

it("waits for explicit consent and guards repeated requests while authorizing", async () => {
  let resolve: (value: unknown) => void = () => {};
  jest.mocked(post).mockReturnValue(
    new Promise((finish) => {
      resolve = finish;
    }),
  );
  render(
    <DesktopAuthorization
      email="person@example.com"
      transaction={transaction}
    />,
  );
  expect(post).not.toHaveBeenCalled();
  expect(screen.getByText(/person@example.com/)).toBeInTheDocument();
  const button = screen.getByRole("button", {
    name: "Continue in the desktop app",
  });
  fireEvent.click(button);
  fireEvent.click(button);
  expect(post).toHaveBeenCalledTimes(1);
  expect(post).toHaveBeenCalledWith("auth/desktop/authorize", {
    ...transaction,
    codeChallengeMethod: "S256",
    redirectUri: "fortyone-desktop://login",
  });
  await act(async () => {
    resolve({ data: { code: "k".repeat(43), state: transaction.state } });
  });
  expect(mockAssign).toHaveBeenCalledWith(
    `fortyone-desktop://login?code=${"k".repeat(43)}&state=${transaction.state}`,
  );
  expect(screen.getByRole("status")).toHaveTextContent(
    "FortyOne should open shortly",
  );
});

it.each([
  { code: "k".repeat(43), state: "x".repeat(43) },
  { code: "unsafe&token=value", state: transaction.state },
])("rejects a mismatched or malformed handoff response", async (data) => {
  jest.mocked(post).mockResolvedValue({ data });
  render(
    <DesktopAuthorization
      email="person@example.com"
      transaction={transaction}
    />,
  );
  await act(async () => {
    fireEvent.click(screen.getByRole("button"));
  });
  expect(mockAssign).not.toHaveBeenCalled();
  expect(screen.getByRole("alert")).toHaveTextContent(/invalid/);
});

it("surfaces authorization failure and permits a deliberate retry", async () => {
  jest
    .mocked(post)
    .mockRejectedValue(new Error("Session expired. Sign in again."));
  render(
    <DesktopAuthorization
      email="person@example.com"
      transaction={transaction}
    />,
  );
  await act(async () => {
    fireEvent.click(screen.getByRole("button"));
  });
  expect(screen.getByRole("alert")).toHaveTextContent("Session expired");
  expect(screen.getByRole("button")).not.toBeDisabled();
});
