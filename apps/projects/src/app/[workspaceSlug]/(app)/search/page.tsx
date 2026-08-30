import type { Metadata } from "next";
import { WorkspaceSearchPage } from "@/shell/search/workspace-search-page";

export const metadata: Metadata = {
  title: "Search",
};

export default function Page() {
  return <WorkspaceSearchPage />;
}
