import type { Metadata } from "next";
import { RoadmapPage } from "@/modules/roadmap";

export const metadata: Metadata = {
  title: "Roadmap",
  description:
    "Strategic roadmap and objectives timeline view for project planning and goal tracking.",
};

export default function Page() {
  return <RoadmapPage />;
}
