import type { ReactNode } from "react";
import { Fragment } from "react";
import { FiEdit } from "react-icons/fi";
import { HiViewGrid } from "react-icons/hi";
import {
  TbBellPlus,
  TbCalendarPlus,
  TbCopy,
  TbFolder,
  TbLink,
  TbMap,
  TbPencil,
  TbStar,
  TbTag,
  TbTrash,
  TbUserShare,
} from "react-icons/tb";
import { Avatar, Box, ContextMenu } from "ui";
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
        icon: <TbUserShare className="h-5 w-auto" />,
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
        icon: <TbTag className="h-5 w-auto" />,
      },
      {
        label: "Sprint",
        icon: <TbMap className="h-5 w-auto" />,
      },
      {
        label: "Module",
        icon: <TbFolder className="h-5 w-auto" />,
      },
      {
        label: "Edit",
        icon: <TbPencil className="h-5 w-auto" />,
      },
      {
        label: "Rename",
        icon: <FiEdit className="h-5 w-auto" />,
      },
    ],
  },
  {
    name: "More",
    options: [
      {
        label: "Project",
        icon: <HiViewGrid className="h-5 w-auto" />,
      },
      {
        label: "Add to sprint",
        icon: <TbTag className="h-5 w-auto" />,
      },
      {
        label: "Due Date",
        icon: <TbCalendarPlus className="h-5 w-auto" />,
      },
      {
        label: "Start Date",
        icon: <TbCalendarPlus className="h-5 w-auto" />,
      },
      {
        label: "Clone",
        icon: <TbCopy className="h-5 w-auto" />,
      },
      {
        label: "Favorite",
        icon: <TbStar className="h-5 w-auto" />,
      },
      {
        label: "Copy",
        icon: <TbLink className="h-5 w-auto" />,
      },
      {
        label: "Subscribe",
        icon: <TbBellPlus className="h-5 w-auto" />,
      },
    ],
  },
  {
    name: "Danger Zone",
    options: [
      {
        label: "Delete",
        icon: <TbTrash className="h-5 w-auto text-danger" />,
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
