import { ArrowRightIcon } from "icons";
import type { ReactNode } from "react";
import { Box, ContextMenu, Flex, Text } from "ui";

export const ContextMenuItem = ({
  label,
  icon,
  subMenu,
  shortCut,
  onSelect,
}: {
  label: string;
  icon: ReactNode;
  subMenu?: {
    label: string;
    icon: ReactNode;
    shortCut?: string;
    onSelect?: () => void;
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
              <span className="flex text-gray dark:text-gray-200">{icon}</span>
              <Text>{label}</Text>
            </Box>
            <Flex align="center" gap={3}>
              {shortCut ? (
                <Flex className="text-sm text-gray dark:text-gray-200">
                  {shortCut}
                </Flex>
              ) : null}
              <ArrowRightIcon
                className="h-3.5 w-auto text-gray dark:text-gray-200"
                strokeWidth={2.8}
              />
            </Flex>
          </ContextMenu.SubTrigger>
          <ContextMenu.SubItems className="min-w-[10rem]" rounded="md">
            <ContextMenu.Group>
              {subMenu.map(
                ({
                  label: subLabel,
                  icon: subIcon,
                  shortCut: subShortCut,
                  onSelect: subOnSelect,
                }) => (
                  <ContextMenu.Item
                    className="mb-1 justify-between py-1.5"
                    key={label}
                    onSelect={subOnSelect}
                  >
                    <Box className="grid grid-cols-[24px_auto] items-center gap-1">
                      <span className="flex text-gray dark:text-gray-200">
                        {subIcon}
                      </span>
                      <Text className="max-w-[10rem] truncate text-[0.95rem]">
                        {subLabel}
                      </Text>
                    </Box>
                    {subShortCut ? (
                      <Flex className="text-gray">{subShortCut}</Flex>
                    ) : null}
                  </ContextMenu.Item>
                ),
              )}
            </ContextMenu.Group>
          </ContextMenu.SubItems>
        </ContextMenu.SubMenu>
      ) : (
        <ContextMenu.Item className="justify-between" onSelect={onSelect}>
          <Box className="grid grid-cols-[24px_auto] items-center gap-[2px]">
            <span className="flex text-gray dark:text-gray-200">{icon}</span>
            <Text className="max-w-[10rem] truncate">{label}</Text>
          </Box>
          {shortCut ? <Flex className="text-gray">{shortCut}</Flex> : null}
        </ContextMenu.Item>
      )}
    </>
  );
};
