/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type {
  ElementType,
  HTMLAttributes,
  InputHTMLAttributes,
  PropsWithChildren,
  ReactNode,
} from "react";
import { useTerminology } from "@/hooks";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useSearch } from "@/modules/search/hooks/use-search";
import { useAddAssociationMutation } from "../hooks/add-association-mutation";
import { useCreateStoryMutation } from "../hooks/create-mutation";
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

jest.mock("@/hooks", () => ({
  useTerminology: jest.fn(),
}));

jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: jest.fn(),
}));

jest.mock("../hooks/add-association-mutation", () => ({
  useAddAssociationMutation: jest.fn(),
}));

jest.mock("../hooks/create-mutation", () => ({
  useCreateStoryMutation: jest.fn(),
}));

const mockedUseSearch = jest.mocked(useSearch);
const mockedUseTerminology = jest.mocked(useTerminology);
const mockedUseTeamStatuses = jest.mocked(useTeamStatuses);
const mockedUseAddAssociationMutation = jest.mocked(useAddAssociationMutation);
const mockedUseCreateStoryMutation = jest.mocked(useCreateStoryMutation);

describe("StoryRelationshipPicker", () => {
  beforeEach(() => {
    mockedUseTerminology.mockReturnValue({
      getTermDisplay: jest.fn(() => "ticket"),
    } as unknown as ReturnType<typeof useTerminology>);
    mockedUseTeamStatuses.mockReturnValue({
      data: [{ color: "#22c55e", id: "status-1", isDefault: true }],
    } as unknown as ReturnType<typeof useTeamStatuses>);
    mockedUseCreateStoryMutation.mockReturnValue({
      isPending: false,
      mutate: jest.fn(),
    } as unknown as ReturnType<typeof useCreateStoryMutation>);
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

    fireEvent.click(screen.getByRole("button", { name: /association/i }));
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

    fireEvent.click(screen.getByRole("button", { name: /association/i }));
    fireEvent.change(screen.getByPlaceholderText(/search story title or id/i), {
      target: { value: "rr" },
    });

    expect(screen.getAllByTestId("relationship-search-skeleton")).toHaveLength(
      2,
    );
  });

  it("creates a related story from the empty search state and associates it", () => {
    const addAssociation = jest.fn();
    const createStory = jest.fn((_payload, options) => {
      options.onSuccess({ id: "story-3" });
    });

    mockedUseAddAssociationMutation.mockReturnValue({
      isPending: false,
      mutate: addAssociation,
    } as unknown as ReturnType<typeof useAddAssociationMutation>);
    mockedUseCreateStoryMutation.mockReturnValue({
      isPending: false,
      mutate: createStory,
    } as unknown as ReturnType<typeof useCreateStoryMutation>);
    mockedUseSearch.mockReturnValue({
      data: undefined,
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

    fireEvent.click(screen.getByRole("button", { name: /association/i }));
    expect(screen.getByText(/search for an existing ticket or/i)).toBeVisible();

    fireEvent.click(
      screen.getByRole("button", { name: /create related ticket/i }),
    );

    expect(createStory).toHaveBeenCalledWith(
      {
        priority: "No Priority",
        statusId: "status-1",
        teamId: "team-1",
        title: "Related ticket",
      },
      expect.any(Object),
    );
    expect(addAssociation).toHaveBeenCalledWith(
      {
        fromStoryId: "story-1",
        toStoryId: "story-3",
        type: "related",
      },
      expect.any(Object),
    );

    fireEvent.change(screen.getByPlaceholderText(/search story title or id/i), {
      target: { value: "rr" },
    });

    expect(
      screen.queryByRole("button", { name: /create related ticket/i }),
    ).not.toBeInTheDocument();
  });
});
