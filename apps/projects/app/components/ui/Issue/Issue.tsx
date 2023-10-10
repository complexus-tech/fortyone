import Link from 'next/link';
import { GoDotFill } from 'react-icons/go';
import { Avatar, Button, Checkbox, Container, Flex, Text } from 'ui';
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
  return (
    <Container className='group border-b hover:bg-gray-50/50 hover:dark:bg-dark-200/50 py-3 dark:border-dark-200 border-gray-50 flex items-center justify-between'>
      <Flex align='center' gap={2} className='relative'>
        <Checkbox className='absolute -left-7 hidden data-[state=checked]:inline-block group-hover:inline-block' />

        <PriorityIcon priority={priority} />

        <Text color='muted' className='w-[55px] truncate'>
          COM-12
        </Text>

        <IssueStatusIcon status={status} />
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
        <Avatar
          name='Joseph Mukorivo'
          color='gray'
          size='sm'
          src='https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo'
        />
      </Flex>
    </Container>
  );
};
