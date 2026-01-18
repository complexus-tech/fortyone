import type { Metadata } from "next";
import { WorkspaceMembersSettings } from "@/modules/settings/workspace/members";

export const metadata: Metadata = {
  title: "Settings › Members",
};

export default function Page() {
  return <WorkspaceMembersSettings />;
}
