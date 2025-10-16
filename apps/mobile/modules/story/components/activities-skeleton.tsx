import React from "react";
import { Col, Tabs } from "@/components/ui";
import { ActivitySkeleton } from "./activity-skeleton";

export const ActivitiesSkeleton = () => {
  return (
    <Tabs defaultValue="updates">
      <Tabs.List className="mb-2">
        <Tabs.Tab value="updates" className="py-1.5 px-4 rounded-[10px]">
          Updates
        </Tabs.Tab>
        <Tabs.Tab value="comments" className="py-1.5 px-4 rounded-[10px]">
          Comments
        </Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="updates">
        <Col asContainer>
          {Array.from({ length: 7 }).map((_, index) => (
            <ActivitySkeleton key={index} />
          ))}
        </Col>
      </Tabs.Panel>
    </Tabs>
  );
};
