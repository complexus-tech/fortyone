"use client";

import { BodyContainer } from "@/components/layout";
import { ProjectsHeader } from "@/components/projects/header";
import { ProjectsList } from "@/components/projects/list";
import type { Project } from "@/components/projects/project";

export default function Page(): JSX.Element {
  const projects: Project[] = [
    {
      id: 1,
      code: "COM-12",
      lead: "John Doe",
      name: "Data migration for Fin connect",
      description: "The quick brown fox jumps over the lazy dog.",
      date: "Sep 27",
    },
    {
      id: 2,
      code: "COM-12",
      lead: "John Doe",
      name: "Complexus data migration",
      description: "Complexus migration to Projects 1.0.0",
      date: "Sep 27",
    },
  ];

  return (
    <>
      <ProjectsHeader />
      <BodyContainer>
        <ProjectsList projects={projects} />
      </BodyContainer>
    </>
  );
}
