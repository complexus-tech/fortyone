import { CheckIcon, ChevronRightIcon } from "icons";
import type { ReactNode } from "react";
import { Box, ContextMenu, Flex, Text } from "ui";
import type { TextProps } from "ui";

export const ContextMenuItem = ({
  label,
  icon,
  labelColor,
  subMenu,
  shortCut,
  disabled,
  onSelect,
}: {
  label: string;
  icon: ReactNode;
  labelColor?: TextProps["color"];
  disabled?: boolean;
  subMenu?: {
    active?: boolean;
    label: string;
    icon: ReactNode;
    shortCut?: string;
    onSelect?: () => void;
    disabled?: boolean;
  }[];
  shortCut?: string;
  onSelect?: () => void;
}) => {
  return (
    <>
      {subMenu ? (
        <ContextMenu.SubMenu>
          <ContextMenu.SubTrigger
            className="justify-between"
            disabled={disabled}
          >
            <Box className="grid grid-cols-[24px_auto] items-center">
              <span className="text-foreground flex">{icon}</span>
              <Text color={labelColor}>{label}</Text>
            </Box>
            <Flex align="center" gap={3}>
              {shortCut ? (
                <Flex className="text-text-muted text-sm">{shortCut}</Flex>
              ) : null}
              <ChevronRightIcon
                className="text-text-muted h-3.5 w-auto"
                strokeWidth={2.8}
              />
            </Flex>
          </ContextMenu.SubTrigger>
          <ContextMenu.SubItems className="min-w-40" rounded="md">
            <ContextMenu.Group>
              {subMenu.map(
                ({
                  label: subLabel,
                  icon: subIcon,
                  shortCut: subShortCut,
                  onSelect: subOnSelect,
                  disabled: subDisabled,
                  active: subActive,
                }) => (
                  <ContextMenu.Item
                    active={subActive}
                    className="mb-1 justify-between py-1.5"
                    disabled={subDisabled}
                    key={subLabel}
                    onSelect={subOnSelect}
                  >
                    <Box className="grid grid-cols-[24px_auto] items-center gap-1">
                      <span className="text-text-muted flex">{subIcon}</span>
                      <Text className="max-w-40 truncate text-[0.95rem]">
                        {subLabel}
                      </Text>
                    </Box>
                    <Flex align="center" gap={2}>
                      {subShortCut ? (
                        <Flex className="text-text-muted">{subShortCut}</Flex>
                      ) : null}
                      {subActive ? (
                        <CheckIcon className="h-4 w-4 shrink-0" />
                      ) : null}
                    </Flex>
                  </ContextMenu.Item>
                ),
              )}
            </ContextMenu.Group>
          </ContextMenu.SubItems>
        </ContextMenu.SubMenu>
      ) : (
        <ContextMenu.Item
          className="justify-between"
          disabled={disabled}
          onSelect={onSelect}
        >
          <Box className="grid grid-cols-[24px_auto] items-center gap-[2px]">
            <span className="text-text-muted flex">{icon}</span>
            <Text className="max-w-40 truncate" color={labelColor}>
              {label}
            </Text>
          </Box>
          {shortCut ? (
            <Flex className="text-text-muted">{shortCut}</Flex>
          ) : null}
        </ContextMenu.Item>
      )}
    </>
  );
};
