import { useMemo } from "react";
import { Text } from "ui";
import { type WalkthroughStep } from "./walkthrough-provider";

export const useWalkthroughSteps = (): WalkthroughStep[] => {
  return useMemo(
    () => [
      {
        id: "welcome",
        target: "[data-workspace-switcher]",
        title: "Welcome to Complexus! 👋",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              Welcome to your workspace! This is where you and your team
              collaborate on objectives and stories.
            </Text>
            <Text color="muted">
              Click on your workspace name to switch between workspaces, invite
              team members, or access settings.
            </Text>
          </div>
        ),
        position: "bottom-start",
      },
      {
        id: "navigation",
        target: "[data-nav-my-work]",
        title: "Navigate Your Work",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              The sidebar is your main navigation hub. Start with &ldquo;My
              Work&rdquo; to see everything assigned to you.
            </Text>
            <Text color="muted">
              You can also browse all Stories, set Objectives, manage Sprints,
              and collaborate with Teams.
            </Text>
          </div>
        ),
        position: "right",
      },
      {
        id: "command-menu",
        target: "body",
        title: "Quick Actions with ⌘K",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              Press <strong>⌘K</strong> (or <strong>Ctrl+K</strong>) anywhere to
              open the command menu.
            </Text>
            <Text color="muted">
              Quickly search for stories, navigate to different sections, or
              execute actions without using your mouse.
            </Text>
          </div>
        ),
        position: "center",
        action: () => {
          // Optional: Could trigger command menu demo
        },
      },
      {
        id: "create-story",
        target: "[data-sidebar-create-story-button]",
        title: "Create Your First Story",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              Stories are the building blocks of your work. Create tasks,
              features, or bugs to track progress.
            </Text>
            <Text color="muted">
              Pro tip: Press <strong>Shift+N</strong> to quickly create a new
              story from anywhere!
            </Text>
          </div>
        ),
        position: "bottom-start",
      },
      {
        id: "team-collaboration",
        target: "[data-invite-button]",
        title: "Collaborate with Your Team",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              Invite your team members to collaborate on objectives and stories
              together.
            </Text>
            <Text color="muted">
              Use <strong>⌘I</strong> to quickly invite someone to your
              workspace.
            </Text>
          </div>
        ),
        position: "top-start",
      },
      {
        id: "keyboard-shortcuts",
        target: "[data-help-button]",
        title: "Master Keyboard Shortcuts",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              Complexus is built for speed. Learn keyboard shortcuts to boost
              your productivity.
            </Text>
            <Text color="muted">
              Press <strong>?</strong> to see all available shortcuts, or check
              the help menu.
            </Text>
          </div>
        ),
        position: "top",
      },
      {
        id: "completion",
        target: "body",
        title: "You&apos;re All Set! 🎉",
        content: (
          <div className="space-y-3">
            <Text color="muted">
              You&apos;ve completed the walkthrough! You can replay this tour
              anytime from the help menu.
            </Text>
            <Text color="muted">
              Start by creating your first story or setting up your team
              objectives. Happy collaborating!
            </Text>
          </div>
        ),
        position: "center",
        showSkip: false,
      },
    ],
    [],
  );
};
