import type { Metadata } from "next";
import { WorkspaceStrategyMapPage } from "@/modules/strategy";

export const metadata: Metadata = {
  title: "Strategy Map",
  description:
    "Connect the workspace goal to strategic pillars, objectives, and key results.",
};

export default function StrategyPage() {
  return <WorkspaceStrategyMapPage />;
}
