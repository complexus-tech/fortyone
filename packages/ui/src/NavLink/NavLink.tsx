import { cn } from 'lib';
import Link, { LinkProps } from 'next/link';
import { FC, ReactNode } from 'react';

interface Props extends LinkProps {
  children: ReactNode;
  className?: string;
  active?: boolean;
}

export const NavLink: FC<Props> = ({ children, className, active, href }) => {
  const classes = cn(
    'mb-2 transition duration-300 ease-linear flex items-center gap-2 dark:text-white rounded py-4 px-2',
    {
      'text-white bg-primary': active,
      'hover:bg-primary/20 bg-primary/5': !active,
    },
    className
  );
  return (
    <Link href={href} className={classes}>
      {children}
    </Link>
  );
};
