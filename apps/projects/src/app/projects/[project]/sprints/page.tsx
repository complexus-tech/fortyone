import { SprintsList } from "@/components/sprints/list/list-sprints";
import type { Sprint } from "@/components/sprints/list/row";

export default function Page(): JSX.Element {
  const sprints: Sprint[] = [
    {
      id: 1,
      code: "COM-12",
      lead: "John Doe",
      name: "Sprint 1",
      description: "The quick brown fox jumps over the lazy dog.",
      date: "Sep 27",
    },
    {
      id: 2,
      code: "COM-12",
      lead: "John Doe",
      name: "Sprint 2",
      description: "Complexus migration to Projects 1.0.0",
      date: "Sep 27",
    },
  ];

  return <SprintsList sprints={sprints} />;
}
