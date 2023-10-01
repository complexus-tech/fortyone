import { useState } from 'react';
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
import { Avatar, Button, Flex, Menu, Text } from 'ui';
import { NewIssueDialog } from '../../ui';

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
              className='px-1'
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
          className='px-[0.6rem]'
          leftIcon={<TbSearch className='h-5 w-auto' />}
          variant='outline'
        >
          <span className='sr-only'>Search</span>
        </Button>
      </Flex>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
