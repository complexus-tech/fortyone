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
  MilestonesIcon,
  StarIcon,
  DocsIcon,
  ChatIcon,
  RetroIcon,
} from "icons";
import { useLocalStorage } from "@/hooks";
import { NavLink } from "../../ui";

type TeamProps = {
  id: number;
  name: string;
  icon?: ReactNode;
};

export const Team = ({ id, name: teamName, icon: teamIcon }: TeamProps) => {
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `teams:${id}:dropdown`,
    false,
  );
  const pathname = usePathname();
  const links = [
    {
      name: "Stories",
      icon: <StoryIcon className="h-5 w-auto" strokeWidth={2} />,
      href: "/teams/web/stories",
    },
    {
      name: "Themes",
      icon: <EpicsIcon className="h-5 w-auto" />,
      href: "/teams/web/themes",
    },
    {
      name: "Milestones",
      icon: <MilestonesIcon className="h-5 w-auto" />,
      href: "/teams/web/milestones",
    },
    {
      name: "Documents",
      icon: <DocsIcon className="h-5 w-auto" />,
      href: "/teams/web/documents",
    },
    {
      name: "Discussions",
      icon: <ChatIcon className="h-[1.35rem] w-auto" strokeWidth={2} />,
      href: "/teams/web/discussions",
    },
    {
      name: "Retrospectives",
      icon: <RetroIcon className="h-5 w-auto" strokeWidth={2} />,
      href: "/teams/web/retrospectives",
    },
  ];

  return (
    <Box>
      <Flex
        align="center"
        className="group h-[2.5rem] select-none rounded-lg pl-2.5 pr-1 outline-none transition hover:bg-gray-250/5 focus:bg-gray-250/5 hover:dark:bg-dark-50/20 focus:dark:bg-dark-50/20"
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
          <span className="text-lg">{teamIcon}</span>
          <span className="block max-w-[15ch] truncate">{teamName}</span>
        </span>
        <Flex align="center" gap={1}>
          <ArrowDownIcon
            className={cn(
              "h-3.5 w-auto -rotate-90 text-gray-300/80 dark:text-gray",
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
                <MoreHorizontalIcon
                  className="relative h-5 w-auto text-gray-300/80 dark:text-gray"
                  strokeWidth={3}
                />
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
          "ml-3 h-0 overflow-hidden border-l border-dotted border-gray-250/15 pl-2 transition-all duration-300 dark:border-dark-50",
          {
            "mt-2 h-max": isOpen,
          },
        )}
        direction="column"
        gap={1}
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
