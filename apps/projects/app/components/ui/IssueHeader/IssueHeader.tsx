'use client';
import { useState } from 'react';
import { TbPlus } from 'react-icons/tb';
import { Button, Container, Flex, Text, Tooltip } from 'ui';
import { IssueStatusIcon } from '../IssueStatusIcon/IssueStatusIcon';
import { NewIssueDialog } from '../NewIssueDialog/NewIssueDialog';

type IssueHeaderProps = {
  status?: 'Backlog' | 'Todo' | 'In Progress' | 'Testing' | 'Done' | 'Canceled';
  count: number;
};
export const IssueHeader = ({
  count,
  status = 'Backlog',
}: IssueHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Container className='bg-gray-50 dark:bg-dark-200 py-1'>
      <Flex justify='between' align='center'>
        <Flex align='center' gap={2}>
          <IssueStatusIcon status={status} />
          <Text fontWeight='medium'>{status}</Text>
          <Text color='muted'>{count}</Text>
        </Flex>
        <Tooltip title='Add issue' side='left'>
          <Button
            color='tertiary'
            variant='naked'
            className='p-0'
            onClick={() => setIsOpen(true)}
            leftIcon={<TbPlus className='h-5 w-auto dark:text-gray-200' />}
          >
            <span className='sr-only'>Add issue</span>
          </Button>
        </Tooltip>
      </Flex>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </Container>
  );
};
