import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { ObjectivesList } from "@/modules/teams/objectives";

export default async function Page() {
  const objectives = await getObjectives();

  return <ObjectivesList objectives={objectives} />;
}
