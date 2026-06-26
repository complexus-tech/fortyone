/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type {
  ElementType,
  HTMLAttributes,
  InputHTMLAttributes,
  PropsWithChildren,
  ReactNode,
} from "react";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useSearch } from "@/modules/search/hooks/use-search";
import { useAddAssociationMutation } from "../hooks/add-association-mutation";
import { StoryRelationshipPicker } from "./story-relationship-picker";

jest.mock("ui", () => {
  function MockPopover({ children }: PropsWithChildren) {
    return <>{children}</>;
  }
  MockPopover.Trigger = function MockPopoverTrigger({
    children,
  }: PropsWithChildren) {
    return <>{children}</>;
  };
  MockPopover.Content = function MockPopoverContent({
    children,
  }: PropsWithChildren) {
    return <div>{children}</div>;
  };

  return {
    Box: ({
      as: Component = "div",
      children,
      ...props
    }: PropsWithChildren<{ as?: ElementType }>) => (
      <Component {...props}>{children}</Component>
    ),
    Button: ({
      active: _active,
      align: _align,
      asIcon: _asIcon,
      children,
      color: _color,
      fullWidth: _fullWidth,
      leftIcon,
      loading,
      rounded: _rounded,
      rightIcon,
      size: _size,
      variant: _variant,
      ...props
    }: PropsWithChildren<{
      active?: boolean;
      align?: string;
      asIcon?: boolean;
      color?: string;
      fullWidth?: boolean;
      leftIcon?: ReactNode;
      loading?: boolean;
      rounded?: string;
      rightIcon?: ReactNode;
      size?: string;
      variant?: string;
    }>) => (
      <button type="button" {...props}>
        {loading ? "Loading..." : null}
        {leftIcon}
        {children}
        {rightIcon}
      </button>
    ),
    Flex: ({
      align: _align,
      as: Component = "div",
      children,
      direction: _direction,
      gap: _gap,
      justify: _justify,
      wrap: _wrap,
      ...props
    }: PropsWithChildren<{
      align?: string;
      as?: ElementType;
      direction?: string;
      gap?: number;
      justify?: string;
      wrap?: boolean;
    }>) => <Component {...props}>{children}</Component>,
    Input: ({
      leftIcon,
      rightIcon,
      ...props
    }: InputHTMLAttributes<HTMLInputElement> & {
      leftIcon?: ReactNode;
      rightIcon?: ReactNode;
    }) => (
      <label>
        {leftIcon}
        <input {...props} />
        {rightIcon}
      </label>
    ),
    Popover: MockPopover,
    Skeleton: (props: HTMLAttributes<HTMLDivElement>) => <div {...props} />,
    Text: ({
      as: Component = "p",
      children,
      color: _color,
      fontSize: _fontSize,
      fontWeight: _fontWeight,
      ...props
    }: PropsWithChildren<{
      as?: ElementType;
      color?: string;
      fontSize?: string;
      fontWeight?: string;
    }>) => <Component {...props}>{children}</Component>,
  };
});

jest.mock("@/modules/search/hooks/use-search", () => ({
  useSearch: jest.fn(),
}));

jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: jest.fn(),
}));

jest.mock("../hooks/add-association-mutation", () => ({
  useAddAssociationMutation: jest.fn(),
}));

const mockedUseSearch = jest.mocked(useSearch);
const mockedUseTeamStatuses = jest.mocked(useTeamStatuses);
const mockedUseAddAssociationMutation = jest.mocked(useAddAssociationMutation);

describe("StoryRelationshipPicker", () => {
  beforeEach(() => {
    mockedUseTeamStatuses.mockReturnValue({
      data: [{ color: "#22c55e", id: "status-1" }],
    } as unknown as ReturnType<typeof useTeamStatuses>);
  });

  it("searches stories in the same team and creates a reversed blocking association for blocked-by", () => {
    const mutate = jest.fn();

    mockedUseAddAssociationMutation.mockReturnValue({
      isPending: false,
      mutate,
    } as unknown as ReturnType<typeof useAddAssociationMutation>);
    mockedUseSearch.mockReturnValue({
      data: {
        objectives: [],
        page: 1,
        pageSize: 8,
        stories: [
          {
            id: "story-2",
            sequenceId: 62,
            statusId: "status-1",
            title: "rrr",
            teamId: "team-1",
          },
        ],
        totalObjectives: 0,
        totalPages: 1,
        totalStories: 1,
      },
      isFetching: false,
    } as unknown as ReturnType<typeof useSearch>);

    render(
      <StoryRelationshipPicker
        currentStoryId="story-1"
        currentStoryTitle="Current story"
        teamCode="QIT"
        teamId="team-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /associate/i }));
    fireEvent.click(screen.getByRole("button", { name: /blocked by/i }));
    fireEvent.change(screen.getByPlaceholderText(/search story title or id/i), {
      target: { value: "rr" },
    });

    expect(mockedUseSearch).toHaveBeenLastCalledWith({
      pageSize: 8,
      query: "rr",
      teamId: "team-1",
      type: "stories",
    });

    fireEvent.click(screen.getByRole("button", { name: /QIT-62 rrr/i }));

    expect(mutate).toHaveBeenCalledWith(
      {
        fromStoryId: "story-2",
        toStoryId: "story-1",
        type: "blocking",
      },
      expect.any(Object),
    );
  });

  it("shows two story-shaped skeleton rows while search results load", () => {
    mockedUseAddAssociationMutation.mockReturnValue({
      isPending: false,
      mutate: jest.fn(),
    } as unknown as ReturnType<typeof useAddAssociationMutation>);
    mockedUseSearch.mockReturnValue({
      data: undefined,
      isFetching: true,
    } as unknown as ReturnType<typeof useSearch>);

    render(
      <StoryRelationshipPicker
        currentStoryId="story-1"
        currentStoryTitle="Current story"
        teamCode="QIT"
        teamId="team-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /associate/i }));
    fireEvent.change(screen.getByPlaceholderText(/search story title or id/i), {
      target: { value: "rr" },
    });

    expect(screen.getAllByTestId("relationship-search-skeleton")).toHaveLength(
      2,
    );
  });
});
