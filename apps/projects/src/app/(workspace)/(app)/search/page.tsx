import type { Metadata } from "next";
import { SearchPage } from "@/modules/search";

export const metadata: Metadata = {
  title: "Search",
};

export default function Page() {
  return <SearchPage />;
}
