import type { Metadata } from "next";
import { CreateWorkspace } from "@/modules/onboarding/create";

export const metadata: Metadata = {
  title: "Create Workspace - Complexus",
  description: "Create a new workspace",
};

export default function CreateWorkspacePage() {
  return <CreateWorkspace />;
}
