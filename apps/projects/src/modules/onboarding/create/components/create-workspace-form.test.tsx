/* global beforeEach, afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode, InputHTMLAttributes } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { useProfile } from "@/lib/hooks/profile";
import { checkWorkspaceAvailability } from "@/lib/queries/check-workspace-availability";
import { createWorkspaceAction } from "@/lib/actions/create-workspace";
import { updateProfile } from "@/lib/actions/update-profile";
import { CreateWorkspaceForm } from "./create-workspace-form";

jest.mock("@/lib/hooks/profile", () => ({ useProfile: jest.fn() }));
jest.mock("@/lib/queries/check-workspace-availability", () => ({
  checkWorkspaceAvailability: jest.fn(),
}));
jest.mock("@/lib/actions/create-workspace", () => ({
  createWorkspaceAction: jest.fn(),
}));
jest.mock("@/lib/actions/update-profile", () => ({ updateProfile: jest.fn() }));
jest.mock("icons", () => ({ CheckIcon: () => null, CloseIcon: () => null }));
jest.mock("ui", () => {
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );
  const Select = ({
    children,
    value,
    onValueChange,
  }: {
    children: ReactNode;
    value: string;
    onValueChange: (value: string) => void;
  }) => (
    <select
      id="team-size"
      onChange={(event) => {
        onValueChange(event.target.value);
      }}
      value={value}
    >
      {children}
    </select>
  );
  Select.Trigger = function SelectTrigger() {
    return null;
  };
  Select.Input = function SelectInput() {
    return null;
  };
  Select.Content = function SelectContent({
    children,
  }: {
    children: ReactNode;
  }) {
    return <>{children}</>;
  };
  Select.Group = Select.Content;
  Select.Option = function SelectOption({
    children,
    value,
  }: {
    children: ReactNode;
    value: string;
  }) {
    return <option value={value}>{children}</option>;
  };
  return {
    Box: Container,
    Flex: Container,
    Text: ({ children, role }: { children: ReactNode; role?: string }) => (
      <p role={role}>{children}</p>
    ),
    Input: ({
      label,
      helpText,
      hasError: _hasError,
      rightIcon: _rightIcon,
      ...props
    }: InputHTMLAttributes<HTMLInputElement> & {
      label: string;
      helpText?: string;
      hasError?: boolean;
      rightIcon?: ReactNode;
    }) => (
      <>
        <label htmlFor={props.name}>{label}</label>
        <input {...props} id={props.name} />
        {helpText ? <span>{helpText}</span> : null}
      </>
    ),
    Button: ({
      children,
      disabled,
      loading,
      loadingText,
      onClick,
      type = "button",
    }: {
      children: ReactNode;
      disabled?: boolean;
      loading?: boolean;
      loadingText?: string;
      onClick?: () => void;
      type?: "button" | "submit";
    }) => (
      <button
        disabled={disabled || loading}
        onClick={onClick}
        type={type === "submit" ? "submit" : "button"}
      >
        {loading ? loadingText : children}
      </button>
    ),
    Select,
  };
});

const profile = { id: "user-1", fullName: "Ada Lovelace" };
const profileMock = jest.mocked(useProfile);
beforeEach(() => {
  jest.useFakeTimers();
  jest.resetAllMocks();
  sessionStorage.clear();
  Element.prototype.scrollIntoView = jest.fn();
  profileMock.mockReturnValue({ data: profile, isError: false } as ReturnType<
    typeof useProfile
  >);
  jest.mocked(checkWorkspaceAvailability).mockResolvedValue({
    data: { available: true, slug: "acme" },
    error: { message: "" },
  });
});
afterEach(() => {
  jest.useRealTimers();
});

const enterWorkspace = async () => {
  fireEvent.change(screen.getByLabelText("Your Workspace"), {
    target: { value: "Acme" },
  });
  fireEvent.change(screen.getByLabelText("Workspace URL"), {
    target: { value: "acme-team" },
  });
  await act(async () => {
    jest.advanceTimersByTime(400);
  });
  fireEvent.click(
    screen.getByRole("button", { name: "Continue to your work" }),
  );
};

describe("CreateWorkspaceForm", () => {
  it("keeps answers and the selected start across completed-step navigation without creating anything", async () => {
    render(<CreateWorkspaceForm />);
    expect(screen.queryByLabelText("Your full name")).not.toBeInTheDocument();
    await enterWorkspace();
    expect(
      screen.getByRole("heading", { name: "What kind of work?" }),
    ).toHaveFocus();
    fireEvent.click(
      screen.getByRole("radio", { name: "Operations & delivery" }),
    );
    fireEvent.change(
      screen.getByLabelText("How many people will use this workspace?"),
      { target: { value: "11-50" } },
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Choose how to start" }),
    );
    fireEvent.click(
      screen.getByRole("radio", { name: /Explore with examples/ }),
    );
    expect(
      screen.getByRole("radio", { name: /Explore with examples/ }),
    ).toBeChecked();
    expect(
      screen.getByRole("radio", { name: /Create my first task/ }),
    ).not.toBeChecked();
    expect(screen.queryByText("Your example tasks")).not.toBeInTheDocument();
    expect(
      screen.queryByText("Document a recurring process"),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Step 2: Your work" }));
    expect(screen.getByRole("button", { name: "Back" })).toBeEnabled();
    expect(
      screen.getByRole("radio", { name: "Operations & delivery" }),
    ).toBeChecked();
    expect(
      screen.getByLabelText("How many people will use this workspace?"),
    ).toHaveValue("11-50");
    fireEvent.click(screen.getByRole("button", { name: "Step 1: Workspace" }));
    expect(screen.getByLabelText("Your Workspace")).toHaveValue("Acme");
    expect(screen.getByLabelText("Workspace URL")).toHaveValue("acme-team");
    fireEvent.click(
      screen.getByRole("button", { name: "Step 3: Get started" }),
    );
    expect(
      screen.getByRole("radio", { name: /Explore with examples/ }),
    ).toBeChecked();
    expect(createWorkspaceAction).not.toHaveBeenCalled();
    expect(updateProfile).not.toHaveBeenCalled();
  });

  it("asks for a missing name and surfaces failed final profile saves without losing answers", async () => {
    profileMock.mockReturnValue({
      data: { ...profile, fullName: "" },
      isError: false,
    } as ReturnType<typeof useProfile>);
    jest
      .mocked(updateProfile)
      .mockRejectedValue(new Error("Your name could not be saved"));
    render(<CreateWorkspaceForm />);
    fireEvent.change(screen.getByLabelText("Your full name"), {
      target: { value: "Ada Lovelace" },
    });
    await enterWorkspace();
    fireEvent.click(screen.getByRole("radio", { name: "Personal projects" }));
    fireEvent.click(
      screen.getByRole("button", { name: "Choose how to start" }),
    );
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Create Workspace" }));
    });
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Your name could not be saved",
    );
    expect(createWorkspaceAction).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Skip" })).toBeEnabled();
    fireEvent.click(screen.getByRole("button", { name: "Step 1: Workspace" }));
    expect(screen.getByLabelText("Your full name")).toHaveValue("Ada Lovelace");
  });

  it.each([/Explore with examples/, /Import existing work/])(
    "creates an empty workspace when Skip overrides %s",
    async (selectedStart) => {
      jest.mocked(updateProfile).mockResolvedValue({
        data: profile,
      } as Awaited<ReturnType<typeof updateProfile>>);
      jest.mocked(createWorkspaceAction).mockResolvedValue({
        error: { message: "Workspace creation is temporarily unavailable" },
      });
      render(<CreateWorkspaceForm />);
      await enterWorkspace();
      fireEvent.click(
        screen.getByRole("radio", { name: "Operations & delivery" }),
      );
      fireEvent.change(
        screen.getByLabelText("How many people will use this workspace?"),
        { target: { value: "11-50" } },
      );
      fireEvent.click(
        screen.getByRole("button", { name: "Choose how to start" }),
      );
      fireEvent.click(screen.getByRole("radio", { name: selectedStart }));
      expect(
        screen.queryByRole("button", { name: "Back" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", {
          name: "Start with an empty workspace",
        }),
      ).not.toBeInTheDocument();

      await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: "Skip" }));
      });

      expect(updateProfile).toHaveBeenCalledTimes(1);
      expect(createWorkspaceAction).toHaveBeenCalledTimes(1);
      expect(createWorkspaceAction).toHaveBeenCalledWith({
        name: "Acme",
        slug: "acme-team",
        teamSize: "11-50",
        workType: "operations",
        includeExamples: false,
      });
      expect(
        screen.getByRole("heading", { name: "Make your first move" }),
      ).toBeInTheDocument();
      expect(screen.getByRole("radio", { name: selectedStart })).toBeChecked();
      expect(screen.getByRole("alert")).toHaveTextContent(
        "Workspace creation is temporarily unavailable",
      );
      expect(screen.getByRole("button", { name: "Skip" })).toBeEnabled();
    },
  );
});
