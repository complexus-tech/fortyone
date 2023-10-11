'use client';
import { cn } from 'lib';
import Link from 'next/link';
import { Fragment } from 'react';
import { FaCaretRight } from 'react-icons/fa';
import { FiEdit } from 'react-icons/fi';
import { GoDotFill } from 'react-icons/go';
import { HiViewGrid } from 'react-icons/hi';
import {
  TbCalendarPlus,
  TbCheck,
  TbCopy,
  TbFolder,
  TbLink,
  TbMap,
  TbPencil,
  TbStar,
  TbTag,
  TbTrash,
  TbUserShare,
} from 'react-icons/tb';
import {
  Avatar,
  Box,
  Button,
  Checkbox,
  Container,
  ContextMenu,
  Flex,
  Menu,
  Text,
} from 'ui';
import { IssueStatusIcon } from '../IssueStatusIcon/IssueStatusIcon';
import { PriorityIcon } from '../PriorityIcon/PriorityIcon';

export type IssueStatus =
  | 'Backlog'
  | 'Todo'
  | 'In Progress'
  | 'Testing'
  | 'Done'
  | 'Duplicate'
  | 'Canceled';

export type IssuePriority =
  | 'No Priority'
  | 'Urgent'
  | 'High'
  | 'Medium'
  | 'Low';

