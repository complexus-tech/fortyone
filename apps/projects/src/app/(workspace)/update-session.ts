"use server";

import { updateSession } from "@/auth";
import type { Workspace } from "@/types";

export async function updateSessionAction(workspace: Workspace) {
  await updateSession({ activeWorkspace: workspace });
}
