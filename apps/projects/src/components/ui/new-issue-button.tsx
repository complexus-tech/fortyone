"use client";
import { useState } from "react";
import { Button } from "ui";
import { Plus } from "lucide-react";
import { NewIssueDialog } from "./new-issue-dialog";

export const NewIssueButton = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Button
        leftIcon={<Plus className="h-5 w-auto" />}
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
