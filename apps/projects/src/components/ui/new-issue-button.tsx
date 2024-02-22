"use client";
import { useState } from "react";
import { Button } from "ui";
import { PlusIcon } from "icons";
import { NewIssueDialog } from "./new-issue-dialog";

export const NewIssueButton = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Button
        leftIcon={<PlusIcon className="h-5 w-auto" />}
        onClick={() => {
          setIsOpen(true);
        }}
        size="sm"
      >
        New issue
      </Button>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