type IssueProps = {
  status?: IssueStatus;
  title: string;
  description?: string;
  priority?: IssuePriority;
};
export const Issue = ({
  title,
  status = 'Backlog',
  priority = 'No Priority',
}: IssueProps) => {
  const priorities = [
    'No Priority',
    'Urgent',
    'High',
    'Medium',
    'Low',
  ] as const;
  const statuses = [
    'Backlog',
    'Todo',
    'In Progress',
    'Testing',
    'Done',
    'Duplicate',
    'Canceled',
  ] as const;

  const users = [
    {
      name: 'Joseph Mukorivo',
      avatar:
        'https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo',
    },
    {
      name: 'Jane Doe',
      avatar:
        'https://images.unsplash.com/photo-1677576874778-df95ea6ff733?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHx0b3BpYy1mZWVkfDI4fHRvd0paRnNrcEdnfHxlbnwwfHx8fHw%3D&auto=format&fit=crop&w=800&q=60',
    },
    {
      name: 'John Doe',
      avatar:
        'https://images.unsplash.com/photo-1696452044585-c6a9389d0c6b?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHx0b3BpYy1mZWVkfDM3fHRvd0paRnNrcEdnfHxlbnwwfHx8fHw%3D&auto=format&fit=crop&w=800&q=60',
    },

    {
      name: 'Doubting Thomas',
    },
  ];

  const contextMenu = [
    {
      name: 'Main',
      options: [
        {
          label: 'Status',
          icon: <IssueStatusIcon />,
        },
        {
          label: 'Assignee',
          icon: <TbUserShare className='h-5 w-auto' />,
        },
        {
          label: 'Priority',
          icon: <PriorityIcon priority='High' />,
        },
        {
          label: 'Labels',
          icon: <TbTag className='h-5 w-auto' />,
        },
        {
          label: 'Sprint',
          icon: <TbMap className='h-5 w-auto' />,
        },
        {
          label: 'Module',
          icon: <TbFolder className='h-5 w-auto' />,
        },
        {
          label: 'Edit',
          icon: <TbPencil className='h-5 w-auto' />,
        },
        {
          label: 'Rename',
          icon: <FiEdit className='h-5 w-auto' />,
        },
      ],
    },
    {
      name: 'More',
      options: [
        {
          label: 'Project',
          icon: <HiViewGrid className='h-5 w-auto' />,
        },
        {
          label: 'Add to sprint',
          icon: <TbTag className='h-5 w-auto' />,
        },
        {
          label: 'Due Date',
          icon: <TbCalendarPlus className='h-5 w-auto' />,
        },
        {
          label: 'Start Date',
          icon: <TbCalendarPlus className='h-5 w-auto' />,
        },
        {
          label: 'Clone',
          icon: <TbCopy className='h-5 w-auto' />,
        },
        {
          label: 'Favorite',
          icon: <TbStar className='h-5 w-auto' />,
        },
        {
          label: 'Copy',
          icon: <TbLink className='h-5 w-auto' />,
        },
      ],
    },
    {
      name: 'Danger Zone',
      options: [
        {
          label: 'Delete',
          icon: <TbTrash className='h-5 w-auto text-danger' />,
        },
      ],
    },
  ];

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Container
          tabIndex={0}
          className='group border-b focus:bg-gray-50/50 focus:dark:bg-dark-200/50 hover:bg-gray-50/50 hover:dark:bg-dark-200/50 py-3 dark:border-dark-200 border-gray-50 flex items-center justify-between'
        >
          <Flex align='center' gap={2} className='relative'>
            <Checkbox className='absolute -left-7 hidden data-[state=checked]:inline-block group-hover:inline-block' />

            <Menu>
              <Menu.Button>
                <Button
                  color='tertiary'
                  variant='naked'
                  className='p-0 h-max dark:hover:bg-transparent hover:bg-transparent focus:bg-transparent dark:focus:bg-transparent'
                  size='sm'
                  leftIcon={<PriorityIcon priority={priority} />}
                >
                  <span className='sr-only'>Change priority</span>
                </Button>
              </Menu.Button>
              <Menu.Items align='center' className='w-64'>
                <Menu.Group className='mb-2 px-4'>
                  <input
                    className='bg-transparent py-1 outline-none w-full'
                    placeholder='Change priority'
                    autoFocus
                  />
                </Menu.Group>
                <Menu.Separator />
                <Menu.Group>
                  {priorities.map((pr, idx) => (
                    <Menu.Item
                      key={pr}
                      active={pr === priority}
                      className='justify-between'
                    >
                      <Box className='grid grid-cols-[24px_auto] items-center'>
                        <PriorityIcon
                          className={cn({
                            'text-gray relative left-[1px]': pr === 'Urgent',
                          })}
                          priority={pr}
                        />
                        <Text>{pr}</Text>
                      </Box>
                      <Flex align='center' gap={2}>
                        {pr === priority && (
                          <TbCheck className='h-5 w-auto' strokeWidth={2.1} />
                        )}
                        <Text color='muted'>{idx}</Text>
                      </Flex>
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>

            <Text color='muted' className='w-[55px] truncate'>
              COM-12
            </Text>

            <Menu>
              <Menu.Button>
                <Button
                  color='tertiary'
                  variant='naked'
                  className='p-0 h-max dark:hover:bg-transparent hover:bg-transparent focus:bg-transparent dark:focus:bg-transparent'
                  size='sm'
                  leftIcon={<IssueStatusIcon status={status} />}
                >
                  <span className='sr-only'>Change status</span>
                </Button>
              </Menu.Button>
              <Menu.Items align='center' className='w-64'>
                <Menu.Group className='mb-2 px-4'>
                  <input
                    className='bg-transparent py-1 outline-none w-full'
                    placeholder='Change status'
                    autoFocus
                  />
                </Menu.Group>
                <Menu.Separator />
                <Menu.Group>
                  {statuses.map((st, idx) => (
                    <Menu.Item
                      key={st}
                      active={st === status}
                      className='justify-between'
                    >
                      <Box className='grid grid-cols-[24px_auto] items-center'>
                        <IssueStatusIcon status={st} />
                        <Text>{st}</Text>
                      </Box>
                      <Flex align='center' gap={2}>
                        {st === status && (
                          <TbCheck className='h-5 w-auto' strokeWidth={2.1} />
                        )}
                        <Text color='muted'>{idx}</Text>
                      </Flex>
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
            <Link href='/issue'>
              <Text className='whitespace-nowrap text-ellipsis overflow-hidden w-[70vh] hover:opacity-90'>
                {title}
              </Text>
            </Link>
          </Flex>
          <Flex align='center' gap={3}>
            <Flex align='center' gap={1}>
              <Button
                size='xs'
                rounded='xl'
                color='tertiary'
                variant='outline'
                leftIcon={<GoDotFill className='text-info' />}
              >
                Feature
              </Button>
              <Button
                size='xs'
                rounded='xl'
                color='tertiary'
                variant='outline'
                leftIcon={<GoDotFill className='text-danger' />}
              >
                Bug
              </Button>
            </Flex>

            <Text color='muted'>Sep 27</Text>

            <Menu>
              <Menu.Button>
                <Button
                  color='tertiary'
                  variant='naked'
                  className='px-1 select-none'
                  size='sm'
                  leftIcon={
                    <Avatar
                      name='Joseph Mukorivo'
                      color='gray'
                      size='sm'
                      src='https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo'
                    />
                  }
                >
                  <span className='sr-only'>Assign user</span>
                </Button>
              </Menu.Button>
              <Menu.Items align='end' className='w-72'>
                <Menu.Group className='mb-2 px-4'>
                  <input
                    className='bg-transparent py-1 outline-none w-full'
                    placeholder='Assign user'
                    autoFocus
                  />
                </Menu.Group>
                <Menu.Separator />
                <Menu.Group>
                  {users.map(({ name, avatar }, idx) => (
                    <Menu.Item
                      key={idx}
                      active={idx === 1}
                      className='justify-between'
                    >
                      <Flex align='center' gap={2}>
                        <Avatar
                          name={name}
                          color='primary'
                          size='sm'
                          src={avatar}
                        />
                        <Text className='truncate max-w-[10rem]'>{name}</Text>
                      </Flex>
                      <Flex align='center' gap={1}>
                        {idx === 1 && (
                          <TbCheck className='h-5 w-auto' strokeWidth={2.1} />
                        )}
                        <Text color='muted'>{idx}</Text>
                      </Flex>
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        </Container>
      </ContextMenu.Trigger>
      <ContextMenu.Items className='w-72'>
        {contextMenu.map(({ name, options }) => (
          <Fragment key={name}>
            <ContextMenu.Group key={name}>
              {options.map(({ label, icon }, idx) => (
                <ContextMenu.Item key={idx} className='justify-between'>
                  <Box className='grid grid-cols-[24px_auto] items-center'>
                    {icon}
                    <Text>{label}</Text>
                  </Box>
                  <FaCaretRight className='text-gray' strokeWidth={2.1} />
                </ContextMenu.Item>
              ))}
            </ContextMenu.Group>
            {name !== 'Danger Zone' && (
              <ContextMenu.Separator className='my-2' />
            )}
          </Fragment>
        ))}
      </ContextMenu.Items>
    </ContextMenu>
  );
};
