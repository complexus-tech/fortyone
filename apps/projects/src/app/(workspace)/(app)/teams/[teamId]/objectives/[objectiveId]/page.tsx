import type { Metadata } from "next";
import { ListStories } from "@/modules/objectives/stories/list-stories";
import { auth } from "@/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { toTitleCase } from "@/utils";

export async function generateMetadata({
  params,
  searchParams,
}: {
  params: Promise<{
    teamId: string;
    objectiveId: string;
  }>;
  searchParams: Promise<{
    tab: "stories" | "overview";
  }>;
}): Promise<Metadata> {
  const { objectiveId } = await params;
  const { tab = "overview" } = await searchParams;
  const session = await auth();
  const objectiveData = await getObjective(objectiveId, session!);
  const name = objectiveData?.name || "Objective";

  return {
    title: `${name} › ${toTitleCase(tab)}`,
  };
}

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    objectiveId: string;
  }>;
}) {
  const params = await props.params;

  const { objectiveId } = params;

  return <ListStories objectiveId={objectiveId} />;
}
