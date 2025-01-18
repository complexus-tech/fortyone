"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
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

  return (
    <>
      <Button
        color="tertiary"
        leftIcon={leftIcon}
        onClick={() => {
          setIsOpen(true);
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
