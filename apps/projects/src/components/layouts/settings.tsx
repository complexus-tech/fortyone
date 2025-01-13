"use client";
import { ArrowLeftIcon, UserIcon } from "icons";
import type { ReactNode } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Badge, Box, Container, Flex, ResizablePanel, Text, Tooltip } from "ui";
import { useRouter, usePathname } from "next/navigation";
import { useLocalStorage } from "@/hooks";
import { BodyContainer } from "../shared";
import { NavLink } from "../ui";

export const SettingsLayout = ({ children }: { children: ReactNode }) => {
  const [prevPage, setPrevPage] = useLocalStorage("pathBeforeSettings", "");
  const router = useRouter();
  const pathname = usePathname();

  const goBack = () => {
    router.push(prevPage || "/my-work");
    setPrevPage("");
  };

  useHotkeys("esc", () => {
    goBack();
  });

  const navigation = [
    {
      category: "Your account",
      icon: <UserIcon className="h-[1.15rem] w-auto" />,
      items: [
        { title: "Profile", href: "/settings/account" },
        { title: "Preferences", href: "/settings/account/preferences" },
        { title: "API", href: "/settings/account/api" },
        { title: "Security", href: "/settings/account/security" },
        { title: "Delete account", href: "/settings/account/delete" },
      ],
    },
    {
      category: "Workspace settings",
      icon: (
        <svg
          className="h-[1.15rem] w-auto"
          color="currentColor"
          fill="currentColor"
          fillOpacity={0.1}
          height={24}
          viewBox="0 0 24 24"
          width={24}
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M14 22V8C14 5.17157 14 3.75736 13.1213 2.87868C12.2426 2 10.8284 2 8 2C5.17157 2 3.75736 2 2.87868 2.87868C2 3.75736 2 5.17157 2 8V16C2 18.8284 2 20.2426 2.87868 21.1213C3.75736 22 5.17157 22 8 22H14Z"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="2"
          />
          <path
            d="M6.5 11H5.5M10.5 11H9.5M6.5 7H5.5M6.5 15H5.5M10.5 7H9.5M10.5 15H9.5"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="2"
          />
          <path
            d="M18.5 15H17.5M18.5 11H17.5"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="2"
          />
          <path
            d="M18 8H14V22H18C19.8856 22 20.8284 22 21.4142 21.4142C22 20.8284 22 19.8856 22 18V12C22 10.1144 22 9.17157 21.4142 8.58579C20.8284 8 19.8856 8 18 8Z"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="2"
          />
        </svg>
      ),
      items: [
        { title: "General", href: "/settings" },
        { title: "Members", href: "/settings/workspace/members" },
        // { title: "Billing", href: "/settings/workspace/billing" },
        // { title: "Imports / Sync", href: "/settings/workspace/imports" },
        // { title: "Integrations", href: "/settings/workspace/integrations" },
        // { title: "Webhooks", href: "/settings/workspace/webhooks" },
        // { title: "Security", href: "/settings/workspace/security" },
      ],
    },
    // workspace features
    {
      category: "Workspace features",
      icon: (
        <svg
          className="h-[1.15rem] w-auto"
          color="currentColor"
          fill="none"
          height={24}
          viewBox="0 0 24 24"
          width={24}
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M3 4C3 2.34533 3.34533 2 5 2H9C10.6547 2 11 2.34533 11 4C11 5.65467 10.6547 6 9 6H5C3.34533 6 3 5.65467 3 4Z"
            stroke="currentColor"
            strokeWidth="2"
          />
          <path
            d="M13 13C13 11.3453 13.3453 11 15 11H19C20.6547 11 21 11.3453 21 13C21 14.6547 20.6547 15 19 15H15C13.3453 15 13 14.6547 13 13Z"
            stroke="currentColor"
            strokeWidth="2"
          />
          <path
            d="M4 20C4 18.3453 4.34533 18 6 18H10C11.6547 18 12 18.3453 12 20C12 21.6547 11.6547 22 10 22H6C4.34533 22 4 21.6547 4 20Z"
            stroke="currentColor"
            strokeWidth="2"
          />
          <path
            d="M17 11C17 10.5353 17 10.303 16.9616 10.1098C16.8038 9.31644 16.1836 8.69624 15.3902 8.53843C15.197 8.5 14.9647 8.5 14.5 8.5H9.5C9.03534 8.5 8.80302 8.5 8.60982 8.46157C7.81644 8.30376 7.19624 7.68356 7.03843 6.89018C7 6.69698 7 6.46466 7 6"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
          />
          <path
            d="M17 15V16C17 17.8856 17 18.8284 16.4142 19.4142C15.8284 20 14.8856 20 13 20H12"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
          />
        </svg>
      ),
      items: [
        { title: "Automations", href: "/settings/workspace/automations" },
        { title: "Teams", href: "/settings/workspace/teams" },
        // { title: "Sprints", href: "/settings/workspace/sprints" },
        // { title: "Custom fields", href: "/settings/workspace/custom-fields" },
        // { title: "Templates", href: "/settings/workspace/templates" },
        { title: "Labels", href: "/settings/workspace/labels" },
      ],
    },
  ];

  return (
    <ResizablePanel autoSaveId="settings:layout" direction="horizontal">
      <ResizablePanel.Panel
        className="bg-gray-50 dark:bg-[#0f0f0f]"
        defaultSize={18}
        maxSize={20}
        minSize={16}
      >
        <Box className="flex h-16 items-center px-4">
          <Tooltip
            title={
              <span className="flex items-center gap-1">
                Close Settings
                <Badge color="tertiary" rounded="sm" size="sm">
                  Esc
                </Badge>
              </span>
            }
          >
            <button
              className="group flex items-center gap-3 text-lg font-medium"
              onClick={goBack}
              type="button"
            >
              <ArrowLeftIcon className="h-[1.1rem] w-auto opacity-50 transition group-hover:opacity-100" />
              Settings
            </button>
          </Tooltip>
        </Box>
        <BodyContainer className="px-4">
          <Flex className="mt-6" direction="column" gap={4}>
            {navigation.map(({ category, items, icon }) => (
              <Box className="mb-3" key={category}>
                <Flex align="center" className="mb-2" gap={4}>
                  {icon}
                  <Text color="muted">{category}</Text>
                </Flex>
                <Flex className="ml-8" direction="column" gap={1}>
                  {items.map(({ href, title }) => (
                    <NavLink
                      active={pathname === href}
                      className="py-1.5"
                      href={href}
                      key={href}
                    >
                      {title}
                    </NavLink>
                  ))}
                </Flex>
              </Box>
            ))}
          </Flex>
        </BodyContainer>
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={82}>
        <Box className="h-screen overflow-y-auto">
          <Container className="max-w-[54rem] py-12">{children}</Container>
        </Box>
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
};
