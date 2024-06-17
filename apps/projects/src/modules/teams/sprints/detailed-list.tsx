"use client";
import { BodyContainer } from "../../../components/shared/body";
import { SprintCard } from "./card";
import { ActiveSprintsHeader } from "./header";

const sprints = [
  { id: 1, name: "Sprint 1", description: "Planning for the first sprint." },
  {
    id: 2,
    name: "Sprint 2",
    description: "Planning for the second sprint.",
  },
  { id: 3, name: "Sprint 3", description: "" },
];

export const DetailedSprintList = () => {
  return (
    <>
      <ActiveSprintsHeader />
      <BodyContainer>
        {sprints.map(({ id, name, description }) => (
          <SprintCard description={description} key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
};
