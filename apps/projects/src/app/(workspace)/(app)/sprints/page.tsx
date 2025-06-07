import type { Metadata } from "next";
import { RunningSprintsList } from "@/modules/sprints/running-sprints";

export const metadata: Metadata = {
  title: "Current Sprints",
};

export default function Page() {
  return <RunningSprintsList />;
}
