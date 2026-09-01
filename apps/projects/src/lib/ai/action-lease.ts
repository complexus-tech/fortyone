import type { MayaToolName } from "./tool-names";
import type { MayaToolDomain } from "./tool-routing";

export type { MayaToolDomain } from "./tool-routing";

export const MAYA_ACTION_LEASE_VERSION = 1 as const;
export const MAYA_ACTION_LEASE_MAX_TURNS = 20;

/**
 * Server-owned routing continuity for a mutation that is still being prepared.
 * This state only controls which schemas the model can see; tool approval and
 * API authorization remain the security boundaries for every mutation.
 */
export type MayaActionLease = {
  domain: MayaToolDomain;
  operations: string[];
  phase: "collecting";
  remainingTurns: number;
  toolNames: MayaToolName[];
  version: typeof MAYA_ACTION_LEASE_VERSION;
};
