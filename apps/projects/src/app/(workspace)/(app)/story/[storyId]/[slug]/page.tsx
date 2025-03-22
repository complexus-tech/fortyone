import type { Metadata } from "next";
import { StoryPage } from "@/modules/story";

export const metadata: Metadata = {
  title: "Story",
};

type Props = {
  params: Promise<{
    storyId: string;
  }>;
};
export default async function Page(props: Props) {
  const params = await props.params;

  const { storyId } = params;

  return <StoryPage storyId={storyId} />;
}
