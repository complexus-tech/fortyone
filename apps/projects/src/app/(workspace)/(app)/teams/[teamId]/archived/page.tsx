import type { Metadata } from "next";
import { ListArchivedStories } from "@/modules/teams/archived/list-stories";

export const metadata: Metadata = {
  title: "Archived",
};

export default function Page() {
  return <ListArchivedStories />;
}
