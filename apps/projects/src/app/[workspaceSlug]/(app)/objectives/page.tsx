import type { Metadata } from "next";
import { WorkspaceObjectivesPage } from "@/modules/roadmap";

export const metadata: Metadata = {
  title: "Objectives",
  description:
    "Manage workspace objectives in a list or strategic timeline view.",
};

export default function Page() {
  return <WorkspaceObjectivesPage />;
}
