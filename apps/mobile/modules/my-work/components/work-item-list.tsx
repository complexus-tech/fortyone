import React from "react";
import { Section } from "@/components/ui";
import { WorkItem } from "./work-item";

export const WorkItemList = () => {
  const handleCyclePress = (cycleTitle: string) => {
    console.log("Cycle pressed:", cycleTitle);
    // Navigate to cycle details
  };

  // Mock data for work items
  const workItems = [
    {
      title: "hey",
      status: "in-review" as const,
      assignee: { name: "John Doe" },
    },
    {
      title: "check again",
      status: "in-progress" as const,
      assignee: { name: "Jane Smith" },
    },
    { title: "upgrade nextjs", status: "todo" as const },
  ];

  return (
    <Section title="Work Items">
      {workItems.map((item, index) => (
        <WorkItem
          key={index}
          title={item.title}
          status={item.status}
          assignee={item.assignee}
          onPress={() => handleCyclePress(item.title)}
        />
      ))}
    </Section>
  );
};
