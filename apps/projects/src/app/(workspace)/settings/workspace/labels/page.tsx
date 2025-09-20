import type { Metadata } from "next";
import { WorkspaceLabelsSettings } from "@/modules/settings/workspace/labels";

export const metadata: Metadata = {
  title: "Settings › Labels",
};

export default function Page() {
  return <WorkspaceLabelsSettings />;
}
