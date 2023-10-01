import Link from 'next/link';
import { IoIosArrowForward } from 'react-icons/io';

import { cn } from 'lib';
import { Flex } from '../Flex/Flex';
interface BreadCrumb {
  name: string;
  url?: string;
}

export type BreadCrumbsProps = {
  breadCrumbs: BreadCrumb[];
  className?: string;
};

export const BreadCrumbs = ({ breadCrumbs, className }: BreadCrumbsProps) => {
  return (
    <Flex className={cn('gap-2', className)}>
      {breadCrumbs.map(({ name, url = '' }, idx) => (
        <Link
          key={idx}
          href={url}
          className={cn(
            'flex items-center gap-2 capitalize text-gray-300 transition dark:text-gray-200',
            {
              'text-gray dark:text-gray': idx + 1 === breadCrumbs.length,
            }
          )}
        >
          <span className='hover:text-primary'>
            {name?.toLocaleLowerCase()}
          </span>
          <IoIosArrowForward
            className={cn('h-[0.8rem] w-auto', {
              hidden: idx + 1 === breadCrumbs.length,
            })}
          />
        </Link>
      ))}
    </Flex>
  );
};
