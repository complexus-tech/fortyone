"use client";
import { BodyContainer } from "../../shared/body";
import { MilestoneCard } from "./card";
import { ActiveMilestonesHeader } from "./header";

const milestones = [
  { id: 1, name: "Milestone 1", description: "Planning for the first sprint." },
  {
    id: 2,
    name: "Milestone 2",
    description: "Planning for the second sprint.",
  },
  { id: 3, name: "Milestone 3", description: "" },
];

export const DetailedMilestoneList = () => {
  return (
    <>
      <ActiveMilestonesHeader />
      <BodyContainer>
        {milestones.map(({ id, name, description }) => (
          <MilestoneCard description={description} key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
};
