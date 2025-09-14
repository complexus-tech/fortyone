import type { Metadata } from "next";
import { TeamsList } from "@/modules/settings/workspace/teams/list";

export const metadata: Metadata = {
  title: "Settings › Teams",
};

export default function TeamsPage() {
  return <TeamsList />;
}
