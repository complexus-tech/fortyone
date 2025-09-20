import type { Metadata } from "next";
import { WorkflowSettings } from "@/modules/settings/workspace/objectives/workflow";

export const metadata: Metadata = {
  title: "Settings › Objectives",
};

export default function ObjectivesWorkflowPage() {
  return <WorkflowSettings />;
}
