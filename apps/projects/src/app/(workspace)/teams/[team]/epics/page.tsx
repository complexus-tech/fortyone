import { EpicsPage } from "@/modules/teams/epics";
import { getTeamEpics } from "@/modules/teams/epics/queries/get-epics";

export default async function Page() {
  const epics = await getTeamEpics();

  return <EpicsPage epics={epics} />;
}
