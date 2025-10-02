import React from "react";
import { Pressable } from "react-native";
import { SafeContainer, Tabs, Text, Row, Back } from "@/components/ui";
import { GroupedStoriesList } from "@/modules/my-work/components/grouped-list";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

const StoriesHeader = () => {
  return (
    <Row className="pb-2" asContainer justify="between" align="center">
      <Row align="center" gap={2}>
        <Back />
      </Row>
      <Text fontSize="xl" fontWeight="semibold">
        Product /{" "}
        <Text
          fontSize="xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          Stories
        </Text>
      </Text>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [
          pressed && { backgroundColor: colors.gray[50] },
        ]}
      >
        <SymbolView name="ellipsis" tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};

export default function TeamStories() {
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
      <StoriesHeader />

      <Tabs defaultValue="all">
        <Tabs.List>
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="active">Active</Tabs.Tab>
          <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="all">
          <GroupedStoriesList
            sections={mockSections}
            onStoryPress={handleStoryPress}
          />
        </Tabs.Panel>
        <Tabs.Panel value="active">
          <GroupedStoriesList
            sections={mockSections}
            onStoryPress={handleStoryPress}
          />
        </Tabs.Panel>
        <Tabs.Panel value="backlog">
          <GroupedStoriesList
            sections={mockSections}
            onStoryPress={handleStoryPress}
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
}
