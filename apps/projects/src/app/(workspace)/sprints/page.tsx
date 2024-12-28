import { getRunningSprints } from "@/lib/queries/sprints/get-running-sprints";
import { DetailedSprintList } from "@/modules/sprints";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Sprints",
};

export default async function Page() {
  const runningSprints = await getRunningSprints();
  return <DetailedSprintList sprints={runningSprints!} />;
}
