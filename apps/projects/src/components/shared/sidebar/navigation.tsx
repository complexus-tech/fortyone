import { usePathname } from "next/navigation";
import { cn } from "lib";
import { Box, Collapsible, Flex } from "ui";
import {
  ActiveSprintIcon,
  AiIcon,
  CalendarIcon,
  ChevronRightIcon,
  DashboardIcon,
  DocsIcon,
  RoadmapIcon,
  StrategyIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import {
  useFeatures,
  useLocalStorage,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { useRunningSprints } from "@/modules/sprints/hooks/running-sprints";

type MenuItem = {
  name: string;
  icon: ReactNode;
  href: string;
  disabled?: boolean;
};

export const Navigation = () => {
  const pathname = usePathname();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const { data: runningSprints = [] } = useRunningSprints();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();

  const features = useFeatures();
  const [isWorkspaceOpen, setIsWorkspaceOpen] = useLocalStorage(
    `sidebar:${workspaceSlug}:workspace-expanded`,
    true,
  );

  const getSprintsItem = (): MenuItem | null => {
    if (runningSprints.length === 0 || userRole === "guest") return null;

    const sprint = runningSprints[0];
    const sprintTerm = getTermDisplay("sprintTerm", {
      variant: runningSprints.length > 1 ? "plural" : "singular",
    });

    return {
      name: `Active ${sprintTerm.charAt(0).toUpperCase()}${sprintTerm.slice(1)}`,
      icon: <ActiveSprintIcon />,
      href:
        runningSprints.length > 1
          ? withWorkspace("/sprints")
          : withWorkspace(
              `/teams/${sprint.teamId}/sprints/${sprint.id}/stories`,
            ),
    };
  };

  const sprintItem = getSprintsItem();
  const primaryLinks: MenuItem[] = [
    {
      name: "My work",
      icon: <UserIcon />,
      href: withWorkspace("/my-work"),
    },
    {
      name: "Calendar",
      icon: <CalendarIcon />,
      href: withWorkspace("/calendar"),
    },
    ...(sprintItem ? [sprintItem] : []),
    {
      name: "Summary",
      icon: <DashboardIcon />,
      href: withWorkspace("/summary"),
    },
    {
      name: "AI Agent",
      icon: <AiIcon />,
      href: withWorkspace("/maya"),
    },
    // {
    //   name: "Summary",
    //   icon: <DashboardIcon />,
    //   href: withWorkspace("/summary"),
    // },
  ];
  const workspaceLinks: MenuItem[] = [
    {
      name: "Roadmap",
      icon: <RoadmapIcon />,
      href: withWorkspace("/roadmap"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Strategy Map",
      icon: <StrategyIcon />,
      href: withWorkspace("/strategy"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Documents",
      icon: <DocsIcon />,
      href: withWorkspace("/docs"),
    },
  ];

  const isLinkActive = (href: string) =>
    pathname === href || pathname.startsWith(`${href}/`);

  const renderLinks = (links: MenuItem[]) =>
    links.map(({ name, icon, href, disabled }) => {
      if (disabled) return null;

      const isActive = isLinkActive(href);

      return (
        <NavLink
          active={isActive}
          aria-current={isActive ? "page" : undefined}
          className={isActive ? "text-foreground" : undefined}
          data-nav-ai-assistant={
            href === withWorkspace("/maya") ? "" : undefined
          }
          data-nav-calendar={
            href === withWorkspace("/calendar") ? "" : undefined
          }
          data-nav-my-work={href === withWorkspace("/my-work") ? "" : undefined}
          data-nav-summary={href === withWorkspace("/summary") ? "" : undefined}
          href={href}
          key={name}
        >
          <span className="shrink-0">{icon}</span>
          <span className="line-clamp-1 min-w-0 flex-1 first-letter:capitalize">
            {name}
          </span>
        </NavLink>
      );
    });

  const workspaceIsActive = workspaceLinks.some(
    ({ disabled, href }) => !disabled && isLinkActive(href),
  );

  return (
    <Box>
      <Flex className="gap-1.5" direction="column">
        {renderLinks(primaryLinks)}
      </Flex>
      <Collapsible onOpenChange={setIsWorkspaceOpen} open={isWorkspaceOpen}>
        <Box className="mt-4">
          <Collapsible.Trigger asChild>
            <button
              className={cn(
                "text-text-muted focus-visible:ring-primary/40 hover:text-foreground flex h-8 items-center gap-1 rounded-lg px-2.5 text-left font-medium transition-colors outline-none focus-visible:ring-2",
                {
                  "text-foreground": workspaceIsActive,
                },
              )}
              suppressHydrationWarning
              type="button"
            >
              <span>Workspace</span>
              <ChevronRightIcon
                className={cn(
                  "h-3.5 w-auto transition-transform duration-200",
                  {
                    "rotate-90": isWorkspaceOpen,
                  },
                )}
                strokeWidth={3.5}
              />
            </button>
          </Collapsible.Trigger>
          <Collapsible.Content>
            <Flex
              aria-label="Workspace navigation"
              className="mt-1 gap-1.5"
              direction="column"
              role="group"
            >
              {renderLinks(workspaceLinks)}
            </Flex>
          </Collapsible.Content>
        </Box>
      </Collapsible>
    </Box>
  );
};
