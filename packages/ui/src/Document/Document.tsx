import React from 'react';
import { MouseEvent, ReactNode, useState } from 'react';
import { ImFilePdf } from 'react-icons/im';
import { FcRemoveImage } from 'react-icons/fc';
import { cn } from '@/lib';
import { MdOutlineOpenInNew } from 'react-icons/md';
import { RiDeleteBinLine } from 'react-icons/ri';
import { Text } from '../Text/Text';
import { Flex } from '../Flex/Flex';
import { Box } from '../Box/Box';
import { Tooltip } from '../Tooltip/Tooltip';
import { Dialog } from '../Dialog/Dialog';
import { ObjectViewer } from '../ObjectViewer/ObjectViewer';

type FileType = 'image' | 'pdf';

const DocumentWrapper = ({
  children,
  handleClick,
  className,
}: {
  children: ReactNode;
  handleClick: (event: MouseEvent<HTMLDivElement>) => void;
  className?: string;
}) => (
  <Box
    id='document'
    tabIndex={0}
    onClick={handleClick}
    className={cn(
      'group transition duration-200 ease-linear hover:ring ring-primary ring-offset-2 dark:ring-offset-blue-dark rounded-xl cursor-pointer border-[1.5px] dark:border-gray-300 border-gray-100/60 shadow',
      className
    )}
  >
    {children}
  </Box>
);

export const Document = ({
  name,
  src,
  timestamp,
  onDelete,
  className,
}: {
  name: string;
  src: string;
  timestamp?: string;
  onDelete?: () => void;
  className?: string;
}) => {
  const [isOpen, setOpen] = useState(false);
  const [hasError, setError] = useState(false);

  const getDocType = (src: string): FileType => {
    if (src.includes('.pdf') || hasError) return 'pdf';
    return 'image';
  };

  const handleClick = (e: MouseEvent<HTMLDivElement>) => {
    const { tagName } = e.target as HTMLElement;
    if (!['a', 'svg'].includes(tagName.toLowerCase())) {
      setOpen(true);
    }
  };

  return (
    <>
      <DocumentWrapper {...{ src, handleClick, className }}>
        <Box
          className={cn(
            'flex h-44 items-center group justify-center overflow-hidden rounded-t-xl relative',
            {
              'bg-gray-50 dark:bg-blue-darker': getDocType(src) === 'pdf',
              'bg-gray-400': getDocType(src) === 'image',
            }
          )}
        >
          {getDocType(src) === 'pdf' && (
            <ImFilePdf className='h-16 w-auto text-danger' />
          )}
          {getDocType(src) === 'image' && (
            <>
              {hasError ? (
                <Box className='flex flex-col'>
                  <FcRemoveImage className='h-10 w-auto mb-2' />
                  <Text
                    color='muted'
                    align='center'
                    fontWeight='bold'
                    fontSize='sm'
                  >
                    The image failed to load.
                  </Text>
                </Box>
              ) : (
                <img
                  src={src}
                  onError={() => setError(true)}
                  alt={name}
                  className='h-full w-full object-cover rounded-t-xl transition duration-200 ease-linear group-hover:scale-[1.01]'
                />
              )}
            </>
          )}
          <Box
            className={cn(
              'opacity-0 bg-black/20 transition duration-200 ease-linear absolute px-4 pt-6 inset-0 dark:bg-blue-darker/80 backdrop-blur-sm group-hover:opacity-100 rounded-t-lg flex justify-end items-start',
              {
                'justify-between': !!onDelete,
              }
            )}
          >
            {!!onDelete && (
              <Tooltip title='Delete document'>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete();
                  }}
                  className='h-12 aspect-square rounded-full bg-danger/80 transition hover:ring-[0.17rem] ring-danger ring-offset-2 ring-offset-transparent flex items-center justify-center shadow border-2 border-danger/50 dark:text-white'
                >
                  <span className='sr-only'>Delete</span>
                  <RiDeleteBinLine className='h-6 w-auto text-white' />
                </button>
              </Tooltip>
            )}

            <Tooltip title='Open in new tab'>
              <a
                href={src}
                target='_blank'
                className='h-12 shadow bg-primary/80 text-white aspect-square rounded-full flex items-center justify-center transition hover:ring-[0.17rem] ring-primary ring-offset-2 ring-offset-transparent'
              >
                <MdOutlineOpenInNew className='h-6 w-auto' />
              </a>
            </Tooltip>
          </Box>
        </Box>
        <Box className='border-t-[1.5px] border-gray-100/60 dark:border-gray-300 px-3 pt-3 pb-2'>
          <Flex justify='between' className='mb-1'>
            <Text
              title={name}
              fontWeight='bold'
              fontSize='sm'
              className='max-w-[13rem] truncate mb-1'
            >
              {name}
            </Text>
            <span className='rounded-lg bg-gray-100 dark:text-white dark:bg-blue-darker px-2 pb-[1px] text-sm font-semibold uppercase tracking-wider'>
              File
            </span>
          </Flex>

          <Flex justify='between' align='center' className='my-2'>
            <Text
              transform='uppercase'
              color='muted'
              fontWeight='medium'
              as='span'
              fontSize='sm'
            >
              {getDocType(src)}
            </Text>
            {timestamp && (
              <Text color='muted' as='span' fontSize='sm'>
                Created: {new Date(timestamp).toLocaleString()}
              </Text>
            )}
          </Flex>
        </Box>
      </DocumentWrapper>

      <Dialog open={isOpen} onOpenChange={setOpen}>
        <Dialog.Content size='lg'>
          <Dialog.Header>
            <Dialog.Title variant='bordered' className='mb-0'>
              {name}
            </Dialog.Title>
          </Dialog.Header>
          {getDocType(src) === 'image' ? (
            <Box className='min-h-[40vh] overflow-y-auto max-h-[70vh] rounded-b overflow-hidden flex justify-center items-center'>
              {/* {hasError ? (
              <Box className='flex flex-col relative -top-4'>
                <FcRemoveImage className='h-20 w-auto mb-2' />
                <Text color='muted' fontWeight='bold' align='center'>
                  The was an error loading the image.
                </Text>
              </Box>
            ) : (
              <img src={src} alt={name} className='w-full h-auto' />
            )} */}
              <img src={src} alt={name} className='w-full h-auto' />
            </Box>
          ) : (
            <ObjectViewer
              type='application/pdf'
              data={src}
              className='min-h-[80vh]'
            />
          )}
        </Dialog.Content>
      </Dialog>
    </>
  );
};
