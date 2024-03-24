"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { Box, Flex, Menu } from "ui";
import {
  ArrowDownIcon,
  DeleteIcon,
  StoryIcon,
  LinkIcon,
  EpicsIcon,
  MoreHorizontalIcon,
  SettingsIcon,
  SprintsIcon,
  StarIcon,
  DocsIcon,
} from "icons";
import { useLocalStorage } from "@/hooks";
import { NavLink } from "../../ui";

type ObjectiveProps = {
  id: number;
  name: string;
  icon?: ReactNode;
};

export const Objective = ({
  id,
  name: objectiveName,
  icon: objectiveIcon,
}: ObjectiveProps) => {
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `objectives:${id}:dropdown`,
    false,
  );
  const pathname = usePathname();
  const links = [
    {
      name: "Stories",
      icon: <StoryIcon className="h-[1.35rem] w-auto" strokeWidth={2} />,
      href: "/objectives/web/stories",
    },
    {
      name: "Sprints",
      icon: <SprintsIcon className="h-[1.3rem] w-auto" />,
      href: "/objectives/web/sprints",
    },
    {
      name: "Epics",
      icon: <EpicsIcon className="h-[1.3rem] w-auto" />,
      href: "/objectives/web/epics",
    },
    {
      name: "Documents",
      icon: <DocsIcon className="h-5 w-auto" />,
      href: "/objectives/web/documents",
    },
    {
      name: "Settings",
      icon: <SettingsIcon className="h-5 w-auto" />,
      href: "/my-stories",
    },
  ];

  return (
    <Box className="my-1">
      <Flex
        align="center"
        className="group h-[2.5rem] select-none rounded-lg pl-2.5 pr-1 outline-none transition hover:bg-gray-50/70 focus:bg-gray-50/70 hover:dark:bg-dark-50/20 focus:dark:bg-dark-50/20"
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
          <span className="text-lg">{objectiveIcon}</span>
          <span className="block max-w-[15ch] truncate">{objectiveName}</span>
        </span>
        <Flex align="center" gap={1}>
          <ArrowDownIcon
            className={cn(
              "h-4 w-auto -rotate-90 text-gray-300/80 dark:text-gray",
              {
                "rotate-0": isOpen,
              },
            )}
            strokeWidth={3.5}
          />
          <Menu>
            <Menu.Button>
              <button
                className={cn("px-1 py-2 opacity-0 group-hover:opacity-100", {
                  "opacity-100": isOpen,
                })}
                type="button"
              >
                <MoreHorizontalIcon className="relative top-[1px] h-4 w-auto text-gray-300/80 dark:text-gray" />
                <span className="sr-only">Objective options</span>
              </button>
            </Menu.Button>
            <Menu.Items align="start">
              <Menu.Group>
                <Menu.Item>
                  <StarIcon className="h-[1.15rem] w-auto" />
                  Add to favorites
                </Menu.Item>
                <Menu.Item>
                  <LinkIcon className="h-5 w-auto" />
                  Copy objective link
                </Menu.Item>
                <Menu.Item>
                  <SettingsIcon className="h-5 w-auto" />
                  Settings
                </Menu.Item>
              </Menu.Group>
              <Menu.Separator />
              <Menu.Group>
                <Menu.Item className="text-danger">
                  <DeleteIcon className="h-5 w-auto" />
                  Delete objective
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Flex>
      <Flex
        className={cn(
          "ml-5 h-0 overflow-hidden border-l border-gray-100 pl-2 transition-all duration-300 dark:border-dark-100",
          {
            "mt-2 h-max": isOpen,
          },
        )}
        direction="column"
        gap={2}
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
