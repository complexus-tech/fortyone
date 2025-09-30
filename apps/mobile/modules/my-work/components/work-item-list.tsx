import React from "react";

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
        avatarUrl:
          "https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s288-c-no",
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
        avatarUrl:
          "https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s288-c-no",
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
    {
      id: "4",
      title: "upgrade nextjs 15",
      status: {
        id: "todo",
        name: "To Do",
        color: "#6b7280", // gray
      },
      priority: "Urgent" as const,
    },
    {
      id: "5",
      title: "upgrade nextjs 16",
      status: {
        id: "todo",
        name: "To Do",
        color: "#6b7280", // gray
      },
      priority: "No Priority" as const,
    },
  ];

  return (
    <>
      {stories.map((story) => (
        <WorkItem key={story.id} story={story} onPress={handleStoryPress} />
      ))}
    </>
  );
};
