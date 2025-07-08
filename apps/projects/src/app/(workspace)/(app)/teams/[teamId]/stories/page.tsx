import type { Metadata } from "next";
import { ListStories } from "@/modules/teams/stories/list-stories";

export const metadata: Metadata = {
  title: "Stories",
};

export default function Page() {
  return <ListStories />;
}
