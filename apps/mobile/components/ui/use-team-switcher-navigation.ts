import type { TeamSwitcherProps } from "./team-switcher.types";
import { useRef } from "react";
import { Keyboard } from "react-native";
import { useNavigation, useRouter } from "expo-router";
import { useAuthStore } from "@/store/auth";
import { createTeamReplacementAction } from "./team-switcher-navigation";

export function useTeamSwitcherNavigation({
  teams,
  currentTeamId,
  setIsOpened,
}: TeamSwitcherProps) {
  const router = useRouter();
  const navigation = useNavigation();
  const rootNavigation = useNavigation("/");
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const pending = useRef<{ teamId: string; sessionEpoch: number } | null>(null);

  const finishSelection = () => {
    const selection = pending.current;
    pending.current = null;
    const session = useAuthStore.getState();
    if (
      !selection ||
      !session.isAuthenticated ||
      session.isLoading ||
      selection.sessionEpoch !== session.sessionEpoch ||
      !navigation.isFocused() ||
      !teams.some((team) => team.id === selection.teamId)
    ) {
      return;
    }
    if (currentTeamId) {
      const action = createTeamReplacementAction(
        rootNavigation.getState(),
        selection.teamId,
      );
      if (action) rootNavigation.dispatch(action);
    } else {
      router.push(`/teams/${selection.teamId}`);
    }
  };

  const close = () => {
    pending.current = null;
    Keyboard.dismiss();
    setIsOpened(false);
  };

  const selectTeam = (teamId: string, waitForDismissal = true) => {
    const session = useAuthStore.getState();
    if (
      pending.current ||
      !session.isAuthenticated ||
      session.isLoading ||
      session.sessionEpoch !== sessionEpoch ||
      !teams.some((team) => team.id === teamId)
    ) {
      return;
    }
    if (teamId === currentTeamId) {
      close();
      return;
    }
    pending.current = { teamId, sessionEpoch };
    Keyboard.dismiss();
    setIsOpened(false);
    // Android's Modal has no onDismiss event. Its controlled visibility and
    // navigation update together; iOS waits for its native sheet to finish.
    if (!waitForDismissal) finishSelection();
  };

  return { close, selectTeam, finishSelection };
}
