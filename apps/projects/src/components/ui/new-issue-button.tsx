"use client";
import { useState } from "react";
import type { ButtonProps } from "ui";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { NewIssueDialog } from "./new-issue-dialog";

export const NewIssueButton = ({
  size = "sm",
  children,
  leftIcon = <PlusIcon className="h-5 w-auto" />,
  ...rest
}: ButtonProps) => {
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
        {children || "New issue"}
      </Button>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
