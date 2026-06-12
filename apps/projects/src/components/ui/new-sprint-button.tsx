"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { useUserRole, useTerminology } from "@/hooks";
import { NewSprintDialog } from "./new-sprint-dialog";

export const NewSprintButton = ({
  teamId,
  size = "sm",
  children,
  leftIcon,
  ...rest
}: ButtonProps & {
  teamId?: string;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const icon = leftIcon ?? <PlusIcon className="h-[1.1rem]" />;

  return (
    <>
      <Button
        color="tertiary"
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
        {children ||
          `New ${getTermDisplay("sprintTerm", { capitalize: true })}`}
      </Button>
      <NewSprintDialog isOpen={isOpen} setIsOpen={setIsOpen} teamId={teamId} />
    </>
  );
};
