import { usePathname } from "next/navigation";
import { cn } from "lib";
import { Box, Collapsible, Divider, Flex, Tooltip } from "ui";
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

export const Navigation = ({
  isCollapsed = false,
}: {
  isCollapsed?: boolean;
}) => {
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
      icon: <ActiveSprintIcon className={isCollapsed ? "h-6" : undefined} />,
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
      icon: <UserIcon className={isCollapsed ? "h-6" : undefined} />,
      href: withWorkspace("/my-work"),
    },
    {
      name: "Calendar",
      icon: <CalendarIcon className={isCollapsed ? "h-6" : undefined} />,
      href: withWorkspace("/calendar"),
    },
    ...(sprintItem ? [sprintItem] : []),
    {
      name: "Summary",
      icon: <DashboardIcon className={isCollapsed ? "h-6" : undefined} />,
      href: withWorkspace("/summary"),
    },
    {
      name: "AI Agent",
      icon: <AiIcon className={isCollapsed ? "h-6" : undefined} />,
      href: withWorkspace("/maya"),
    },
  ];
  const workspaceLinks: MenuItem[] = [
    {
      name: "Roadmap",
      icon: <RoadmapIcon className={isCollapsed ? "h-6" : undefined} />,
      href: withWorkspace("/roadmap"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Strategy Map",
      icon: <StrategyIcon className={isCollapsed ? "h-6" : undefined} />,
      href: withWorkspace("/strategy"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Documents",
      icon: <DocsIcon className={isCollapsed ? "h-6" : undefined} />,
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
          aria-label={name}
          className={cn(
            isActive ? "text-foreground" : undefined,
            isCollapsed && "justify-center px-0",
            isCollapsed && "[&_svg]:!h-6 [&_svg]:!w-auto",
          )}
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
          title={isCollapsed ? name : undefined}
        >
          <Tooltip side="right" title={isCollapsed ? name : null}>
            <span className="shrink-0">{icon}</span>
          </Tooltip>
          <span
            className={cn(
              "line-clamp-1 min-w-0 flex-1 first-letter:capitalize",
              isCollapsed && "hidden",
            )}
          >
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
      <Flex
        className={cn("gap-1.5", isCollapsed && "gap-2")}
        direction="column"
      >
        {renderLinks(primaryLinks)}
      </Flex>
      <Collapsible
        onOpenChange={setIsWorkspaceOpen}
        open={isCollapsed || isWorkspaceOpen}
      >
        <Box className="mt-4">
          {isCollapsed ? (
            <Divider className="my-3" />
          ) : (
            <Collapsible.Trigger asChild>
              <button
                aria-label="Workspace navigation"
                className={cn(
                  "text-text-muted focus-visible:ring-primary/40 hover:text-foreground flex h-8 items-center gap-1 rounded-lg px-2.5 text-left font-medium transition-colors outline-none focus-visible:ring-2",
                  workspaceIsActive && "text-foreground",
                )}
                suppressHydrationWarning
                type="button"
              >
                <span>Workspace</span>
                <ChevronRightIcon
                  className={cn(
                    "h-3.5 w-auto transition-transform duration-200",
                    isWorkspaceOpen && "rotate-90",
                  )}
                  strokeWidth={3.5}
                />
              </button>
            </Collapsible.Trigger>
          )}
          <Collapsible.Content>
            <Flex
              aria-label="Workspace navigation"
              className={cn("mt-1 gap-1.5", isCollapsed && "gap-2")}
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
