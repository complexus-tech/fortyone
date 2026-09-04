import type { Team } from "@/modules/teams/public/types";
import type { ImportDraft } from "./schema";
import { deriveImportTeamCode, deriveImportTeamColor } from "./execution";
import { normalizeImportMatch } from "./import-entity-matching";

export const toImportTeamName = (value: string) => {
  const name = value.trim().replace(/\s+/g, " ").slice(0, 24);
  if (name.length >= 3) return name;
  return `${name || "New"} team`.slice(0, 24);
};

export const resolveImportSourceTeam = (
  sourceTeam: ImportDraft["teams"][number],
  existingTeams: readonly Team[],
) => {
  const privacyCompatibleTeams = existingTeams.filter(
    (team) => team.isPrivate === sourceTeam.isPrivate,
  );
  const sourceName = normalizeImportMatch(toImportTeamName(sourceTeam.name));
  const nameMatches = privacyCompatibleTeams.filter(
    (team) => normalizeImportMatch(team.name) === sourceName,
  );
  const sourceCode = sourceTeam.code
    ?.trim()
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 3);
  if (!sourceCode) {
    if (nameMatches.length === 1) {
      return { kind: "unique" as const, team: nameMatches[0] };
    }
    return nameMatches.length > 1
      ? { kind: "ambiguous" as const }
      : { kind: "none" as const };
  }
  const codeMatches = privacyCompatibleTeams.filter(
    (team) => team.code.trim().toUpperCase() === sourceCode,
  );
  if (nameMatches.length === 0) {
    return codeMatches.length > 0
      ? { kind: "ambiguous" as const }
      : { kind: "none" as const };
  }
  if (nameMatches.length === 1) {
    if (codeMatches.length === 0) {
      return { kind: "unique" as const, team: nameMatches[0] };
    }
    return codeMatches.length === 1 && codeMatches[0].id === nameMatches[0].id
      ? { kind: "unique" as const, team: nameMatches[0] }
      : { kind: "ambiguous" as const };
  }
  if (
    codeMatches.length === 1 &&
    nameMatches.some((team) => team.id === codeMatches[0].id)
  ) {
    return { kind: "unique" as const, team: codeMatches[0] };
  }
  return { kind: "ambiguous" as const };
};

export const getImportSourceTeamDestination = (
  sourceTeam: ImportDraft["teams"][number],
) => ({
  code: deriveImportTeamCode(sourceTeam),
  color: deriveImportTeamColor(sourceTeam),
  isPrivate: sourceTeam.isPrivate,
  name: toImportTeamName(sourceTeam.name),
});
