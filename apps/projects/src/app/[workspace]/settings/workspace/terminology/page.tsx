import type { Metadata } from "next";
import { WorkspaceTerminologySettings } from "@/modules/settings/workspace/terminology";

export const metadata: Metadata = {
  title: "Settings › Terminology",
};

export default function Page() {
  return <WorkspaceTerminologySettings />;
}
