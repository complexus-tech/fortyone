/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
} from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { useRouter } from "next/navigation";
import { inviteOnboardingMembers } from "@/modules/invitations/public/onboarding";
import type { Workspace } from "@/types/workspace";
import { InviteTeam } from "./index";

jest.mock("next/navigation", () => ({ useRouter: jest.fn() }));
jest.mock("@/modules/invitations/public/onboarding", () => ({
  inviteOnboardingMembers: jest.fn(),
}));
jest.mock("@/components/ui/logo", () => ({ Logo: () => null }));
jest.mock("@/utils", () => ({ buildWorkspaceUrl: jest.fn() }));
jest.mock("icons", () => ({
  CheckIcon: () => null,
  CloseIcon: () => null,
  InfoIcon: () => null,
  PlusIcon: () => null,
}));
jest.mock("ui", () => {
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Box: Container,
    Flex: Container,
    Text: Container,
    Input: (props: InputHTMLAttributes<HTMLInputElement>) => (
      <input {...props} />
    ),
    Button: ({
      children,
      disabled,
      loading,
      loadingText,
      onClick,
      type = "button",
    }: ButtonHTMLAttributes<HTMLButtonElement> & {
      loading?: boolean;
      loadingText?: string;
    }) => (
      <button
        disabled={disabled || loading}
        onClick={onClick}
        type={type === "submit" ? "submit" : "button"}
      >
        {loading ? loadingText : children}
      </button>
    ),
  };
});

const push = jest.fn();
const router: ReturnType<typeof useRouter> = {
  back: jest.fn(),
  forward: jest.fn(),
  refresh: jest.fn(),
  push,
  replace: jest.fn(),
  prefetch: jest.fn(),
  bfcacheId: "onboarding-invite",
};

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useRouter).mockReturnValue(router);
});

describe("InviteTeam", () => {
  it("skips invitations and preserves the next destination even when an email is filled in", () => {
    render(
      <InviteTeam
        activeWorkspace={{ slug: "acme" } as Workspace}
        callbackUrl="/my-work?view=board"
        teams={[]}
      />,
    );
    expect(
      screen.getByRole("button", { name: "Invite members" }),
    ).toBeDisabled();
    const [emailInput] = screen.getAllByPlaceholderText(
      "colleague@company.com",
    );
    fireEvent.change(emailInput, { target: { value: "ada@example.com" } });
    expect(
      screen.getByRole("button", { name: "Invite members" }),
    ).toBeEnabled();

    fireEvent.click(screen.getByRole("button", { name: "Skip" }));

    expect(inviteOnboardingMembers).not.toHaveBeenCalled();
    expect(push).toHaveBeenCalledTimes(1);
    expect(push).toHaveBeenCalledWith(
      "/onboarding/welcome?callbackUrl=%2Fmy-work%3Fview%3Dboard",
    );
  });
});
