import { usePathname } from "next/navigation";
import { Flex } from "ui";
import {
  AiIcon,
  CalendarIcon,
  DashboardIcon,
  DocsIcon,
  RoadmapIcon,
  StrategyIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import { useWorkspacePath, useFeatures } from "@/hooks";

type MenuItem = {
  name: string;
  icon: ReactNode;
  href: string;
  disabled?: boolean;
};

export const Navigation = () => {
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();

  const features = useFeatures();
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
    {
      name: "AI Assistant",
      icon: <AiIcon />,
      href: withWorkspace("/maya"),
    },
    {
      name: "Summary",
      icon: <DashboardIcon />,
      href: withWorkspace("/summary"),
    },
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

  const renderLinks = (links: MenuItem[]) =>
    links.map(({ name, icon, href, disabled }) => {
      if (disabled) return null;

      const isActive = pathname === href || pathname.startsWith(`${href}/`);

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

  return (
    <Flex className="gap-1.5" direction="column">
      {renderLinks(primaryLinks)}
    </Flex>
  );
};
