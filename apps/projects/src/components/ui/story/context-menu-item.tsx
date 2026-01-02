import { ArrowRightIcon } from "icons";
import type { ReactNode } from "react";
import { Box, ContextMenu, Flex, Text } from "ui";

export const ContextMenuItem = ({
  label,
  icon,
  subMenu,
  shortCut,
  disabled,
  onSelect,
}: {
  label: string;
  icon: ReactNode;
  disabled?: boolean;
  subMenu?: {
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
          <ContextMenu.SubTrigger className="justify-between">
            <Box className="grid grid-cols-[24px_auto] items-center">
              <span className="text-foreground flex">{icon}</span>
              <Text>{label}</Text>
            </Box>
            <Flex align="center" gap={3}>
              {shortCut ? (
                <Flex className="text-text-muted text-sm">{shortCut}</Flex>
              ) : null}
              <ArrowRightIcon
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
                }) => (
                  <ContextMenu.Item
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
                    {subShortCut ? (
                      <Flex className="text-text-muted">{subShortCut}</Flex>
                    ) : null}
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
            <Text className="max-w-40 truncate">{label}</Text>
          </Box>
          {shortCut ? (
            <Flex className="text-text-muted">{shortCut}</Flex>
          ) : null}
        </ContextMenu.Item>
      )}
    </>
  );
};
