import { useMemo } from "react";
import { Box, Kbd, Text } from "ui";
import confetti from "canvas-confetti";
import { useSession } from "next-auth/react";
import { useUserRole } from "@/hooks";
import { type WalkthroughStep } from "./walkthrough-provider";
import { usePathname } from "next/navigation";

export const useWalkthroughSteps = (): WalkthroughStep[] => {
  const { data: session } = useSession();
  const pathname = usePathname();
  const { userRole } = useUserRole();
  return useMemo(
    () =>
      [
        {
          id: "welcome",
          target: "[data-workspace-switcher]",
          title: `Welcome ${session?.user?.name || "to FortyOne"}! 👋`,
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                Welcome to your workspace! This is where you and your team
                collaborate on objectives and stories.
              </Text>
              <Text color="muted">
                Click on your workspace name to switch between workspaces,
                invite team members, or access settings.
              </Text>
            </Box>
          ),
          position: "bottom-start",
        },
        ...(!pathname?.includes("/maya")
          ? [
              {
                id: "ai-assistant-floating",
                target: "[data-chat-button]",
                title: "Meet Maya, Your AI Assistant",
                content: (
                  <Box className="space-y-3">
                    <Text color="muted">
                      Maya is your always-on AI assistant. She can help you
                      manage tasks, answer questions, and navigate your
                      workspace.
                    </Text>
                    <Text color="muted">
                      Click here internally or press <Kbd>Shift + M</Kbd>{" "}
                      anytime to start chatting.
                    </Text>
                  </Box>
                ),
                position: "top-end",
                highlight: true,
              },
            ]
          : []),
        {
          id: "ai-assistant-nav",
          target: "[data-nav-ai-assistant]",
          title: pathname?.includes("/maya")
            ? "Meet Maya, Your AI Assistant"
            : "Dedicated AI Space",
          content: pathname?.includes("/maya") ? (
            <Box className="space-y-3">
              <Text color="muted">
                Maya is your always-on AI assistant. She can help you manage
                tasks, answer questions, and navigate your workspace.
              </Text>
              <Text color="muted">
                This contains your chat history and dedicated workspace for AI
                interactions.
              </Text>
            </Box>
          ) : (
            <Box className="space-y-3">
              <Text color="muted">
                You can also access the full-page AI experience here. Perfect
                for long conversations or complex planning.
              </Text>
            </Box>
          ),
          position: "right",
          highlight: pathname?.includes("/maya"),
        },
        {
          id: "my-notifications",
          target: "[data-sidebar-notifications-button]",
          title: "Stay Updated",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                Never miss important updates! The notifications section shows
                you mentions, assignments, and team activity.
              </Text>
              <Text color="muted">
                The badge shows unread notifications so you can stay on top of
                what needs your attention.
              </Text>
            </Box>
          ),
          position: "bottom-start",
        },
        {
          id: "create-story",
          target: "[data-sidebar-create-story-button]",
          title: "Create Your First Story",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                Stories are the building blocks of your work. They are used to
                track progress and collaborate with your team.
              </Text>
              <Text color="muted">
                Press <Kbd className="inline-flex">Shift + N</Kbd> to quickly
                create a new story from anywhere!
              </Text>
            </Box>
          ),
          position: "bottom-start",
        },
        {
          id: "navigation",
          target: "[data-nav-my-work]",
          title: "Navigate Your Work",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                The sidebar is your main navigation hub. Start with{" "}
                <Kbd className="inline-flex capitalize">My stories</Kbd> to see
                everything you created or assigned to you.
              </Text>
              <Text color="muted">
                This is your personal workspace for tracking your tasks and
                progress.
              </Text>
            </Box>
          ),
          position: "bottom-start",
        },

        {
          id: "teams",
          target: "[data-teams-heading]",
          title: "Your Teams",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                This is where you&apos;ll find all your teams. Each team has its
                own space for collaboration and shared objectives.
              </Text>
              <Text color="muted">
                Click on any team to see team-specific work, members, and
                progress.
              </Text>
            </Box>
          ),
          position: "right",
        },
        {
          id: "manage-teams",
          target: "[data-manage-teams-button]",
          title: "Manage Teams",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                Use this menu to join new teams, leave teams you&apos;re no
                longer part of, or manage team settings.
              </Text>
              <Text color="muted">
                You can discover available teams to join or create new ones if
                you&apos;re an admin. You can also right-click on a team to
                manage it.
              </Text>
            </Box>
          ),
          position: "bottom-start",
        },
        ...(userRole === "admin"
          ? [
              {
                id: "team-collaboration",
                target: "[data-invite-button]",
                title: "Collaborate with Your Team",
                content: (
                  <Box className="space-y-3">
                    <Text color="muted">
                      Invite your team members to collaborate on objectives and
                      stories together.
                    </Text>
                    <Text color="muted">
                      Use <Kbd className="inline-flex">⌘ I</Kbd> to quickly
                      invite someone to your workspace.
                    </Text>
                  </Box>
                ),
                position: "top-start",
              },
            ]
          : []),
        {
          id: "keyboard-shortcuts",
          target: "[data-help-button]",
          title: "Master Keyboard Shortcuts",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                FortyOne is built for speed. Learn keyboard shortcuts to boost
                your productivity.
              </Text>
              <Text color="muted">
                Press <Kbd className="inline-flex">⌘ /</Kbd> to see all
                available shortcuts, or check the help menu.
              </Text>
            </Box>
          ),
          position: "top-start",
        },
        {
          id: "completion",
          target: "body",
          title: "You're All Set! 🎉",
          content: (
            <Box className="space-y-3">
              <Text color="muted">
                Press <Kbd className="inline-flex">⌘ + K</Kbd> (or{" "}
                <Kbd className="inline-flex">Ctrl + K</Kbd>) anywhere to open
                the command menu for quick actions.
              </Text>
              <Text color="muted">
                You&apos;ve completed the walkthrough! Start by creating your
                first story or setting up your team objectives. Happy
                collaborating!
              </Text>
            </Box>
          ),
          position: "center",
          showSkip: false,
          action: () => {
            // Trigger confetti celebration!
            confetti({
              particleCount: 1000,
              spread: 150,
              origin: { y: 0.6 },
              colors: ["#f43f5e", "#06b6d4", "#22c55e", "#eab308", "#a855f7"],
            });
          },
        },
      ] as WalkthroughStep[],
    [session?.user?.name, userRole],
  );
};
