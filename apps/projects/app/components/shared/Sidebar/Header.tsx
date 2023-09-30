import { useState } from 'react';
import { CgArrowsExpandRight } from 'react-icons/cg';
import { IoIosArrowForward } from 'react-icons/io';
import {
  TbCheck,
  TbLayoutGridAdd,
  TbLogout2,
  TbPlus,
  TbSearch,
  TbSettings,
  TbUser,
  TbUsersPlus,
} from 'react-icons/tb';
import { Avatar, Badge, Button, Dialog, Flex, Menu, Switch, Text } from 'ui';

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Flex align='center' className='h-16' justify='between'>
        <Menu>
          <Menu.Button>
            <Button
              color='tertiary'
              variant='naked'
              className='pl-1'
              size='sm'
              leftIcon={
                <Avatar name='Complexus Technologies' rounded='md' size='sm' />
              }
            >
              Complexus
            </Button>
          </Menu.Button>
          <Menu.Items align='start' className='w-72'>
            <Menu.Group className='mb-2 px-4'>
              <Text>Workspaces</Text>
            </Menu.Group>
            <Menu.Group>
              <Menu.Item className='justify-between'>
                <span className='flex gap-2 items-center'>
                  <Avatar
                    name='Complexus Technologies'
                    rounded='md'
                    size='sm'
                  />
                  Complexus
                </span>
                <TbCheck
                  className='h-5 w-auto text-primary'
                  strokeWidth={2.1}
                />
              </Menu.Item>
              <Menu.Item>
                <Avatar
                  color='secondary'
                  name='fin Kenya'
                  rounded='md'
                  size='sm'
                />
                Fin Kenya
              </Menu.Item>
              <Menu.Item asChild>
                <Button
                  color='tertiary'
                  leftIcon={<TbPlus className='h-5 w-auto' />}
                  variant='naked'
                >
                  Create workspace
                </Button>
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <TbSettings className='h-5 w-auto' />
                Workspace settings
              </Menu.Item>
              <Menu.Item>
                <TbUsersPlus className='h-5 w-auto' />
                Invite members
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <TbLogout2 className='h-5 w-auto' />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>

        <Menu>
          <Menu.Button>
            <Button
              color='tertiary'
              variant='naked'
              size='sm'
              leftIcon={
                <Avatar
                  name='Joseph Mukorivo'
                  size='sm'
                  src='https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo'
                />
              }
            >
              <span className='sr-only'>Joseph Mukorivo</span>
            </Button>
          </Menu.Button>
          <Menu.Items align='start'>
            <Menu.Group className='mb-3 px-4 mt-2'>
              <Text fontSize='sm' fontWeight='medium'>
                josemukorivo@gmail.com
              </Text>
            </Menu.Group>
            <Menu.Group>
              <Menu.Item>
                <TbUser className='h-5 w-auto' />
                View profile
              </Menu.Item>
              <Menu.Item>
                <TbSettings className='h-5 w-auto' />
                Settings
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <TbLogout2 className='h-5 w-auto' />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Flex align='center' className='mb-4' gap={2} justify='between'>
        <Button
          color='tertiary'
          fullWidth
          leftIcon={<TbLayoutGridAdd className='h-5 w-auto' />}
          variant='outline'
          onClick={() => setIsOpen(!isOpen)}
        >
          New issue
        </Button>
        <Button
          color='tertiary'
          align='center'
          leftIcon={<TbSearch className='h-5 w-auto' />}
          variant='outline'
        >
          <span className='sr-only'>Search</span>
        </Button>
      </Flex>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <Dialog.Content size='lg' hideClose>
          <Dialog.Header className='flex justify-between items-center px-6 pt-6'>
            <Dialog.Title className='text-lg flex items-center gap-1'>
              <Badge color='tertiary'>COMP-1</Badge>
              <IoIosArrowForward className='h-4 w-auto' />
              <Text fontSize='sm'>New issue</Text>
            </Dialog.Title>
            <Flex gap={4}>
              <Button
                color='tertiary'
                variant='naked'
                size='xs'
                href='/'
                className='px-[0.35rem] dark:hover:bg-dark-100'
              >
                <CgArrowsExpandRight className='h-[1.2rem] w-auto' />
                <span className='sr-only'>Expand issue to full screen</span>
              </Button>
              <Dialog.Close />
            </Flex>
          </Dialog.Header>
          <Dialog.Body>
            <textarea
              className='bg-transparent py-2 resize-none outline-none w-full text-2xl'
              placeholder='Issue title'
              autoComplete='off'
              spellCheck={false}
            />
            <textarea
              className='bg-transparent min-h-[5rem] text-gray-200/80 mb-4 outline-none w-full text-lg resize-none'
              placeholder='Issue description'
            />
            <Flex gap={1}>
              <Badge color='tertiary'>COMP-1</Badge>
              <Badge color='tertiary'>COMP-1</Badge>
              <Badge color='tertiary'>COMP-1</Badge>
              <Badge color='tertiary'>COMP-1</Badge>
              <Badge color='tertiary'>COMP-1</Badge>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer className='flex justify-between gap-2 items-center'>
            <Text color='muted' className='flex items-center gap-2'>
              Create more <Switch />
            </Text>
            <Button size='md' leftIcon={<TbPlus className='h-5 w-auto' />}>
              Create issue
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
