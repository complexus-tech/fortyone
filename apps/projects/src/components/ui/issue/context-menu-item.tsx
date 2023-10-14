import type { ReactNode } from 'react';
import { FaCaretRight } from 'react-icons/fa';
import { Box, ContextMenu, Flex, Text } from 'ui';

export const ContextMenuItem = ({
  label,
  icon,
  subMenu,
  shortCut,
}: {
  label: string;
  icon: ReactNode;
  subMenu?: { label: string; icon: ReactNode; shortCut?: string }[];
  shortCut?: string;
}) => {
  return (
    <>
      {subMenu ? (
        <ContextMenu.SubMenu>
          <ContextMenu.SubTrigger className='justify-between'>
            <Box className='grid grid-cols-[24px_auto] items-center'>
              <span className='text-gray-300/70 dark:text-gray-200 flex'>
                {icon}
              </span>
              <Text>{label}</Text>
            </Box>
            <Flex align='center' gap={3}>
              {shortCut ? (
                <Flex className='text-gray text-sm'>{shortCut}</Flex>
              ) : null}
              <FaCaretRight className='text-gray' strokeWidth={2.1} />
            </Flex>
          </ContextMenu.SubTrigger>
          <ContextMenu.SubItems className='min-w-[10rem]' rounded='md'>
            <ContextMenu.Group>
              {subMenu.map(
                ({ label: subLabel, icon: subIcon, shortCut: subShortCut }) => (
                  <ContextMenu.Item
                    className='justify-between py-1.5 mb-1'
                    key={label}
                  >
                    <Box className='grid grid-cols-[24px_auto] gap-1 items-center'>
                      <span className='text-gray-300/70 dark:text-gray-200 flex'>
                        {subIcon}
                      </span>
                      <Text className='text-[0.95rem] max-w-[10rem] truncate'>
                        {subLabel}
                      </Text>
                    </Box>
                    {subShortCut ? (
                      <Flex className='text-gray'>{subShortCut}</Flex>
                    ) : null}
                  </ContextMenu.Item>
                )
              )}
            </ContextMenu.Group>
          </ContextMenu.SubItems>
        </ContextMenu.SubMenu>
      ) : (
        <ContextMenu.Item className='justify-between'>
          <Box className='grid grid-cols-[24px_auto] gap-[2px] items-center'>
            <span className='text-gray-300/70 dark:text-gray-200 flex'>
              {icon}
            </span>
            <Text className='max-w-[10rem] truncate'>{label}</Text>
          </Box>
          {shortCut ? <Flex className='text-gray'>{shortCut}</Flex> : null}
        </ContextMenu.Item>
      )}
    </>
  );
};
