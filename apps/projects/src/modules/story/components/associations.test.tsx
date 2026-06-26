/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type {
  ElementType,
  HTMLAttributes,
  PropsWithChildren,
  ReactNode,
} from "react";
import { useWorkspacePath } from "@/hooks";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useRemoveAssociationMutation } from "@/modules/story/hooks/remove-association-mutation";
import { useUpdateAssociationMutation } from "@/modules/story/hooks/update-association-mutation";
import { Associations } from "./associations";

jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children, href }: PropsWithChildren<{ href: string }>) => (
    <a href={href}>{children}</a>
  ),
}));

jest.mock("@/components/ui", () => {
  function MockRowWrapper({
    children,
    ...props
  }: PropsWithChildren<HTMLAttributes<HTMLDivElement>>) {
    return <div {...props}>{children}</div>;
  }

  return {
    RowWrapper: MockRowWrapper,
  };
});

jest.mock("ui", () => {
  function MockMenu({ children }: PropsWithChildren) {
    return <>{children}</>;
  }
  MockMenu.Button = function MockMenuButton({ children }: PropsWithChildren) {
    return <>{children}</>;
  };
  MockMenu.Items = function MockMenuItems({ children }: PropsWithChildren) {
    return <div>{children}</div>;
  };
  MockMenu.Group = function MockMenuGroup({ children }: PropsWithChildren) {
    return <div>{children}</div>;
  };
  MockMenu.Separator = function MockMenuSeparator(
    props: HTMLAttributes<HTMLHRElement>,
  ) {
    return <hr {...props} />;
  };
  MockMenu.Item = function MockMenuItem({
    active: _active,
    children,
    className,
    onSelect,
  }: PropsWithChildren<{
    active?: boolean;
    className?: string;
    onSelect?: () => void;
  }>) {
    return (
      <button className={className} onClick={onSelect} type="button">
        {children}
      </button>
    );
  };

  return {
    Badge: ({
      children,
      color: _color,
      rounded: _rounded,
      ...props
    }: PropsWithChildren<{
      color?: string;
      rounded?: string;
    }>) => <span {...props}>{children}</span>,
    Box: ({
      as: Component = "div",
      children,
      ...props
    }: PropsWithChildren<{ as?: ElementType }>) => (
      <Component {...props}>{children}</Component>
    ),
    Button: ({
      asIcon: _asIcon,
      children,
      color: _color,
      leftIcon,
      rounded: _rounded,
      rightIcon,
      size: _size,
      variant: _variant,
      ...props
    }: PropsWithChildren<{
      asIcon?: boolean;
      color?: string;
      leftIcon?: ReactNode;
      rounded?: string;
      rightIcon?: ReactNode;
      size?: string;
      variant?: string;
    }>) => (
      <button type="button" {...props}>
        {leftIcon}
        {children}
        {rightIcon}
      </button>
    ),
    Flex: ({
      align: _align,
      as: Component = "div",
      children,
      gap: _gap,
      justify: _justify,
      ...props
    }: PropsWithChildren<{
      align?: string;
      as?: ElementType;
      gap?: number;
      justify?: string;
    }>) => <Component {...props}>{children}</Component>,
    Menu: MockMenu,
    Text: ({
      as: Component = "p",
      children,
      color: _color,
      fontSize: _fontSize,
      fontWeight: _fontWeight,
      transform: _transform,
      ...props
    }: PropsWithChildren<{
      as?: ElementType;
      color?: string;
      fontSize?: string;
      fontWeight?: string;
      transform?: string;
    }>) => <Component {...props}>{children}</Component>,
  };
});

jest.mock("@/hooks", () => ({
  useWorkspacePath: jest.fn(),
}));

jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: jest.fn(),
}));

jest.mock("@/modules/story/hooks/remove-association-mutation", () => ({
  useRemoveAssociationMutation: jest.fn(),
}));

jest.mock("@/modules/story/hooks/update-association-mutation", () => ({
  useUpdateAssociationMutation: jest.fn(),
}));

