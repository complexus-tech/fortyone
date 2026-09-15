/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
} from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { requestMagicEmail } from "@/lib/actions/request-magic-email";
import { signInWithGoogle, signInWithMicrosoft } from "@/lib/actions/sign-in";
import { getMobileAuthPath } from "@/lib/mobile-auth";
import { withCallbackUrl } from "@/utils/callback-url";
import { AuthLayout } from "./index";

const mockPush = jest.fn<undefined, [string]>();
jest.mock("next/navigation", () => ({ useRouter: () => ({ push: mockPush }) }));
jest.mock("@/lib/actions/request-magic-email", () => ({
  requestMagicEmail: jest.fn(),
}));
jest.mock("@/lib/actions/sign-in", () => ({
  signInWithGoogle: jest.fn(),
  signInWithMicrosoft: jest.fn(),
}));
jest.mock("@/components/ui", () => ({
  Logo: () => null,
  GoogleIcon: () => null,
  MicrosoftIcon: () => null,
}));
jest.mock("@/components/ui/otp-input", () => ({
  OTPInput: ({
    value,
    onChange,
  }: {
    value: string;
    onChange: (value: string) => void;
  }) => (
    <input
      aria-label="Verification code"
      onChange={(event) => {
        onChange(event.target.value);
      }}
      value={value}
    />
  ),
}));
jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Box: Container,
    Flex: Container,
    Text: Container,
    Input: ({
      label,
      helpText,
      name,
      onChange,
      value,
    }: InputHTMLAttributes<HTMLInputElement> & {
      label: string;
      helpText?: string;
    }) => (
      <>
        <label htmlFor={name}>{label}</label>
        <input id={name} name={name} onChange={onChange} value={value} />
        {helpText ? <span>{helpText}</span> : null}
      </>
    ),
    Button: ({
      children,
      disabled,
      loading,
      onClick,
      type = "button",
    }: ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => (
      <button
        disabled={disabled || loading}
        onClick={onClick}
        type={type === "submit" ? "submit" : "button"}
      >
        {children}
      </button>
    ),
  };
});

const callback = getMobileAuthPath({
  state: "s".repeat(43),
  codeChallenge: "c".repeat(43),
});

beforeEach(() => {
  jest.clearAllMocks();
  jest
    .mocked(requestMagicEmail)
    .mockResolvedValue({ data: null, error: { message: null } });
  jest.mocked(signInWithGoogle).mockResolvedValue(undefined);
  jest.mocked(signInWithMicrosoft).mockResolvedValue(undefined);
});

describe("mobile email-only login", () => {
  it("shows a static deletion confirmation and optional pending calendar cleanup", () => {
    const { rerender } = render(
      <AuthLayout accountDeleted cleanupPending page="login" />,
    );
    expect(
      screen.getByText("Your account has been deleted."),
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        "Connected-service cleanup will continue in the background.",
      ),
    ).toBeInTheDocument();
    rerender(<AuthLayout cleanupPending page="login" />);
    expect(
      screen.queryByText("Your account has been deleted."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText(
        "Connected-service cleanup will continue in the background.",
      ),
    ).not.toBeInTheDocument();
  });
  it.each(["login", "signup"] as const)(
    "does not offer social providers on mobile %s",
    (page) => {
      render(<AuthLayout isMobileApp page={page} />);
      expect(screen.getByLabelText("Enter your email")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Continue with Google" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Continue with Microsoft" }),
      ).not.toBeInTheDocument();
      expect(screen.queryByText("OR")).not.toBeInTheDocument();
    },
  );

  it("infers mobile mode from the handoff and preserves it through OTP verification", async () => {
    render(<AuthLayout callbackUrl={callback} page="login" />);
    expect(
      screen.queryByRole("button", { name: "Continue with Google" }),
    ).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Enter your email"), {
      target: { value: "member@example.com" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Continue" }));
    await waitFor(() => {
      expect(requestMagicEmail).toHaveBeenCalledWith(
        "member@example.com",
        true,
        callback,
      );
    });
    await screen.findByText(/A secure sign-in code has been sent/);
    fireEvent.change(screen.getByLabelText("Verification code"), {
      target: { value: "123456" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Verify Code" }));
    const destination = new URL(
      mockPush.mock.calls[0][0],
      "https://cloud.fortyone.app",
    );
    expect(destination.pathname).toBe("/verify/member%40example.com/123456");
    expect(destination.searchParams.get("mobileApp")).toBe("true");
    expect(destination.searchParams.get("callbackUrl")).toBe(callback);
    expect(signInWithGoogle).not.toHaveBeenCalled();
    expect(signInWithMicrosoft).not.toHaveBeenCalled();
  });

  it("keeps signup and error retries email-only through nested onboarding callbacks", () => {
    const continuation = withCallbackUrl(
      "/onboarding/join?token=invitation",
      callback,
    );
    render(
      <AuthLayout
        callbackUrl={continuation}
        errorMessage="Your code expired. Try again."
        page="signup"
      />,
    );
    expect(
      screen.queryByRole("button", { name: "Continue with Microsoft" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText("Your code expired. Try again."),
    ).toBeInTheDocument();
    const href = screen
      .getByRole("link", { name: "Sign in" })
      .getAttribute("href")!;
    const destination = new URL(href, "https://cloud.fortyone.app");
    expect(destination.searchParams.get("mobileApp")).toBe("true");
    expect(destination.searchParams.get("callbackUrl")).toBe(continuation);
  });

  it("preserves desktop social providers and their callback destinations", async () => {
    render(<AuthLayout callbackUrl="/my-work" page="login" />);
    fireEvent.click(
      screen.getByRole("button", { name: "Continue with Google" }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Continue with Microsoft" }),
    );
    await waitFor(() => {
      expect(signInWithGoogle).toHaveBeenCalledWith(
        "/auth-callback?callbackUrl=%2Fmy-work",
      );
      expect(signInWithMicrosoft).toHaveBeenCalledWith(
        "/auth-callback?callbackUrl=%2Fmy-work",
      );
    });
    expect(screen.getByRole("link", { name: "Create one" })).toHaveAttribute(
      "href",
      "/signup?callbackUrl=%2Fmy-work",
    );
  });
});
