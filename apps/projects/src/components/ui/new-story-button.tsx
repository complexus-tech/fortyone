"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { NewStoryDialog } from "./new-story-dialog";

export const NewStoryButton = ({
  size = "sm",
  children,
  leftIcon = <PlusIcon className="h-5 w-auto" />,
  teamId,
  ...rest
}: ButtonProps & {
  teamId?: string;
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
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} teamId={teamId} />
    </>
  );
};
