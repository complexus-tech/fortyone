"use client";

import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import {
  ONBOARDING_CALENDAR_PATH,
  ONBOARDING_START_QUERY,
  ONBOARDING_TASK_START,
} from "@/modules/onboarding/public/start";
import { useJoinedTeams } from "@/modules/teams/public/queries";

export const useStoryCreationDialog = () => {
  const [isOpen, setIsOpen] = useState(false);
  const hasCreatedStoryRef = useRef(false);
  const isOpenRef = useRef(false);
  const onboardingStoryRef = useRef(false);
  const consumedOnboardingWorkspaceRef = useRef<string | null>(null);
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const onboardingStart = searchParams.get(ONBOARDING_START_QUERY);
  const { data: session } = useSession();
  const {
    data: members,
    isPending: membersPending,
    isError: membersError,
  } = useMembers();
  const {
    data: teams,
    isPending: teamsPending,
    isError: teamsError,
  } = useJoinedTeams();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const {
    closeWalkthrough,
    completeWalkthroughAction,
    state: walkthroughState,
  } = useWalkthrough();
  const calendarPath = withWorkspace(ONBOARDING_CALENDAR_PATH);
  const myWorkPath = withWorkspace("/my-work");

  const completeCreatedStoryWalkthroughAction = useCallback(() => {
    if (isOpenRef.current || !hasCreatedStoryRef.current) {
      return;
    }

    hasCreatedStoryRef.current = false;
    completeWalkthroughAction("story-created");
    if (onboardingStoryRef.current) {
      onboardingStoryRef.current = false;
      toast("Plan around your meetings", {
        description:
          "Optional: connect your calendar to see meetings alongside your work.",
        action: {
          label: "Connect calendar",
          onClick: () => {
            router.push(calendarPath);
          },
        },
        duration: 10_000,
      });
    }
  }, [calendarPath, completeWalkthroughAction, router]);
  const setDialogOpen: Dispatch<SetStateAction<boolean>> = useCallback(
    (nextState) => {
      const nextOpen =
        typeof nextState === "function"
          ? nextState(isOpenRef.current)
          : nextState;

      isOpenRef.current = nextOpen;
      setIsOpen(nextOpen);

      if (!nextOpen) {
        completeCreatedStoryWalkthroughAction();
      }
    },
    [completeCreatedStoryWalkthroughAction],
  );
  const onCreated = () => {
    hasCreatedStoryRef.current = true;

    queueMicrotask(completeCreatedStoryWalkthroughAction);
  };

  const currentMemberReady = Boolean(
    !membersPending &&
      !membersError &&
      members.some(
        (member) =>
          member.id === session?.user.id && member.isActive && !member.isSystem,
      ),
  );
  const joinedTeamsReady =
    !teamsPending && !teamsError && Boolean(teams.length);
  useEffect(() => {
    if (
      onboardingStart !== ONBOARDING_TASK_START ||
      !workspaceSlug ||
      !session?.user.id ||
      !userRole ||
      pathname !== myWorkPath ||
      consumedOnboardingWorkspaceRef.current === workspaceSlug
    )
      return;

    const canCreate = userRole === "admin" || userRole === "member";
    if (canCreate && (!currentMemberReady || !joinedTeamsReady)) return;

    consumedOnboardingWorkspaceRef.current = workspaceSlug;
    const url = new URL(window.location.href);
    url.searchParams.delete(ONBOARDING_START_QUERY, ONBOARDING_TASK_START);
    window.history.replaceState(
      window.history.state,
      "",
      `${url.pathname}${url.search}${url.hash}`,
    );
    if (!canCreate) return;

    onboardingStoryRef.current = true;
    closeWalkthrough();
    setDialogOpen(true);
  }, [
    closeWalkthrough,
    currentMemberReady,
    joinedTeamsReady,
    myWorkPath,
    onboardingStart,
    pathname,
    session?.user.id,
    setDialogOpen,
    userRole,
    workspaceSlug,
  ]);

  // A tour whose asynchronous progress load finishes after this handoff must
  // yield to the task dialog without completing or skipping the tour.
  useEffect(() => {
    if (isOpen && onboardingStoryRef.current && walkthroughState.isActive) {
      closeWalkthrough();
    }
  }, [closeWalkthrough, isOpen, walkthroughState.isActive]);

  return { isOpen, setIsOpen: setDialogOpen, onCreated };
};
