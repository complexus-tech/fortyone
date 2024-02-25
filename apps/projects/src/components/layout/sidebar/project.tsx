"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { Box, Button, Flex, Menu } from "ui";
import {
  ArrowDownIcon,
  DeleteIcon,
  IssuesIcon,
  LinkIcon,
  ModulesIcon,
  MoreHorizontalIcon,
  SettingsIcon,
  SprintsIcon,
  StarIcon,
  DocsIcon,
} from "icons";
import { useLocalStorage } from "@/hooks";
import { NavLink } from "../../ui";

type ProjectProps = {
  name: string;
  icon?: ReactNode;
};

export const Project = ({
  name: projectName,
  icon: projectIcon,
}: ProjectProps) => {
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `project-${projectName}-dropdown`,
    true,
  );
  const pathname = usePathname();
  const links = [
    {
      name: "Issues",
      icon: <IssuesIcon className="h-[1.35rem] w-auto" />,
      href: "/projects/web/issues",
    },
    {
      name: "Sprints",
      icon: <SprintsIcon className="h-[1.3rem] w-auto" />,
      href: "/projects/web/sprints",
    },
    {
      name: "Modules",
      icon: <ModulesIcon className="h-[1.3rem] w-auto" />,
      href: "/projects/web/modules",
    },
    {
      name: "Docs",
      icon: <DocsIcon className="h-5 w-auto" />,
      href: "/projects/web/docs",
    },
    {
      name: "Settings",
      icon: <SettingsIcon className="h-5 w-auto" />,
      href: "/my-issues",
    },
  ];
  return (
    <Box>
      <Button
        className="group mt-2 justify-between"
        color="tertiary"
        fullWidth
        onClick={() => {
          setIsOpen(!isOpen);
        }}
        rightIcon={
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
                  <span className="sr-only">Project options</span>
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
                    Copy project link
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
                    Delete project
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        }
        variant="naked"
      >
        <span className="flex items-center gap-2">
          <span className="text-lg">{projectIcon}</span>
          <span className="block max-w-[15ch] truncate">{projectName}</span>
        </span>
      </Button>
      <Flex
        className={cn(
          "ml-5 h-0 overflow-hidden border-l border-gray-100 pl-2 transition-all duration-300 dark:border-dark-200",
          {
            "mt-2 h-max": isOpen,
          },
        )}
        direction="column"
        gap={2}
      >
        {links.map(({ name, icon, href }) => (
          <NavLink active={pathname === href} href={href} key={name}>
            {icon}
            {name}
          </NavLink>
        ))}
      </Flex>
    </Box>
  );
};
