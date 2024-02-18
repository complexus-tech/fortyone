"use client";
import { useState } from "react";
import { Plus } from "lucide-react";
import { Button, Container, Flex, Text, Tooltip } from "ui";
import type { IssueStatus } from "../../types/issue";
import { IssueStatusIcon } from "./issue-status-icon";
import { NewIssueDialog } from "./new-issue-dialog";

type IssueHeaderProps = {
  status?: IssueStatus;
  count: number;
};
export const IssuesHeader = ({
  count,
  status = "Backlog",
}: IssueHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Container className="sticky top-0 z-[1] select-none bg-gray-50 py-2 backdrop-blur dark:bg-dark-200/60">
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          <IssueStatusIcon status={status} />
          <Text fontWeight="medium">{status}</Text>
          <Text color="muted">{count}</Text>
        </Flex>
        <Tooltip side="left" title="Add issue">
          <Button
            color="tertiary"
            leftIcon={<Plus className="h-5 w-auto dark:text-gray-200" />}
            onClick={() => {
              setIsOpen(true);
            }}
            size="sm"
            variant="outline"
          >
            <span className="sr-only">Add issue</span>
          </Button>
        </Tooltip>
      </Flex>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} status={status} />
    </Container>
  );
};
