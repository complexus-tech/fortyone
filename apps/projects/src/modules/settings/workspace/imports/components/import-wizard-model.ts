import type { ImportDraft } from "../schema";

export type DestinationChoice =
  | { kind: "existing"; teamId: string }
  | {
      kind: "new";
      name: string;
      code: string;
      color: string;
      isPrivate: boolean;
    };

export type ObjectiveTargetPlan = {
  teamConflict: boolean;
  teamId: string | null;
  teamKey: string;
  teamLabel: string;
};

export const WIZARD_STEP = {
  upload: 1,
  teams: 2,
  members: 3,
  review: 4,
  import: 5,
} as const;
export const initialNewTeam = {
  kind: "new" as const,
  name: "",
  code: "",
  color: "#4A90E2",
  isPrivate: false,
};

export const formatTeamCode = (value: string) =>
  value
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 3);

const fileNameToTeamName = (fileName: string) =>
  fileName
    .replace(/\.[^.]+$/, "")
    .replace(/[-_]+/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, 24);

export const getSuggestedTeamName = (
  sourceType: ImportDraft["sourceType"] | undefined,
  fileName: string,
) => {
  if (sourceType === "jira_csv") return "Jira Import";
  return fileNameToTeamName(fileName);
};

export const getNewTeamImportSignature = (
  fileHash: string,
  destination: Extract<DestinationChoice, { kind: "new" }>,
) =>
  JSON.stringify([
    fileHash,
    destination.name.trim(),
    destination.code.trim().toUpperCase(),
    destination.color,
    destination.isPrivate,
  ]);

export const normalizeImportReviewName = (value: string) =>
  value.normalize("NFKC").trim().toLocaleLowerCase().replace(/\s+/g, " ");
