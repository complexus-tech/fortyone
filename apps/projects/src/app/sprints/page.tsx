"use client";

import { BodyContainer } from "@/components/layout";
import { SprintCard } from "@/components/sprint/card";
import { Header } from "./header";

export default function Page(): JSX.Element {
  const sprints = [
    { id: 1, name: "Sprint 1", description: "Planning for the first sprint." },
    { id: 2, name: "Sprint 2", description: "Planning for the second sprint." },
    { id: 3, name: "Sprint 3", description: "" },
  ];
  return (
    <>
      <Header />
      <BodyContainer>
        {sprints.map(({ id, name, description }) => (
          <SprintCard description={description} key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
}
