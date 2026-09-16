import React from "react";
import { Col, Tabs } from "@/components/ui";
import { ActivitySkeleton } from "./activity-skeleton";

export const ActivitiesSkeleton = () => {
  return (
    <Tabs defaultValue="updates">
      <Tabs.List
        className="mb-[6px]"
        accessibilityLabel="Story activity"
        labelSize={14}
        options={[
          { value: "updates", label: "Updates" },
          { value: "comments", label: "Comments" },
        ]}
      />
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
