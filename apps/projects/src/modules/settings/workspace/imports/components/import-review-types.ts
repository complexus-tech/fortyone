export type ObjectiveImportDestinationPreview = {
  kind:
    | "none"
    | "unique"
    | "ambiguous"
    | "pillar_conflict"
    | "privacy_conflict"
    | "source_conflict"
    | "team_conflict";
  teamLabel: string;
  objectiveName?: string;
  matchCount?: number;
  locked?: boolean;
  pillarLabel?: string;
};

export type StrategicPillarImportDestinationPreview = {
  kind: "none" | "unique" | "ambiguous" | "source_conflict";
  pillarName?: string;
  matchCount?: number;
};
