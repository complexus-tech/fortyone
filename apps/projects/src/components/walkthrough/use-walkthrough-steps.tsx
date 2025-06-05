import { useMemo } from "react";
import { Box, Kbd, Text } from "ui";
import { type WalkthroughStep } from "./walkthrough-provider";

export const useWalkthroughSteps = (): WalkthroughStep[] => {
  return useMemo(
    () => [
      {
        id: "welcome",
        target: "[data-workspace-switcher]",
        title: "Welcome to Complexus! 👋",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Welcome to your workspace! This is where you and your team
              collaborate on objectives and stories.
            </Text>
            <Text color="muted">
              Click on your workspace name to switch between workspaces, invite
              team members, or access settings.
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
              You can also see notifications, summary, the roadmap, and your
              teams.
            </Text>
          </Box>
        ),
        position: "right",
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
        id: "team-collaboration",
        target: "[data-invite-button]",
        title: "Collaborate with Your Team",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Invite your team members to collaborate on objectives and stories
              together.
            </Text>
            <Text color="muted">
              Use <Kbd className="inline-flex">⌘ I</Kbd> to quickly invite
              someone to your workspace.
            </Text>
          </Box>
        ),
        position: "top-start",
      },
      {
        id: "keyboard-shortcuts",
        target: "[data-help-button]",
        title: "Master Keyboard Shortcuts",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Complexus is built for speed. Learn keyboard shortcuts to boost
              your productivity.
            </Text>
            <Text color="muted">
              Press <Kbd className="inline-flex">⌘ /</Kbd> to see all available
              shortcuts, or check the help menu.
            </Text>
          </Box>
        ),
        position: "top",
      },
      {
        id: "command-menu",
        target: "body",
        title: "Quick Actions",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Press <Kbd className="inline-flex">⌘ + K</Kbd> (or{" "}
              <Kbd className="inline-flex">Ctrl + K</Kbd>) anywhere to open the
              command menu.
            </Text>
            <Text color="muted">
              Quickly search for stories, navigate to different sections, or
              execute actions without using your mouse.
            </Text>
          </Box>
        ),
        position: "center",
        action: () => {
          // Optional: Could trigger command menu demo
        },
      },
      {
        id: "completion",
        target: "body",
        title: "You're All Set! 🎉",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              You&apos;ve completed the walkthrough! You can replay this tour
              anytime from the help menu.
            </Text>
            <Text color="muted">
              Start by creating your first story or setting up your team
              objectives. Happy collaborating!
            </Text>
          </Box>
        ),
        position: "center",
        showSkip: false,
      },
    ],
    [],
  );
};
