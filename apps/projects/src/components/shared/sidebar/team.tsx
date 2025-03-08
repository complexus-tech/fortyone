"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { Box, Flex } from "ui";
import { ArrowDownIcon, ObjectiveIcon, SprintsIcon, StoryIcon } from "icons";
import { useLocalStorage } from "@/hooks";
import { NavLink, TeamColor } from "../../ui";

type TeamProps = {
  id: string;
  name: string;
  icon?: ReactNode;
  color?: string;
};

export const Team = ({ id, name: teamName, color }: TeamProps) => {
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `teams:${id}:dropdown`,
    false,
  );
  const pathname = usePathname();
  const links = [
    {
      name: "Stories",
      icon: <StoryIcon strokeWidth={2} />,
      href: `/teams/${id}/stories`,
    },
    {
      name: "Sprints",
      icon: <SprintsIcon />,
      href: `/teams/${id}/sprints`,
    },

    {
      name: "Objectives",
      icon: <ObjectiveIcon className="h-[1.1rem]" strokeWidth={2} />,
      href: `/teams/${id}/objectives`,
    },
  ];

  return (
    <Box>
      <Flex
        align="center"
        className="group h-[2.5rem] select-none rounded-lg pl-3 pr-2 outline-none transition hover:bg-gray-250/5 focus:bg-gray-250/5 hover:dark:bg-dark-50/20 focus:dark:bg-dark-50/20"
        justify="between"
        onClick={() => {
          setIsOpen(!isOpen);
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            setIsOpen(!isOpen);
          }
        }}
        role="button"
        tabIndex={0}
      >
        <span className="flex items-center gap-2">
          <TeamColor color={color} />
          <span className="block max-w-[15ch] truncate">{teamName}</span>
        </span>
        <ArrowDownIcon
          className={cn(
            "h-3.5 w-auto -rotate-90 text-gray dark:text-gray-300",
            {
              "rotate-0": isOpen,
            },
          )}
          strokeWidth={3.5}
          suppressHydrationWarning
        />
      </Flex>
      <Flex
        className={cn(
          "ml-5 h-0 overflow-hidden border-l border-gray-250/40 pl-2 transition-all duration-300 dark:border-dark-100",
          {
            "mt-2 h-max": isOpen,
          },
        )}
        direction="column"
        gap={1}
        suppressHydrationWarning
      >
        {links.map(({ name, icon, href }) => {
          const isActive =
            href === "/"
              ? pathname === href || pathname.startsWith("/dashboard")
              : pathname.startsWith(href);
          return (
            <NavLink active={isActive} href={href} key={name}>
              {icon}
              {name}
            </NavLink>
          );
        })}
      </Flex>
    </Box>
  );
};
