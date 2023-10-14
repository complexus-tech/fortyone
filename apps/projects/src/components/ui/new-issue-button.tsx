"use client";
import { useState } from "react";
import { TbPlus } from "react-icons/tb";
import { Button } from "ui";
import { NewIssueDialog } from "./new-issue-dialog";

export const NewIssueButton = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Button
        leftIcon={<TbPlus className="h-5 w-auto" />}
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
