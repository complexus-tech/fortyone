import type { Metadata } from "next";
import { ListDeletedStories } from "@/modules/teams/deleted/list-stories";

export const metadata: Metadata = {
  title: "Deleted",
};

export default function Page() {
  return <ListDeletedStories />;
}
