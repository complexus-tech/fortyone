"use client";
import { useState } from "react";
import { Flex, Button, Text } from "ui";
import { Plus, Minus } from "lucide-react";
import type { Issue, IssueStatus } from "@/types/issue";
import { IssueStatusIcon, NewIssueDialog } from "@/components/ui";

export const IssuesKanbanHeader = ({
  status,
  issues,
}: {
  status: IssueStatus;
  issues: Issue[];
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const filteredIssues = issues.filter((issue) => issue.status === status);
  return (
    <>
      <Flex
        align="center"
        className="w-[340px] pl-1"
        gap={2}
        justify="between"
        key={status}
      >
        <Flex align="center" gap={2}>
          <IssueStatusIcon status={status} />
          {status}
          <Text as="span" color="muted">
            {filteredIssues.length}
          </Text>
        </Flex>
        <span className="flex items-center gap-1">
          <Button color="tertiary" size="sm" variant="naked">
            <Minus className="h-5 w-auto" />
          </Button>
          <Button
            color="tertiary"
            onClick={() => {
              setIsOpen(true);
            }}
            size="sm"
            variant="naked"
          >
            <Plus className="h-5 w-auto" />
          </Button>
        </span>
      </Flex>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} status={status} />
    </>
  );
};
