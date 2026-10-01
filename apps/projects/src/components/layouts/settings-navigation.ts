import type { UserRole } from "@/types/user-role";

export type SettingsNavigationItem = {
  title: string;
  href: string;
  activePathPrefixes?: string[];
};

export type SettingsNavigationCategory = {
  category: "Account" | "Workspace" | "Administration" | "Features";
  items: SettingsNavigationItem[];
};

type BuildSettingsNavigationOptions = {
  userRole?: UserRole;
  hasCustomTerminology: boolean;
  hasInvitations: boolean;
  objectiveTitle: string;
  withWorkspace: (path: string) => string;
};

export const buildSettingsNavigation = ({
  userRole,
  hasCustomTerminology,
  hasInvitations,
  objectiveTitle,
  withWorkspace,
}: BuildSettingsNavigationOptions): SettingsNavigationCategory[] => {
  const isAdmin = userRole === "admin";
  const canUseIntegrations =
    isAdmin || userRole === "member" || userRole === "guest";
  const integrationsHref = withWorkspace("/settings/integrations");
  const importsHref = withWorkspace("/settings/workspace/imports");
  const teamsHref = withWorkspace("/settings/workspace/teams");

  const accountItems: SettingsNavigationItem[] = [
    { title: "Profile", href: withWorkspace("/settings/account") },
    { title: "Calendar", href: withWorkspace("/settings/account/calendar") },
    {
      title: "Google Drive",
      href: withWorkspace("/settings/account/google-drive"),
    },
    {
      title: "Preferences",
      href: withWorkspace("/settings/account/preferences"),
    },
    {
      title: "Notifications",
      href: withWorkspace("/settings/account/notifications"),
    },
    ...(hasInvitations
      ? [
          {
            title: "Invitations",
            href: withWorkspace("/settings/invitations"),
          },
        ]
      : []),
  ];

  const workspaceItems: SettingsNavigationItem[] = [
    ...(isAdmin
      ? [
          { title: "General", href: withWorkspace("/settings") },
          ...(hasCustomTerminology
            ? [
                {
                  title: "Terminology",
                  href: withWorkspace("/settings/workspace/terminology"),
                },
              ]
            : []),
        ]
      : []),
    { title: "API", href: withWorkspace("/settings/workspace/api") },
    ...(canUseIntegrations
      ? [
          {
            title: "Integrations",
            href: integrationsHref,
            activePathPrefixes: [
              integrationsHref,
              withWorkspace("/settings/workspace/integrations"),
            ],
          },
        ]
      : []),
  ];

  const navigation: SettingsNavigationCategory[] = [
    { category: "Account", items: accountItems },
    ...(workspaceItems.length > 0
      ? [{ category: "Workspace" as const, items: workspaceItems }]
      : []),
  ];

  if (!isAdmin) return navigation;

  return [
    ...navigation,
    {
      category: "Administration",
      items: [
        // Enable Workspace security in Administration when it is ready.
        // {
        //   title: "Workspace security",
        //   href: withWorkspace("/settings/workspace/security"),
        // },
        {
          title: "Members",
          href: withWorkspace("/settings/workspace/members"),
        },
        {
          title: "Teams",
          href: teamsHref,
          activePathPrefixes: [teamsHref],
        },
        {
          title: "Billing & plans",
          href: withWorkspace("/settings/workspace/billing"),
        },
        {
          title: "Imports & exports",
          href: importsHref,
          activePathPrefixes: [importsHref],
        },
      ],
    },
    {
      category: "Features",
      items: [
        {
          title: "Labels",
          href: withWorkspace("/settings/workspace/labels"),
        },
        {
          title: objectiveTitle,
          href: withWorkspace("/settings/workspace/objectives"),
        },
        {
          title: "Feedback",
          href: withWorkspace("/settings/workspace/feedback"),
        },
      ],
    },
  ];
};

export const isSettingsItemActive = (
  pathname: string,
  item: SettingsNavigationItem,
) =>
  pathname === item.href ||
  item.activePathPrefixes?.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  ) === true;
