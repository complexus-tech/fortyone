import type { Metadata } from "next";
import { WorkspaceKeyResultsPage } from "@/modules/key-results";

export const metadata: Metadata = {
  title: "Key Results",
  description:
    "Track measurable key results across every objective and team in the workspace.",
};

export default function Page() {
  return <WorkspaceKeyResultsPage />;
}
