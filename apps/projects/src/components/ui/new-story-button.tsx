"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { cn } from "lib";
import { useUserRole, useTerminology } from "@/hooks";
import { NewStoryDialog } from "./new-story-dialog";

export const NewStoryButton = ({
  size = "sm",
  color = "invert",
  children,
  leftIcon,
  teamId,
  sprintId,
  objectiveId,
  className,
  ...rest
}: ButtonProps & {
  teamId?: string;
  sprintId?: string;
  objectiveId?: string;
  className?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();
  const icon = leftIcon ?? (
    <PlusIcon className="text-current dark:text-current" />
  );

  return (
    <>
      <Button
        className={cn("shrink-0", className)}
        color={color}
        data-new-story-button
        disabled={rest.disabled || userRole === "guest"}
        leftIcon={icon}
        onClick={() => {
          if (userRole !== "guest") {
            setIsOpen(true);
          }
        }}
        size={size}
        {...rest}
      >
        {children || `New ${getTermDisplay("storyTerm")}`}
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
