'use client';
import { cn } from 'lib';
import { useState } from 'react';
import { HiViewGrid } from 'react-icons/hi';
import { TbCaretDownFilled, TbPlus } from 'react-icons/tb';
import { Box, Button, Flex } from 'ui';
import { Project } from './Project';

export const Projects = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Box className='mt-8'>
      <Button
        color='tertiary'
        align='between'
        fullWidth
        size='sm'
        variant='naked'
        className='group'
        onClick={() => setIsOpen(!isOpen)}
      >
        <span className='flex items-center gap-1'>
          <HiViewGrid className='relative h-[1.1rem] text-gray-300/80 dark:text-gray-200 w-auto -top-[0.1px]' />
          Projects
          <TbCaretDownFilled
            className={cn(
              'text-gray-300/60 dark:text-gray-200 transition-transform -rotate-90',
              {
                'rotate-0': isOpen,
              }
            )}
          />
        </span>

        <TbPlus className='h-5 hidden text-gray-300/60 dark:dark:text-gray-200 justify-self-end w-auto group-hover:inline' />
      </Button>
      <Flex
        direction='column'
        className={cn(
          'h-0 overflow-y-auto transition-all duration-300 max-h-[55vh]',
          {
            'h-max mt-1': isOpen,
          }
        )}
      >
        <Project icon='🚀' name='Website design' />
        <Project icon='🇦🇫' name='Data migration' />
        <Project icon='🏀' name='CRM development' />
      </Flex>
    </Box>
  );
};
