"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { useUserRole } from "@/hooks";
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
  const { userRole } = useUserRole();

  return (
    <>
      <Button
        disabled={rest.disabled || userRole === "guest"}
        leftIcon={leftIcon}
        onClick={() => {
          if (userRole !== "guest") {
            setIsOpen(true);
          }
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
