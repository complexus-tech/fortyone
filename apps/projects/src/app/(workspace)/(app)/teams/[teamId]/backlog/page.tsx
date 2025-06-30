import type { Metadata } from "next";
import { ListBacklogStories } from "@/modules/teams/backlog/list-stories";

export const metadata: Metadata = {
  title: "Backlog",
};

export default function Page() {
  return <ListBacklogStories />;
}
