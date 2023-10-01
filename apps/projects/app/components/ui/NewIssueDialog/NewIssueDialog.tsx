import { CgArrowsExpandRight } from 'react-icons/cg';
import { IoIosArrowForward } from 'react-icons/io';
import { TbPlus } from 'react-icons/tb';
import { Badge, Button, Dialog, Flex, Switch, Text } from 'ui';

type NewIssueDialogProps = {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
};

export const NewIssueDialog = ({ isOpen, setIsOpen }: NewIssueDialogProps) => {
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <Dialog.Content size='lg' hideClose>
        <Dialog.Header className='flex justify-between items-center px-6 pt-6'>
          <Dialog.Title className='text-lg flex items-center gap-1'>
            <Badge color='tertiary'>COMP-1</Badge>
            <IoIosArrowForward className='h-3 opacity-40 w-auto' />
            <Text color='muted'>New issue</Text>
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
          <Text color='muted'>
            <label className='flex items-center gap-2'>
              Create more <Switch />
            </label>
          </Text>
          <Button size='md' leftIcon={<TbPlus className='h-5 w-auto' />}>
            Create issue
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
