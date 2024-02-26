import type { Issue, IssueStatus } from "@/types/issue";
import { ListIssues } from "./list-issues";

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
    {
      id: 16,
      priority: "Low",
      status: "Canceled",
      title: "These issues are no longer being worked on.",
    },
    {
      id: 17,
      priority: "High",
      status: "Canceled",
      title: "These issues are no longer being worked on.",
    },
    {
      id: 18,
      priority: "Urgent",
      status: "Canceled",
      title: "These issues are no longer being worked on.",
    },
    {
      id: 19,
      priority: "Medium",
      status: "Canceled",
      title: "These issues are no longer being worked on.",
    },
    {
      id: 20,
      priority: "Low",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 21,
      priority: "High",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 22,
      priority: "Urgent",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 23,
      priority: "Medium",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 24,
      priority: "Low",
      status: "Todo",
      title:
        "These issues are at the top of the backlog and are ready to be worked on.",
    },
    {
      id: 25,
      priority: "High",
      status: "Todo",
      title:
        "These issues are at the top of the backlog and are ready to be worked on.",
    },
    {
      id: 26,
      priority: "Urgent",
      status: "In Progress",
      title: "These issues are being actively worked on.",
    },
    {
      id: 27,
      priority: "Medium",
      status: "In Progress",
      title: "These issues are being actively worked on.",
    },
    {
      id: 28,
      priority: "Low",
      status: "In Progress",
      title: "These issues are being actively worked on.",
    },
    {
      id: 29,
      priority: "High",
      status: "Testing",
      title: "These issues are being tested by the QA team.",
    },
    {
      id: 30,
      priority: "Urgent",
      status: "Testing",
      title: "These issues are being tested by the QA team.",
    },
    {
      id: 31,
      priority: "Medium",
      status: "Testing",
      title: "These issues are being tested by the QA team.",
    },
    {
      id: 32,
      priority: "Low",
      status: "Done",
      title: "These issues are completed and ready to be deployed.",
    },
    {
      id: 33,
      priority: "High",
      status: "Done",
      title: "These issues are completed and ready to be deployed.",
    },
    {
      id: 34,
      priority: "Low",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
    },
    {
      id: 35,
      priority: "Low",
      status: "Backlog",
      title: "These issues are not assigned to any sprint.",
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
  return <ListIssues issues={issues} statuses={statuses} />;
}
