'use client';
import { useState } from 'react';
import { TbPlus } from 'react-icons/tb';
import { Button } from 'ui';
import { NewIssueDialog } from '../../ui';

export const NewIssueButton = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Button
        size='sm'
        onClick={() => setIsOpen(true)}
        leftIcon={<TbPlus className='h-5 w-auto' />}
      >
        New issue
      </Button>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
