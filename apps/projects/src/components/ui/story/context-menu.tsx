"use client";
import type { ReactNode } from "react";
import { Fragment } from "react";
import { Avatar, Box, ContextMenu } from "ui";
import {
  AssigneeIcon,
  BellIcon,
  CalendarPlusIcon,
  CopyIcon,
  DeleteIcon,
  DuplicateIcon,
  EditIcon,
  EpicsIcon,
  ObjectiveIcon,
  SprintsIcon,
  StarIcon,
  TagsIcon,
} from "icons";
import type { Story } from "@/modules/stories/types";
import { StoryStatusIcon } from "../story-status-icon";
import { PriorityIcon } from "../priority-icon";
import { ContextMenuItem } from "./context-menu-item";

export const contextMenu = [
  {
    name: "Main",
    options: [
      {
        label: "Status",
        icon: <StoryStatusIcon className="text-gray dark:text-gray-200" />,
        subMenu: [
          {
            label: "Backlog",
            icon: <StoryStatusIcon />,
          },
          {
            label: "To Do",
            icon: <StoryStatusIcon />,
          },
          {
            label: "In Progress",
            icon: <StoryStatusIcon />,
          },
          {
            label: "Testing",
            icon: <StoryStatusIcon />,
          },

          {
            label: "Done",
            icon: <StoryStatusIcon />,
          },
          {
            label: "Duplicate",
            icon: <StoryStatusIcon />,
          },
          {
            label: "Canceled",
            icon: <StoryStatusIcon />,
          },
        ],
      },
      {
        label: "Assignee",
        icon: <AssigneeIcon className="h-5 w-auto" />,
        shortCut: "⌘+[",
        subMenu: [
          {
            label: "Joseph Mukorivo",
            icon: (
              <Avatar
                name="Joseph Mukorivo"
                size="sm"
                src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
              />
            ),
          },
          {
            label: "John Doe",
            icon: (
              <Avatar
                name="John Doe"
                size="sm"
                src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
              />
            ),
          },
          {
            label: "Abraham Lincoln",
            icon: <Avatar name="Abraham Lincoln" size="sm" />,
          },
        ],
      },
      {
        label: "Priority",
        icon: (
          <PriorityIcon
            className="text-gray dark:text-gray-200"
            priority="High"
          />
        ),
      },
      {
        label: "Labels",
        icon: <TagsIcon className="h-5 w-auto" />,
      },
      {
        label: "Sprint",
        icon: <SprintsIcon className="h-5 w-auto" />,
      },
      {
        label: "Module",
        icon: <EpicsIcon className="h-5 w-auto" />,
      },
      {
        label: "Edit",
        icon: <EditIcon className="h-5 w-auto" />,
      },
    ],
  },
  {
    name: "More",
    options: [
      {
        label: "Objective",
        icon: <ObjectiveIcon className="h-5 w-auto" />,
      },
      {
        label: "Add to sprint",
        icon: <SprintsIcon className="h-5 w-auto" />,
      },
      {
        label: "Start Date",
        icon: <CalendarPlusIcon className="h-5 w-auto" />,
      },
      {
        label: "Due Date",
        icon: <CalendarPlusIcon className="h-5 w-auto" />,
      },
      {
        label: "Duplicate",
        icon: <DuplicateIcon className="h-5 w-auto" />,
      },
      {
        label: "Favorite",
        icon: <StarIcon className="h-5 w-auto" />,
      },
      {
        label: "Copy",
        icon: <CopyIcon className="h-5 w-auto" />,
      },
      {
        label: "Subscribe",
        icon: <BellIcon className="h-5 w-auto" />,
      },
    ],
  },
  {
    name: "Danger Zone",
    options: [
      {
        label: "Delete",
        icon: <DeleteIcon className="h-5 w-auto text-danger" />,
      },
    ],
  },
];

export const StoryContextMenu = ({
  children,
}: {
  children: ReactNode;
  story: Story;
}) => {
  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Box>{children}</Box>
      </ContextMenu.Trigger>
      <ContextMenu.Items className="w-72">
        {contextMenu.map(({ name, options }) => (
          <Fragment key={name}>
            <ContextMenu.Group key={name}>
              {options.map(({ label, icon, subMenu, shortCut }) => (
                <ContextMenuItem
                  icon={icon}
                  key={label}
                  label={label}
                  shortCut={shortCut}
                  subMenu={subMenu}
                />
              ))}
            </ContextMenu.Group>
            {name !== "Danger Zone" && (
              <ContextMenu.Separator className="my-2" />
            )}
          </Fragment>
        ))}
      </ContextMenu.Items>
    </ContextMenu>
  );
};
