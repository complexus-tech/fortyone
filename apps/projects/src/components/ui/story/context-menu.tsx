"use client";
import type { ReactNode } from "react";
import { Fragment } from "react";
import { Box, ContextMenu } from "ui";
import { DeleteIcon, DuplicateIcon, EditIcon, LinkIcon } from "icons";
import type { Story } from "@/modules/stories/types";
import { ContextMenuItem } from "./context-menu-item";

export const StoryContextMenu = ({
  children,
}: {
  children: ReactNode;
  story: Story;
}) => {
  const contextMenu = [
    {
      name: "Main",
      options: [
        {
          label: "Edit",
          icon: <EditIcon className="h-5 w-auto" />,
        },
        {
          label: "Duplicate",
          icon: <DuplicateIcon className="h-5 w-auto" />,
        },
        {
          label: "Open in new tab",
          icon: <LinkIcon className="h-5 w-auto" />,
          onSelect: () => {
            window.open(``, "_blank");
          },
        },
        {
          label: "Copy link",
          icon: <LinkIcon className="h-5 w-auto" />,
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
  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Box>{children}</Box>
      </ContextMenu.Trigger>
      <ContextMenu.Items className="w-56">
        {contextMenu.map(({ name, options }) => (
          <Fragment key={name}>
            <ContextMenu.Group key={name}>
              {options.map(({ label, icon }) => (
                <ContextMenuItem icon={icon} key={label} label={label} />
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
