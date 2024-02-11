import { BreadCrumbs, Button, Flex } from "ui";
import { SlidersHorizontal, ListTodo } from "lucide-react";
import type { Issue, IssueStatus } from "@/types/issue";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";
import { IssuesList } from "./components";

export default function Page() {
  const issues: Issue[] = [
    {
      id: 1,
      priority: "High",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 2,
      priority: "Urgent",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 3,
      priority: "Medium",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 4,
      priority: "Low",
      status: "Todo",
      title:
        "These issues are at the top of the backlog and are ready to be worked on.",
    },
    {
      id: 5,
      priority: "High",
      status: "Todo",
      title:
        "These issues are at the top of the backlog and are ready to be worked on.",
    },
    {
      id: 6,
      priority: "Urgent",
      status: "In Progress",
      title: "These issues are being actively worked on.",
    },
    {
      id: 7,
      priority: "Medium",
      status: "In Progress",
      title: "These issues are being actively worked on.",
    },
    {
      id: 8,
      priority: "Low",
      status: "In Progress",
      title: "These issues are being actively worked on.",
    },
    {
      id: 9,
      priority: "High",
      status: "Testing",
      title: "These issues are being tested by the QA team.",
    },
    {
      id: 10,
      priority: "Urgent",
      status: "Testing",
      title: "These issues are being tested by the QA team.",
    },
    {
      id: 11,
      priority: "Medium",
      status: "Testing",
      title: "These issues are being tested by the QA team.",
    },
    {
      id: 12,
      priority: "Low",
      status: "Done",
      title: "These issues are completed and ready to be deployed.",
    },
    {
      id: 13,
      priority: "High",
      status: "Done",
      title: "These issues are completed and ready to be deployed.",
    },
    {
      id: 14,
      priority: "Urgent",
      status: "Done",
      title: "These issues are completed and ready to be deployed.",
    },
    {
      id: 15,
      priority: "Medium",
      status: "Done",
      title: "These issues are completed and ready to be deployed.",
    },
  ];
  const statuses: IssueStatus[] = [
    "Backlog",
    "Todo",
    "In Progress",
    "Testing",
    "Done",
    "Canceled",
  ];
  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            { icon: <ListTodo className="h-5 w-auto" />, name: "My issues" },
            { name: "Assigned" },
          ]}
        />
        <Flex gap={2}>
          <NewIssueButton />
          <Button
            color="tertiary"
            leftIcon={<SlidersHorizontal className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
        </Flex>
      </HeaderContainer>
      <IssuesList issues={issues} statuses={statuses} />
    </>
  );
}
