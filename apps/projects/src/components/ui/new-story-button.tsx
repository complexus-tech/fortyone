"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { NewStoryDialog } from "./new-story-dialog";

export const NewStoryButton = ({
  size = "sm",
  children,
  leftIcon = <PlusIcon className="text-white dark:text-gray-200" />,
  teamId,
  sprintId,
  objectiveId,
  ...rest
}: ButtonProps & {
  teamId?: string;
  sprintId?: string;
  objectiveId?: string;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Button
        leftIcon={leftIcon}
        onClick={() => {
          setIsOpen(true);
        }}
        size={size}
        {...rest}
      >
        {children || "New Story"}
      </Button>
      <NewStoryDialog
        isOpen={isOpen}
        objectiveId={objectiveId}
        setIsOpen={setIsOpen}
        sprintId={sprintId}
        teamId={teamId}
      />
    </>
  );
};
