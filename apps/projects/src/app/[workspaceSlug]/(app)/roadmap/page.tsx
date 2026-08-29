import type { Metadata } from "next";
import { RoadmapPage } from "@/modules/roadmap";

export const metadata: Metadata = {
  title: "Roadmap",
  description:
    "View workspace objectives across timeline, board, and list layouts.",
};

export default function Page() {
  return <RoadmapPage />;
}
