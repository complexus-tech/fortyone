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
        id: "summary",
        target: "[data-nav-summary]",
        title: "Dashboard Overview",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Get a bird&apos;s-eye view of your workspace activity, team
              progress, and recent updates all in one place.
            </Text>
            <Text color="muted">
              Perfect for understanding what&apos;s happening across your
              workspace and staying aligned with your team.
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
        position: "right",
      },
      {
        id: "notifications",
        target: "[data-nav-notifications]",
        title: "Stay Updated",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Never miss important updates! The notifications section shows you
              mentions, assignments, and team activity.
            </Text>
            <Text color="muted">
              The badge shows unread notifications so you can stay on top of
              what needs your attention.
            </Text>
          </Box>
        ),
        position: "right",
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
              Use this menu to join new teams, leave teams you&apos;re no longer
              part of, or manage team settings.
            </Text>
            <Text color="muted">
              You can discover available teams to join or create new ones if
              you&apos;re an admin. You can also right-click on a team to manage
              it.
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
        id: "completion",
        target: "body",
        title: "You're All Set! 🎉",
        content: (
          <Box className="space-y-3">
            <Text color="muted">
              Press <Kbd className="inline-flex">⌘ + K</Kbd> (or{" "}
              <Kbd className="inline-flex">Ctrl + K</Kbd>) anywhere to open the
              command menu for quick actions.
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
          // Optional: Could trigger command menu demo
        },
      },
    ],
    [],
  );
};
