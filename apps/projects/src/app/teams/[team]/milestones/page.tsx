import { MilestonesList } from "@/components/teams/milestones/list/list-milestones";
import type { Milestone } from "@/components/teams/milestones/list/row";

export default function Page(): JSX.Element {
  const milestones: Milestone[] = [
    {
      id: 1,
      code: "COM-12",
      lead: "John Doe",
      name: "Milestone 1",
      description: "The quick brown fox jumps over the lazy dog.",
      date: "Sep 27",
    },
    {
      id: 2,
      code: "COM-12",
      lead: "John Doe",
      name: "Milestone 2",
      description: "Complexus migration to Objectives 1.0.0",
      date: "Sep 27",
    },
  ];

  return <MilestonesList milestones={milestones} />;
}
