import { useMemo } from "react";
import { Box, Kbd, Text } from "ui";
import { usePathname, useRouter } from "next/navigation";
import { useChatContext } from "@/context/chat-context";
import {
  useFeatures,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import {
  getWalkthroughTargetSelector,
  walkthroughTargetSelectors,
  walkthroughTargets,
} from "@/shared/walkthrough/targets";
import { WalkthroughStartChoiceIllustration } from "./walkthrough-illustrations";
import {
  type WalkthroughStep,
  type WalkthroughWelcomeChoice,
} from "./walkthrough-provider";
import { useMayaMessageAvailability } from "@/modules/maya/hooks/use-maya-message-availability";

export const useWalkthroughSteps = (): WalkthroughStep[] => {
  const pathname = usePathname();
  const router = useRouter();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const { openChat } = useChatContext();
  const {
    isLimited: isMayaMessageLimitReached,
    isPending: isMayaMessageAvailabilityPending,
    isUnavailable: isMayaMessageAvailabilityUnavailable,
  } = useMayaMessageAvailability();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const isMayaPage = pathname.includes("/maya");
  const myWorkPath = withWorkspace("/my-work");
  const roadmapPath = withWorkspace("/roadmap");
  const isWalkthroughReady =
    Boolean(userRole) &&
    !features.isPending &&
    !isMayaMessageAvailabilityPending;

  return useMemo(() => {
    if (!isWalkthroughReady) {
      return [];
    }

    const openRoadmap = () => {
      if (!pathname.startsWith(roadmapPath)) {
        router.push(roadmapPath);
      }
    };
    const openMyWork = () => {
      if (!pathname.startsWith(myWorkPath)) {
        router.push(myWorkPath);
      }
    };
    const revealTeams = () => {
      if (typeof document === "undefined") {
        return;
      }

      document
        .querySelector(getWalkthroughTargetSelector(walkthroughTargets.teams))
        ?.scrollIntoView({ block: "nearest" });
    };
    const calendarChoice: WalkthroughWelcomeChoice = {
      description: "Give the work a real place in your week.",
      id: "calendar",
      illustration: (
        <WalkthroughStartChoiceIllustration
          choice="calendar"
          className="h-full w-auto"
        />
      ),
      targetStepId: "calendar",
      title: "Plan my time",
    };
    const mayaChoice: WalkthroughWelcomeChoice = {
      description: "Use Maya to shape a clear first move.",
      id: "maya",
      illustration: (
        <WalkthroughStartChoiceIllustration
          choice="maya"
          className="h-full w-auto"
        />
      ),
      targetStepId: "maya",
      title: "Get help from Maya",
    };
    const memberChoices: WalkthroughWelcomeChoice[] = [
      {
        description: "Turn an idea into a visible next step.",
        id: "create-story",
        illustration: (
          <WalkthroughStartChoiceIllustration
            choice="task"
            className="h-full w-auto"
          />
        ),
        targetStepId: "create-story",
        title: `Create my first ${storyTerm}`,
      },
      ...(features.objectiveEnabled
        ? [
            {
              description: "Connect the work to a bigger outcome.",
              id: "roadmap",
              illustration: (
                <WalkthroughStartChoiceIllustration
                  choice="objective"
                  className="h-full w-auto"
                />
              ),
              targetStepId: "roadmap",
              title: `Create my first ${objectiveTerm}`,
            },
          ]
        : [calendarChoice]),
      features.objectiveEnabled ? calendarChoice : mayaChoice,
    ];
    const guestChoices: WalkthroughWelcomeChoice[] = [
      {
        description: "See the work already shared with you.",
        id: "my-work",
        illustration: (
          <WalkthroughStartChoiceIllustration
            choice="task"
            className="h-full w-auto"
          />
        ),
        targetStepId: "my-work",
        title: `See my ${storyTermPlural}`,
      },
      calendarChoice,
      mayaChoice,
    ];
    const welcomeChoices = userRole === "guest" ? guestChoices : memberChoices;

    const shouldDeferMayaStep =
      isMayaMessageLimitReached || isMayaMessageAvailabilityUnavailable;

    return [
      {
        id: "welcome",
        target: "body",
        title: "What would you like to do first?",
        content: (
          <Text className="text-lg leading-7" color="muted">
            Choose a starting point and we’ll guide you to the next useful
            action. You can always explore more later.
          </Text>
        ),
        panelLayout: "welcome",
        position: "center",
        welcomeChoices,
      },
      ...(userRole !== "guest"
        ? [
            {
              id: "create-story",
              target: walkthroughTargetSelectors.createStory,
              title: `Create your first ${getTermDisplay("storyTerm", {
                capitalize: true,
              })}`,
              action: openMyWork,
              content: (
                <Box className="space-y-3">
                  <Text color="muted">
                    Click Create to add and save a real {storyTerm}. You can add
                    people, dates, and context when they become useful.
                  </Text>
                  <Text color="muted">
                    Start small. One clear {storyTerm} is enough to make the
                    next piece of work visible and easier to move forward.
                  </Text>
                </Box>
              ),
              position: "bottom-start",
              requiredAction: {
                actionLabel: `Create my first ${storyTerm}`,
                id: "story-created",
              },
            },
          ]
        : []),
      ...(userRole === "guest"
        ? [
            {
              id: "my-work",
              target: getWalkthroughTargetSelector(walkthroughTargets.myWork),
              title: `Keep your ${storyTermPlural} in view`,
              content: (
                <Box className="space-y-3">
                  <Text color="muted">
                    My work keeps what you created, own, and collaborate on
                    within easy reach.
                  </Text>
                  <Text color="muted">
                    Switch between list and board views whenever a different
                    lens helps you move forward.
                  </Text>
                </Box>
              ),
              position: "right" as const,
            },
          ]
        : []),
      {
        id: "calendar",
        target: getWalkthroughTargetSelector(walkthroughTargets.calendar),
        title: "Make room for the work",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Calendar brings your planned work and schedule into the same view,
              so priorities have time behind them.
            </Text>
            <Text color="muted">
              Use it when a commitment needs a real place in your week.
            </Text>
          </Box>
        ),
        position: "right",
      },
      ...(userRole !== "guest" && features.objectiveEnabled
        ? [
            {
              id: "roadmap",
              target: getWalkthroughTargetSelector(walkthroughTargets.roadmap),
              title: `Connect work to your ${objectiveTerm}`,
              action: openRoadmap,
              content: (
                <Box className="space-y-3">
                  <Text color="muted">
                    Roadmap keeps individual pieces of work connected to the
                    outcome they should move.
                  </Text>
                  <Text color="muted">
                    Start here when you are ready to turn activity into a
                    clearer direction.
                  </Text>
                </Box>
              ),
              position: "right",
            },
          ]
        : []),
      {
        id: "maya",
        target: shouldDeferMayaStep
          ? "body"
          : getWalkthroughTargetSelector(walkthroughTargets.mayaComposer),
        title: "Try Maya with a real request",
        action: shouldDeferMayaStep || isMayaPage ? undefined : openChat,
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Maya can help you plan, clarify, and prepare the next move. She
              will ask before making a change in your workspace.
            </Text>
            {isMayaMessageLimitReached ? (
              <Text color="muted">
                You&apos;ve used the AI messages included with this month&apos;s
                plan. You can complete setup now, then try Maya when messages
                reset or your plan changes.
              </Text>
            ) : isMayaMessageAvailabilityUnavailable ? (
              <Text color="muted">
                Maya isn&apos;t available to start right now. You can complete
                setup and try her again when the connection is back.
              </Text>
            ) : userRole === "guest" ? (
              <Text color="muted">
                I’ve opened Maya for you. Send a real message to see how she can
                help with your next move.
              </Text>
            ) : (
              <Text color="muted">
                You can always reopen her from the lower-right corner with{" "}
                <span className="inline-flex items-center gap-1 whitespace-nowrap">
                  <Kbd>Shift + M</Kbd>
                </span>
                . Send one real message to try her in context.
              </Text>
            )}
          </Box>
        ),
        highlight: !shouldDeferMayaStep,
        nextActionLabel: shouldDeferMayaStep ? "Continue setup" : undefined,
        position: shouldDeferMayaStep ? "center" : "top-end",
        requiredAction: shouldDeferMayaStep
          ? undefined
          : {
              actionLabel: "Write my first Maya message",
              id: "maya-message-completed",
            },
      },
      {
        id: "help",
        target: getWalkthroughTargetSelector(walkthroughTargets.help),
        title: "Find help when you need it",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Open Help for keyboard shortcuts, product guidance, and a direct
              route to support when you get stuck.
            </Text>
            <Text color="muted">
              <span className="inline-flex items-center gap-1 whitespace-nowrap">
                Open the command menu with <Kbd>⌘ + K</Kbd>
              </span>{" "}
              (or <Kbd>Ctrl + K</Kbd>) to find actions from anywhere.
            </Text>
          </Box>
        ),
        position: "bottom-end",
      },
      {
        id: "teams",
        target: getWalkthroughTargetSelector(walkthroughTargets.teams),
        title: "Keep teams aligned",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Your Teams keeps shared work and the people responsible for it
              close to the work you are doing.
            </Text>
            <Text color="muted">
              Use a team when you need a shared place to coordinate the next
              outcome together.
            </Text>
          </Box>
        ),
        action: revealTeams,
        position: "right",
      },
    ] as WalkthroughStep[];
  }, [
    features.objectiveEnabled,
    getTermDisplay,
    isMayaPage,
    isMayaMessageAvailabilityUnavailable,
    isMayaMessageAvailabilityPending,
    isMayaMessageLimitReached,
    isWalkthroughReady,
    myWorkPath,
    objectiveTerm,
    openChat,
    pathname,
    roadmapPath,
    router,
    storyTerm,
    storyTermPlural,
    userRole,
  ]);
};
