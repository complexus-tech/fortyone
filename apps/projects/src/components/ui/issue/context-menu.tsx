import type { ReactNode } from "react";
import { Fragment } from "react";
import { Avatar, Box, ContextMenu } from "ui";
import {
  BellPlus,
  FilePenLine,
  Layers3,
  Pencil,
  Tags,
  Trash2,
  User,
  Clipboard,
  Star,
  Calendar,
  CalendarCheck2,
  GalleryVerticalEnd,
  TimerReset,
} from "lucide-react";
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
        icon: <User className="h-5 w-auto" />,
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
        icon: <Tags className="h-5 w-auto" />,
      },
      {
        label: "Sprint",
        icon: <TimerReset className="h-5 w-auto" />,
      },
      {
        label: "Module",
        icon: <Layers3 className="h-5 w-auto" />,
      },
      {
        label: "Edit",
        icon: <Pencil className="h-5 w-auto" />,
      },
      {
        label: "Rename",
        icon: <FilePenLine className="h-5 w-auto" />,
      },
    ],
  },
  {
    name: "More",
    options: [
      {
        label: "Project",
        icon: <GalleryVerticalEnd className="h-5 w-auto" />,
      },
      {
        label: "Add to sprint",
        icon: <TimerReset className="h-5 w-auto" />,
      },
      {
        label: "Start Date",
        icon: <Calendar className="h-5 w-auto" />,
      },
      {
        label: "Due Date",
        icon: <CalendarCheck2 className="h-5 w-auto" />,
      },
      {
        label: "Clone",
        icon: <Clipboard className="h-5 w-auto" />,
      },
      {
        label: "Favorite",
        icon: <Star className="h-5 w-auto" />,
      },
      {
        label: "Copy",
        icon: <Clipboard className="h-5 w-auto" />,
      },
      {
        label: "Subscribe",
        icon: <BellPlus className="h-5 w-auto" />,
      },
    ],
  },
  {
    name: "Danger Zone",
    options: [
      {
        label: "Delete",
        icon: <Trash2 className="h-5 w-auto text-danger" />,
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