const mockedUseWorkspacePath = jest.mocked(useWorkspacePath);
const mockedUseTeams = jest.mocked(useTeams);
const mockedUseRemoveAssociationMutation = jest.mocked(
  useRemoveAssociationMutation,
);
const mockedUseUpdateAssociationMutation = jest.mocked(
  useUpdateAssociationMutation,
);

describe("Associations", () => {
  beforeEach(() => {
    mockedUseWorkspacePath.mockReturnValue({
      withWorkspace: (path: string) => path,
    } as ReturnType<typeof useWorkspacePath>);
    mockedUseTeams.mockReturnValue({
      data: [{ code: "QIT", id: "team-1" }],
    } as unknown as ReturnType<typeof useTeams>);
    mockedUseRemoveAssociationMutation.mockReturnValue({
      mutateAsync: jest.fn(),
    } as unknown as ReturnType<typeof useRemoveAssociationMutation>);
  });

  it("updates the association type from the menu", () => {
    const updateAssociation = jest.fn();

    mockedUseUpdateAssociationMutation.mockReturnValue({
      isPending: false,
      mutate: updateAssociation,
    } as unknown as ReturnType<typeof useUpdateAssociationMutation>);

    render(
      <Associations
        associations={[
          {
            fromStoryId: "story-1",
            id: "association-1",
            story: {
              id: "story-2",
              archivedAt: null,
              assigneeId: null,
              completedAt: null,
              createdAt: "2026-06-26T00:00:00.000Z",
              deletedAt: null,
              description: "",
              endDate: null,
              epicId: null,
              estimateLabel: null,
              estimateScheme: "points",
              estimateValue: null,
              keyResultId: null,
              labels: [],
              objectiveId: null,
              priority: "No Priority",
              reporterId: "user-1",
              sequenceId: 62,
              sprintId: null,
              startDate: null,
              statusId: "status-1",
              subStories: [],
              teamId: "team-1",
              title: "Blocked story",
              updatedAt: "2026-06-26T00:00:00.000Z",
              workspaceId: "workspace-1",
            },
            toStoryId: "story-2",
            type: "related",
          },
        ]}
        isAssociationsOpen
        setIsAssociationsOpen={jest.fn()}
        storyId="story-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /blocks/i }));

    expect(updateAssociation).toHaveBeenCalledWith({
      associationId: "association-1",
      fromStoryId: "story-1",
      storyId: "story-1",
      toStoryId: "story-2",
      type: "blocking",
    });
  });

  it("reverses the association direction for blocked-by", () => {
    const updateAssociation = jest.fn();

    mockedUseUpdateAssociationMutation.mockReturnValue({
      isPending: false,
      mutate: updateAssociation,
    } as unknown as ReturnType<typeof useUpdateAssociationMutation>);

    render(
      <Associations
        associations={[
          {
            fromStoryId: "story-1",
            id: "association-1",
            story: {
              id: "story-2",
              archivedAt: null,
              assigneeId: null,
              completedAt: null,
              createdAt: "2026-06-26T00:00:00.000Z",
              deletedAt: null,
              description: "",
              endDate: null,
              epicId: null,
              estimateLabel: null,
              estimateScheme: "points",
              estimateValue: null,
              keyResultId: null,
              labels: [],
              objectiveId: null,
              priority: "No Priority",
              reporterId: "user-1",
              sequenceId: 62,
              sprintId: null,
              startDate: null,
              statusId: "status-1",
              subStories: [],
              teamId: "team-1",
              title: "Blocked story",
              updatedAt: "2026-06-26T00:00:00.000Z",
              workspaceId: "workspace-1",
            },
            toStoryId: "story-2",
            type: "related",
          },
        ]}
        isAssociationsOpen
        setIsAssociationsOpen={jest.fn()}
        storyId="story-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /blocked by/i }));

    expect(updateAssociation).toHaveBeenCalledWith({
      associationId: "association-1",
      fromStoryId: "story-2",
      storyId: "story-1",
      toStoryId: "story-1",
      type: "blocking",
    });
  });
});
