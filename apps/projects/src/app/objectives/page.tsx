import { ObjectivesList } from "@/components/objectives/list";
import type { Objective } from "@/components/objectives/objective";

export default function Page(): JSX.Element {
  const objectives: Objective[] = [
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
      description: "Complexus migration to Objectives 1.0.0",
      date: "Sep 27",
    },
  ];

  return <ObjectivesList objectives={objectives} />;
}
