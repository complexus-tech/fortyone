import type { ReactNode } from "react";
import { Fragment } from "react";
import { Avatar, Box, ContextMenu } from "ui";
import {
  BellIcon,
  CalendarPlusIcon,
  CopyIcon,
  DeleteIcon,
  DuplicateIcon,
  EditIcon,
  ModulesIcon,
  ProjectsIcon,
  SprintsIcon,
  StarIcon,
  TagsIcon,
  UserIcon,
} from "icons";
import { IssueStatusIcon } from "../issue-status-icon";
import { PriorityIcon } from "../priority-icon";
import { ContextMenuItem } from "./context-menu-item";

export const contextMenu = [
  {
    name: "Main",
    options: [
      {
        label: "Status",
        icon: (
          <IssueStatusIcon className="text-gray-300/70 dark:text-gray-200" />
        ),
        subMenu: [
          {
            label: "Backlog",
            icon: <IssueStatusIcon status="Backlog" />,
          },
          {
            label: "To Do",
            icon: <IssueStatusIcon status="Todo" />,
          },
          {
            label: "In Progress",
            icon: <IssueStatusIcon status="In Progress" />,
          },
          {
            label: "Testing",
            icon: <IssueStatusIcon status="Testing" />,
          },

          {
            label: "Done",
            icon: <IssueStatusIcon status="Done" />,
          },
          {
            label: "Duplicate",
            icon: <IssueStatusIcon status="Duplicate" />,
          },
          {
            label: "Canceled",
            icon: <IssueStatusIcon status="Canceled" />,
          },
        ],
      },
      {
        label: "Assignee",
        icon: <UserIcon className="h-5 w-auto" />,
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
            className="text-gray-300/70 dark:text-gray-200"
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
        icon: <ModulesIcon className="h-5 w-auto" />,
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
        label: "Project",
        icon: <ProjectsIcon className="h-5 w-auto" />,
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

export const IssueContextMenu = ({ children }: { children: ReactNode }) => {
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
