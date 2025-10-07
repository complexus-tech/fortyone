import { Tabs, Text } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import React, { useState } from "react";

export const Activity = ({ story }: { story: DetailedStory }) => {
  const [activeTab, setActiveTab] = useState("updates");
  return (
    <Tabs
      defaultValue={activeTab}
      onValueChange={(value) => setActiveTab(value)}
    >
      <Tabs.List className="mb-2">
        <Tabs.Tab value="updates" className="py-2 px-4">
          Updates
        </Tabs.Tab>
        <Tabs.Tab value="comments" className="py-2 px-4">
          Comments
        </Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="comments">
        <Text>Comments</Text>
      </Tabs.Panel>
      <Tabs.Panel value="updates">
        <Text>Updates</Text>
      </Tabs.Panel>
    </Tabs>
  );
};
