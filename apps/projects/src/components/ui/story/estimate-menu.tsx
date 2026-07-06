"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, EstimateIcon } from "icons";
import {
  formatEstimate,
  getEstimateOptions,
  type EstimateScheme,
} from "@/lib/estimate";

const EstimateContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useEstimateMenu = () => {
  const { open, setOpen } = useContext(EstimateContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useEstimateMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const EstimateMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <EstimateContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </EstimateContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  align = "center",
  estimateScheme,
  estimateValue,
  setEstimateValue,
}: {
  align?: "start" | "end" | "center";
  estimateScheme: EstimateScheme;
  estimateValue?: number | null;
  setEstimateValue: (estimateValue: number | null) => void;
}) => {
  const { setOpen } = useEstimateMenu();
  const options = getEstimateOptions(estimateScheme);

  return (
    <Popover.Content align={align} className="w-64">
      <Command>
        <Command.Input autoFocus placeholder="Change estimate..." />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No estimate found.</Text>
        </Command.Empty>
        <Command.Group>
          <Command.Item
            active={!estimateValue}
            className="justify-between gap-4 opacity-70"
            onSelect={() => {
              if (estimateValue) {
                setEstimateValue(null);
              }
              setOpen(false);
            }}
            value="No estimate"
          >
            <Box className="grid grid-cols-[24px_auto] items-center">
              <EstimateIcon className="opacity-70" />
              <Text>No estimate</Text>
            </Box>
            <Flex align="center" gap={1}>
              {!estimateValue && (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              )}
              <Text color="muted">0</Text>
            </Flex>
          </Command.Item>
          <Divider className="my-2" />
          {options.map(({ label, value }, idx) => (
            <Command.Item
              active={value === estimateValue}
              className="justify-between gap-4"
              key={value}
              onSelect={() => {
                if (value !== estimateValue) {
                  setEstimateValue(value);
                }
                setOpen(false);
              }}
              value={formatEstimate(estimateScheme, value, "full")}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <EstimateIcon />
                <Text>{formatEstimate(estimateScheme, value, "full")}</Text>
              </Box>
              <Flex align="center" gap={1}>
                {value === estimateValue && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx + 1}</Text>
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

EstimateMenu.Trigger = Trigger;
EstimateMenu.Items = Items;
