"use client";
import { BodyContainer } from "../../components/shared/body";
import { SprintCard } from "./components/card";
import { ActiveSprintsHeader } from "./components/header";
import type { Sprint } from "./types";

export const DetailedSprintList = ({ sprints }: { sprints: Sprint[] }) => {
  return (
    <>
      <ActiveSprintsHeader />
      <BodyContainer>
        {sprints.map((sprint) => (
          <SprintCard key={sprint.id} {...sprint} />
        ))}
      </BodyContainer>
    </>
  );
};
