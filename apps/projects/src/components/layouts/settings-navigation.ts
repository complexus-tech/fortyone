import type { UserRole } from "@/types";

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

  const accountItems: SettingsNavigationItem[] = [
    { title: "Profile", href: withWorkspace("/settings/account") },
    { title: "Calendar", href: withWorkspace("/settings/account/calendar") },
    {
      title: "Preferences",
      href: withWorkspace("/settings/account/preferences"),
    },
    {
      title: "Notifications",
      href: withWorkspace("/settings/account/notifications"),
    },
    { title: "Security", href: withWorkspace("/settings/account/security") },
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
        {
          title: "Members",
          href: withWorkspace("/settings/workspace/members"),
        },
        {
          title: "Billing & plans",
          href: withWorkspace("/settings/workspace/billing"),
        },
        {
          title: "Imports",
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
          title: "Teams",
          href: withWorkspace("/settings/workspace/teams"),
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
