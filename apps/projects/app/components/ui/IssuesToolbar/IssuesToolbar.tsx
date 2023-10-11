import { HiViewGrid } from 'react-icons/hi';
import { RiDeleteBin6Line } from 'react-icons/ri';
import { TbMap } from 'react-icons/tb';
import { Button, Flex, Text } from 'ui';

export const IssuesToolbar = () => {
  return (
    <Flex
      gap={2}
      align='center'
      className='sticky border border-gray-100 dark:border-dark-100/50 -translate-x-1/2 rounded-lg py-2 z-50 bg-gray-50/70 dark:bg-dark-200/70 backdrop-blur px-2.5 bottom-6 left-1/2 right-1/2 w-max shadow-lg shadow-dark/10 dark:shadow-dark/20'
    >
      <Text color='muted' className='mr-4 ml-2'>
        2 selected
      </Text>
      <Button
        color='tertiary'
        variant='outline'
        leftIcon={<TbMap className='h-[1.15rem] w-auto dark:text-gray' />}
      >
        Add to sprint
      </Button>
      <Button
        color='tertiary'
        variant='outline'
        leftIcon={<HiViewGrid className='dark:text-gray' />}
      >
        Add to project
      </Button>
      <Button
        color='danger'
        variant='outline'
        className='border-opacity-30'
        leftIcon={<RiDeleteBin6Line />}
      >
        Delete
      </Button>
    </Flex>
  );
};
