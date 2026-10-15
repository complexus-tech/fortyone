/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
} from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { getDesktopAuthPath } from "@/lib/desktop-auth";
import { WorkspaceSSOSignIn } from "./workspace-sso";
import { WorkspaceSSORecovery } from "./workspace-sso-recovery";

jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.fortyone.app",
}));
jest.mock("@/components/ui/logo", () => ({ Logo: () => null }));
jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Input: ({
    label,
    ...props
  }: InputHTMLAttributes<HTMLInputElement> & { label: string }) => (
    <input aria-label={label} {...props} />
  ),
  Button: ({
    children,
    disabled,
    onClick,
    type,
  }: ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button
      disabled={disabled}
      onClick={onClick}
      type={type === "submit" ? "submit" : "button"}
    >
      {children}
    </button>
  ),
}));
const originalLocation = window.location;
const assign = jest.fn<undefined, [string]>();
const callback = getDesktopAuthPath({
  state: "s".repeat(43),
  codeChallenge: "c".repeat(43),
});
beforeEach(() => {
  assign.mockReset();
  Object.defineProperty(window, "location", {
    configurable: true,
    value: { ...originalLocation, assign },
  });
});
afterEach(() => {
  jest.restoreAllMocks();
  Object.defineProperty(window, "location", {
    configurable: true,
    value: originalLocation,
  });
});

it.each(["development", "production"] as const)(
  "validates a desktop development return target in %s",
  (environment) => {
    jest.replaceProperty(process, "env", {
      ...process.env,
      NODE_ENV: environment,
    });
    const callbackUrl = getDesktopAuthPath({
      state: "s".repeat(43),
      codeChallenge: "c".repeat(43),
      redirectUri: "fortyone-desktop-dev://login",
    });
    render(<WorkspaceSSOSignIn callbackUrl={callbackUrl} />);
    fireEvent.change(screen.getByLabelText("Workspace URL name"), {
      target: { value: "acme" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Continue with SSO" }));
    expect(new URL(assign.mock.calls[0][0]).searchParams.get("returnTo")).toBe(
      environment === "development" ? callbackUrl : null,
    );
  },
);

it("preserves the browser desktop transaction through workspace SSO", () => {
  render(<WorkspaceSSOSignIn callbackUrl={callback} />);
  fireEvent.change(screen.getByLabelText("Workspace URL name"), {
    target: { value: "acme" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Continue with SSO" }));
  const url = new URL(assign.mock.calls[0][0]);
  expect(url.origin).toBe("https://api.fortyone.app");
  expect(url.pathname).toBe("/auth/sso/acme");
  expect(url.searchParams.get("returnTo")).toBe(callback);
});
it("preserves the same transaction when connecting required workspace SSO", () => {
  render(
    <WorkspaceSSORecovery
      callbackUrl={callback}
      email="user@example.com"
      required
      workspaceName="Acme"
      workspaceSlug="acme"
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Continue with SSO" }));
  expect(new URL(assign.mock.calls[0][0]).searchParams.get("returnTo")).toBe(
    callback,
  );
});
it.each([
  "https://attacker.invalid",
  `${callback}&redirect=https://attacker.invalid`,
  `${callback}#extra`,
  "/my-work",
])("does not forward an arbitrary SSO return target: %s", (callbackUrl) => {
  render(
    <WorkspaceSSORecovery
      callbackUrl={callbackUrl}
      email="user@example.com"
      required
      workspaceName="Acme"
      workspaceSlug="acme"
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Continue with SSO" }));
  expect(new URL(assign.mock.calls[0][0]).searchParams.has("returnTo")).toBe(
    false,
  );
});
