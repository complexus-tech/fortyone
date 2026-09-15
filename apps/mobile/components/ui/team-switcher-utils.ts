/** Filtering keeps the backend's saved order and never mutates cached teams. */
export function filterSwitcherTeams<T extends { name: string; code: string }>(
  teams: readonly T[],
  query: string,
): readonly T[] {
  const search = query.trim().toLowerCase();
  if (!search) return teams;
  return teams.filter(
    (team) =>
      team.name.toLowerCase().includes(search) ||
      team.code.toLowerCase().includes(search),
  );
}
