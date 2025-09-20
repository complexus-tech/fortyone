import type { Metadata } from "next";
import { WorkspaceGeneralSettings } from "@/modules/settings/workspace/general";

export const metadata: Metadata = {
  title: "Settings",
};

export default function Page() {
  return <WorkspaceGeneralSettings />;
}
