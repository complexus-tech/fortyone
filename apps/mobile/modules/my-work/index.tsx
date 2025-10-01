import React from "react";
import { Header } from "./components/header";
import { GroupedStoriesList } from "./components/grouped-list";
import { SafeContainer, Tabs } from "@/components/ui";

export const MyWork = () => {
  // Mock grouped data - replace with actual API data later
  const mockSections = [
    {
      title: "To Do",
      color: "#6b7280",
      data: [
        {
          id: "1",
          title: "Implement user authentication",
          status: { id: "todo", name: "To Do", color: "#6b7280" },
          priority: "High" as const,
          assignee: {
            id: "user-1",
            name: "John Doe",
            avatarUrl:
              "https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s288-c-no",
          },
        },
        {
          id: "2",
          title: "Update documentation",
          status: { id: "todo", name: "To Do", color: "#6b7280" },
          priority: "Medium" as const,
        },
      ],
    },
    {
      title: "In Progress",
      color: "#3b82f6",
      data: [
        {
          id: "3",
          title: "Fix navigation bug",
          status: { id: "in-progress", name: "In Progress", color: "#3b82f6" },
          priority: "Urgent" as const,
          assignee: {
            id: "user-2",
            name: "Jane Smith",
            avatarUrl:
              "https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s288-c-no",
          },
        },
      ],
    },
    {
      title: "In Review",
      color: "#eab308",
      data: [
        {
          id: "4",
          title: "Add dark mode support",
          status: { id: "in-review", name: "In Review", color: "#eab308" },
          priority: "High" as const,
          assignee: {
            id: "user-1",
            name: "John Doe",
            avatarUrl:
              "https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s288-c-no",
          },
        },
      ],
    },
    {
      title: "Done",
      color: "#22c55e",
      data: [
        {
          id: "5",
          title: "Setup CI/CD pipeline",
          status: { id: "done", name: "Done", color: "#22c55e" },
          priority: "Low" as const,
          assignee: {
            id: "user-3",
            name: "Bob Johnson",
          },
        },
      ],
    },
  ];

  const handleStoryPress = (storyId: string) => {
    console.log("Story pressed:", storyId);
    // Navigate to story details
  };

  return (
    <SafeContainer isFull>
      <Header />
      <Tabs defaultValue="all">
        <Tabs.List>
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
          <Tabs.Tab value="created">Created</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <GroupedStoriesList
            sections={mockSections}
            onStoryPress={handleStoryPress}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <GroupedStoriesList
            sections={mockSections.map((section) => ({
              ...section,
              data: section.data.filter((story) => story.assignee),
            }))}
            onStoryPress={handleStoryPress}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <GroupedStoriesList
            sections={mockSections}
            onStoryPress={handleStoryPress}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
