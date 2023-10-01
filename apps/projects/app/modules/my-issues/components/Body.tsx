import { HiViewGrid } from 'react-icons/hi';
import { RiDeleteBin6Line } from 'react-icons/ri';
import { TbMap } from 'react-icons/tb';
import { Box, Button, Text } from 'ui';
import { BodyContainer } from '../../../components/shared';
import { Issue, IssueHeader } from '../../../components/ui';

export const Body = () => {
  return (
    <BodyContainer className='relative'>
      <IssueHeader count={3} status='Backlog' />
      {new Array(3).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Backlog'
          title='These issues are not assigned to any sprint.'
        />
      ))}
      <IssueHeader count={2} status='Todo' />
      {new Array(2).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Todo'
          title='These issues are at the top of the backlog and are ready to be worked on.'
        />
      ))}
      <IssueHeader count={3} status='In Progress' />
      {new Array(3).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='In Progress'
          title='These issues are being actively worked on.'
        />
      ))}
      <IssueHeader count={3} status='Testing' />
      {new Array(3).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Testing'
          title='These issues are being tested by the QA team.'
        />
      ))}
      <IssueHeader count={4} status='Done' />
      {new Array(4).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Done'
          title='These issues are completed and ready to be deployed.'
        />
      ))}
      <IssueHeader count={2} status='Canceled' />
      {new Array(2).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Canceled'
          title='These issues are no longer being worked on.'
        />
      ))}

      <Box className='sticky flex gap-2 items-center border border-gray-100 dark:border-dark-100/50 -translate-x-1/2 rounded-lg py-2 z-50 bg-gray-50/70 dark:bg-dark-200/70 backdrop-blur px-2.5 bottom-3 left-1/2 right-1/2 w-max'>
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
      </Box>
    </BodyContainer>
  );
};
