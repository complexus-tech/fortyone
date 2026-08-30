/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  PropsWithChildren,
} from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { SubstorySuggestions } from "./substory-suggestions";

type ButtonProps = PropsWithChildren<
  ButtonHTMLAttributes<HTMLButtonElement> & {
    color?: string;
    variant?: string;
  }
>;

type ContainerProps = PropsWithChildren<{
  className?: string;
}>;

type FlexProps = ContainerProps & {
  align?: string;
  gap?: number;
  justify?: string;
};

type TextProps = ContainerProps & {
  color?: string;
};

type CheckboxProps = InputHTMLAttributes<HTMLInputElement> & {
  onCheckedChange?: (checked: boolean) => void;
};

jest.mock("ui", () => ({
  Box: ({ children, className }: ContainerProps) => (
    <div className={className}>{children}</div>
  ),
  Button: ({
    children,
    color: _color,
    variant: _variant,
    ...props
  }: ButtonProps) => (
    <button {...props} type="button">
      {children}
    </button>
  ),
  Checkbox: ({ onCheckedChange, ...props }: CheckboxProps) => (
    <input
      {...props}
      onChange={() => {
        onCheckedChange?.(!props.checked);
      }}
      type="checkbox"
    />
  ),
  Flex: ({
    align: _align,
    children,
    className,
    gap: _gap,
    justify: _justify,
  }: FlexProps) => <div className={className}>{children}</div>,
  Text: ({ children, className, color: _color }: TextProps) => (
    <span className={className}>{children}</span>
  ),
  Wrapper: ({ children, className }: ContainerProps) => (
    <div className={className}>{children}</div>
  ),
}));

jest.mock("icons", () => ({
  AiIcon: () => null,
  InfoIcon: () => null,
}));

const terms = {
  plural: "stories",
  pluralCapitalized: "Stories",
  singular: "story",
  singularCapitalized: "Story",
};

const renderSuggestions = ({
  canCreateSuggestedSubstories = true,
  isCreatingSuggestedSubstories = false,
  isShowingSuggestionError = false,
  onCancelSuggestions = jest.fn(),
  onCreateSelectedSubstories = jest.fn(),
  onRequestSuggestions = jest.fn(),
  onToggleSelectedSubstory = jest.fn(),
  selectedSubstories = new Set(["Plan onboarding"]),
  showSuggestions = true,
  suggestedSubstories = [{ title: "Plan onboarding" }],
}: Partial<Parameters<typeof SubstorySuggestions>[0]> = {}) => {
  render(
    <SubstorySuggestions
      SuggestionRow={({ children }) => <div>{children}</div>}
      canCreateSuggestedSubstories={canCreateSuggestedSubstories}
      isCreatingSuggestedSubstories={isCreatingSuggestedSubstories}
      isShowingSuggestionError={isShowingSuggestionError}
      onCancelSuggestions={onCancelSuggestions}
      onCreateSelectedSubstories={onCreateSelectedSubstories}
      onRequestSuggestions={onRequestSuggestions}
      onToggleSelectedSubstory={onToggleSelectedSubstory}
      selectedSubstories={selectedSubstories}
      showSuggestions={showSuggestions}
      suggestedSubstories={suggestedSubstories}
      terms={terms}
    />,
  );

  return {
    onCancelSuggestions,
    onCreateSelectedSubstories,
    onRequestSuggestions,
    onToggleSelectedSubstory,
  };
};

describe("SubstorySuggestions", () => {
  it("shows a retry state instead of stale suggestions after a stream failure", () => {
    const onRequestSuggestions = jest.fn();
    renderSuggestions({
      isShowingSuggestionError: true,
      onRequestSuggestions,
    });

    expect(
      screen.getByText(/Maya could not finish generating sub stories/i),
    ).toBeVisible();
    expect(screen.queryByText("Plan onboarding")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Try again" }));

    expect(onRequestSuggestions).toHaveBeenCalledTimes(1);
  });

  it("keeps the create action disabled until the streamed result is complete", () => {
    const onCreateSelectedSubstories = jest.fn();
    renderSuggestions({
      canCreateSuggestedSubstories: false,
      onCreateSelectedSubstories,
    });

    const createButton = screen.getByRole("button", {
      name: "Create 1 Sub Story",
    });
    expect(createButton).toBeDisabled();

    fireEvent.click(createButton);

    expect(onCreateSelectedSubstories).not.toHaveBeenCalled();
  });

  it("forwards selection changes for a complete suggestion", () => {
    const onToggleSelectedSubstory = jest.fn();
    renderSuggestions({ onToggleSelectedSubstory });

    fireEvent.click(screen.getByRole("checkbox"));

    expect(onToggleSelectedSubstory).toHaveBeenCalledWith("Plan onboarding");
  });
});
