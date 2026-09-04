/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen } from "@testing-library/react";
import { JOIN_ONBOARDING_STEPS, OnboardingStepper } from "./onboarding-stepper";

jest.mock("icons", () => ({ CheckIcon: () => <svg aria-hidden="true" /> }));

describe("OnboardingStepper", () => {
  it("keeps the existing workspace stages and visited-step navigation by default", () => {
    const onStepChange = jest.fn();
    render(<OnboardingStepper currentStep={1} onStepChange={onStepChange} />);
    expect(screen.getByText("Step 2 of 3: Your work")).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(3);
    fireEvent.click(screen.getByRole("button", { name: "Step 1: Workspace" }));
    expect(onStepChange).toHaveBeenCalledWith(0);
    expect(
      screen.queryByRole("button", { name: "Step 3: Get started" }),
    ).not.toBeInTheDocument();
  });

  it("allows returning to a visited stage after navigating back", () => {
    const onStepChange = jest.fn();
    render(
      <OnboardingStepper
        currentStep={0}
        furthestStep={2}
        onStepChange={onStepChange}
      />,
    );
    expect(screen.getByText("Step 1 of 3: Workspace")).toBeInTheDocument();
    const items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(3);
    expect(items[0]).toHaveAttribute("aria-current", "step");
    fireEvent.click(
      screen.getByRole("button", { name: "Step 3: Get started" }),
    );
    expect(onStepChange).toHaveBeenCalledWith(2);
  });

  it("shows the invited-user stages as status when no navigation is supplied", () => {
    render(<OnboardingStepper currentStep={1} steps={JOIN_ONBOARDING_STEPS} />);
    expect(screen.getByText("Step 2 of 3: Your profile")).toBeInTheDocument();
    expect(screen.getByText("Join workspace")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
