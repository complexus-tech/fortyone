"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { Box, Flex, Text, Button, Command, Popover, Divider } from "ui";
import { ArrowDown2Icon, CheckIcon, InternetIcon } from "icons";
import { SectionHeader } from "@/modules/settings/components";
import { timezones } from "@/lib/timezones";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";

const TimezoneContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useTimezoneMenu = () => {
  const { open, setOpen } = useContext(TimezoneContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useTimezoneMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const TimezoneMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <TimezoneContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </TimezoneContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  placeholder = "Search timezones...",
  align = "end",
  currentTimezone,
  onTimezoneSelected,
}: {
  placeholder?: string;
  align?: "start" | "end" | "center";
  currentTimezone?: string;
  onTimezoneSelected: (timezone: string) => void;
}) => {
  const { setOpen } = useTimezoneMenu();

  return (
    <Popover.Content align={align} className="mr-0 w-80">
      <Command>
        <Command.Input
          autoFocus
          className="text-base"
          placeholder={placeholder}
        />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No timezone found.</Text>
        </Command.Empty>
        <Command.Group className="max-h-80 overflow-y-auto md:max-h-100">
          {timezones.map((timezone) => (
            <Command.Item
              className="flex items-center gap-2 overflow-clip text-base"
              key={timezone.label}
              onSelect={() => {
                onTimezoneSelected(timezone.tzCode);
                setOpen(false);
              }}
            >
              <Flex align="center" className="flex-1" gap={2}>
                <Text as="span" color="muted">
                  (GMT{timezone.utc})
                </Text>
                <Text as="span" className="line-clamp-1">
                  {timezone.tzCode.replace(/_/g, " ")}
                </Text>
              </Flex>
              {timezone.tzCode === currentTimezone && (
                <CheckIcon className="shrink-0" />
              )}
            </Command.Item>
          ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

export const Timezone = () => {
  const { data: profile } = useProfile();
  const { mutate: updateProfile } = useUpdateProfileMutation();

  const handleTimezoneSelected = (timezone: string) => {
    updateProfile({ timezone });
  };

  return (
    <Box className="mt-6 rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Set your local timezone for accurate time displays."
        title="Timezone"
      />
      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex className="items-end gap-3 md:items-center" justify="between">
            <Box>
              <Text className="font-medium">Timezone</Text>
              <Text className="line-clamp-1" color="muted">
                Current timezone setting
              </Text>
            </Box>
            <TimezoneMenu>
              <TimezoneMenu.Trigger>
                <Button
                  className="shrink-0 text-opacity-80"
                  color="tertiary"
                  variant="outline"
                >
                  <Flex align="center" gap={2}>
                    <InternetIcon className="h-4 w-4" />
                    <span className="truncate">{profile?.timezone}</span>
                  </Flex>
                  <ArrowDown2Icon className="h-4 w-4" />
                </Button>
              </TimezoneMenu.Trigger>
              <TimezoneMenu.Items
                currentTimezone={profile?.timezone}
                onTimezoneSelected={handleTimezoneSelected}
              />
            </TimezoneMenu>
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};

TimezoneMenu.Trigger = Trigger;
TimezoneMenu.Items = Items;
