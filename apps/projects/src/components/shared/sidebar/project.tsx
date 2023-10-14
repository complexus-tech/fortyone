"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { useState } from "react";
import { PiCaretDownBold } from "react-icons/pi";
import {
  TbBoxMultiple,
  TbClipboardText,
  TbDots,
  TbFileText,
  TbFolder,
  TbLink,
  TbMap,
  TbSettings,
  TbStar,
  TbTrash,
} from "react-icons/tb";
import { Box, Button, Flex, Menu } from "ui";
import { NavLink } from "../../ui";

type ProjectProps = {
  name: string;
  icon?: ReactNode;
};

export const Project = ({
  name: projectName,
  icon: projectIcon,
}: ProjectProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();
  const links = [
    {
      name: "Issues",
      icon: <TbClipboardText className="h-5 w-auto" />,
      href: "/projects",
    },
    {
      name: "Sprints",
      icon: <TbMap className="h-5 w-auto" />,
      href: "/analytics",
    },
    {
      name: "Modules",
      icon: <TbFolder className="h-5 w-auto" />,
      href: "/inbox",
    },
    {
      name: "Views",
      icon: <TbBoxMultiple className="h-5 w-auto" />,
      href: "/views",
    },
    {
      name: "Pages",
      icon: <TbFileText className="h-5 w-auto" />,
      href: "/my-issues",
    },
    {
      name: "Settings",
      icon: <TbSettings className="h-5 w-auto" />,
      href: "/my-issues",
    },
  ];
  return (
    <Box>
      <Button
        className="mt-2 justify-between"
        color="tertiary"
        fullWidth
        onClick={() => {
          setIsOpen(!isOpen);
        }}
        rightIcon={
          <Flex align="center" gap={1}>
            <PiCaretDownBold
              className={cn("-rotate-90", {
                "rotate-0": isOpen,
              })}
            />
            <Menu>
              <Menu.Button>
                <button className="py-2" type="button">
                  <TbDots className="relative top-[1px] h-5 w-auto" />
                  <span className="sr-only">Project options</span>
                </button>
              </Menu.Button>
              <Menu.Items align="start">
                <Menu.Group>
                  <Menu.Item>
                    <TbStar className="h-[1.15rem] w-auto" />
                    Add to favorites
                  </Menu.Item>
                  <Menu.Item>
                    <TbLink className="h-5 w-auto" />
                    Copy project link
                  </Menu.Item>
                  <Menu.Item>
                    <TbSettings className="h-5 w-auto" />
                    Settings
                  </Menu.Item>
                </Menu.Group>
                <Menu.Separator />
                <Menu.Group>
                  <Menu.Item className="text-danger">
                    <TbTrash className="h-5 w-auto" />
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
          <span className="inline-block max-w-[110px] truncate">
            {projectName}
          </span>
        </span>
      </Button>
      <Flex
        className={cn(
          "ml-6 h-0 overflow-hidden border-l border-gray-100 pl-2 transition-all duration-300 dark:border-dark-100",
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
