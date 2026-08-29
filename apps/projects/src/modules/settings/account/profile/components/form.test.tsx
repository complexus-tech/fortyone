/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import type { User } from "@/types";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";
import { ProfileForm } from "./form";

jest.mock("@/lib/hooks/profile", () => ({
  useProfile: jest.fn(),
}));

jest.mock("@/lib/hooks/update-profile-mutation", () => ({
  useUpdateProfileMutation: jest.fn(),
}));

jest.mock("@/modules/settings/components", () => ({
  SectionHeader: ({ title }: { title: string }) => <h2>{title}</h2>,
}));

jest.mock("./profile-picture", () => ({
  ProfilePicture: () => <div>Profile picture</div>,
}));

jest.mock("ui", () => ({
  Box: ({ children, ...props }: ReactTypes.HTMLAttributes<HTMLDivElement>) => (
    <div {...props}>{children}</div>
  ),
  Button: ({
    children,
    loading: _loading,
    loadingText: _loadingText,
    type = "button",
    ...props
  }: ReactTypes.ButtonHTMLAttributes<HTMLButtonElement> & {
    loading?: boolean;
    loadingText?: string;
  }) => (
    <button {...props} type={type === "submit" ? "submit" : "button"}>
      {children}
    </button>
  ),
  Input: ({
    label,
    ...props
  }: ReactTypes.InputHTMLAttributes<HTMLInputElement> & { label: string }) => (
    <label>
      {label}
      <input aria-label={label} {...props} />
    </label>
  ),
}));

const profile: User = {
  avatarUrl: null,
  createdAt: "2026-07-20T08:00:00.000Z",
  email: "ada@example.com",
  fullName: "Ada Ndlovu",
  githubUsername: null,
  hasSeenWalkthrough: true,
  id: "user-1",
  isActive: true,
  isInternal: false,
  isSystem: false,
  lastUsedWorkspaceId: "workspace-1",
  timezone: "Africa/Harare",
  updatedAt: "2026-07-20T08:00:00.000Z",
  username: "ada",
};

const useProfileMock = jest.mocked(useProfile);
const useUpdateProfileMutationMock = jest.mocked(useUpdateProfileMutation);
const updateProfileMock = jest.fn();

describe("ProfileForm", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useProfileMock.mockReturnValue({ data: profile } as ReturnType<
      typeof useProfile
    >);
    useUpdateProfileMutationMock.mockReturnValue({
      isPending: false,
      mutate: updateProfileMock,
    } as unknown as ReturnType<typeof useUpdateProfileMutation>);
  });

  it("preserves edits made while an earlier profile save is pending", () => {
    let completeSave:
      | ((response: { error?: { message: string } }) => void)
      | undefined;
    updateProfileMock.mockImplementation(
      (
        _updates: unknown,
        options?: {
          onSuccess?: (response: { error?: { message: string } }) => void;
        },
      ) => {
        completeSave = options?.onSuccess;
      },
    );
    render(<ProfileForm initialProfile={profile} />);

    const fullNameInput = screen.getByLabelText("Full name");
    fireEvent.change(fullNameInput, { target: { value: "Ada Moyo" } });
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));

    expect(updateProfileMock).toHaveBeenCalledWith(
      { fullName: "Ada Moyo", username: "ada" },
      expect.any(Object),
    );

    fireEvent.change(fullNameInput, { target: { value: "Ada Moyo-Ndlovu" } });
    act(() => {
      completeSave?.({});
    });

    expect(fullNameInput).toHaveValue("Ada Moyo-Ndlovu");
    expect(screen.getByRole("button", { name: "Save changes" })).toBeEnabled();
  });
});
