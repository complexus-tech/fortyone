import type {
  NavigationState,
  StackActionType,
} from "expo-router/react-navigation";

export function createTeamReplacementAction(
  rootState: Pick<NavigationState, "key" | "index" | "routes"> | undefined,
  teamId: string,
) {
  if (!rootState) return null;
  const currentRoute = rootState.routes[rootState.index];
  if (currentRoute?.name !== "teams/[teamId]") return null;

  // Expo's URL replacement can target the nested index for this collapsed
  // dynamic route. Replace its root entry so both layouts receive the new team.
  return {
    type: "REPLACE",
    target: rootState.key,
    source: currentRoute.key,
    payload: {
      name: "teams/[teamId]",
      params: { teamId, screen: "index", params: { teamId } },
    },
  } satisfies StackActionType;
}
