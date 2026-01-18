import type { Metadata } from "next";
import { CreateTeam } from "@/modules/settings/workspace/teams/create";

export const metadata: Metadata = {
  title: "Settings › Create Team",
};

export default function CreateTeamPage() {
  return <CreateTeam />;
}
