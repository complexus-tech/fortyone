import { cn } from 'lib';
import Link, { LinkProps } from 'next/link';
import { ReactNode } from 'react';

type NavLinkProps = LinkProps & {
  active?: boolean;
  children: ReactNode;
};

export const NavLink = ({ href, active, children, ...rest }: NavLinkProps) => {
  return (
    <Link
      className={cn(
        'flex outline-none ring-gray items-center gap-2 transition-colors duration-200 dark:hover:bg-dark-200/60 hover:bg-gray-50/70 px-2 py-2 rounded-lg text-gray-300 dark:text-gray-200',
        {
          'bg-gray-50 dark:bg-dark-200': active,
        }
      )}
      href={href}
      {...rest}
    >
      {children}
    </Link>
  );
};
