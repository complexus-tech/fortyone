import { ObjectivesList } from "@/modules/objectives";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Objectives",
};

export default async function Page() {
  const objectives = await getObjectives();

  return <ObjectivesList objectives={objectives} />;
}
