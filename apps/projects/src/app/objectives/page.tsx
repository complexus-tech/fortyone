import { ObjectivesList } from "@/modules/objectives";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";

export default async function Page() {
  const objectives = await getObjectives();

  return <ObjectivesList objectives={objectives} />;
}
