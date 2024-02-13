"use client";
import { useState } from "react";
import { Button, Container, Flex, Text, Tooltip } from "ui";
import { Plus } from "lucide-react";
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
    <Container className="select-none bg-gray-50 py-1 dark:bg-dark-200/70">
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          <IssueStatusIcon status={status} />
          <Text fontWeight="medium">{status}</Text>
          <Text color="muted">{count}</Text>
        </Flex>
        <Tooltip side="left" title="Add issue">
          <Button
            className="p-0"
            color="tertiary"
            leftIcon={<Plus className="h-5 w-auto dark:text-gray-200" />}
            onClick={() => {
              setIsOpen(true);
            }}
            variant="naked"
          >
            <span className="sr-only">Add issue</span>
          </Button>
        </Tooltip>
      </Flex>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </Container>
  );
};
