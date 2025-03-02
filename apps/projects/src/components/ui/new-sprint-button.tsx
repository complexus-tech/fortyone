"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { useUserRole } from "@/hooks";
import { NewSprintDialog } from "./new-sprint-dialog";

export const NewSprintButton = ({
  teamId,
  size = "sm",
  children,
  leftIcon = <PlusIcon className="h-[1.1rem]" />,
  ...rest
}: ButtonProps & {
  teamId?: string;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();
  return (
    <>
      <Button
        color="tertiary"
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
        {children || "New Sprint"}
      </Button>
      <NewSprintDialog isOpen={isOpen} setIsOpen={setIsOpen} teamId={teamId} />
    </>
  );
};
