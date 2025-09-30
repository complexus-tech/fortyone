import React from "react";
import { Section } from "@/components/ui";
import { WorkItem } from "./work-item";

export const WorkItemList = () => {
  const handleStoryPress = (storyId: string) => {
    console.log("Story pressed:", storyId);
    // Navigate to story details
  };

  // Mock data for stories
  const stories = [
    {
      id: "1",
      title: "hey",
      status: {
        id: "in-review",
        name: "In Review",
        color: "#3b82f6", // blue
      },
      priority: "High" as const,
      assignee: {
        id: "user-1",
        name: "John Doe",
        avatarUrl: undefined,
      },
    },
    {
      id: "2",
      title: "check again",
      status: {
        id: "in-progress",
        name: "In Progress",
        color: "#eab308", // yellow
      },
      priority: "Medium" as const,
      assignee: {
        id: "user-2",
        name: "Jane Smith",
        avatarUrl: undefined,
      },
    },
    {
      id: "3",
      title: "upgrade nextjs",
      status: {
        id: "todo",
        name: "To Do",
        color: "#6b7280", // gray
      },
      priority: "Low" as const,
    },
  ];

  return (
    <Section title="Work Items">
      {stories.map((story) => (
        <WorkItem key={story.id} story={story} onPress={handleStoryPress} />
      ))}
    </Section>
  );
};
