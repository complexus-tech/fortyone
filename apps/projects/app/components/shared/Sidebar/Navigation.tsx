import { usePathname } from 'next/navigation';
import { HiChartBar } from 'react-icons/hi';
import { HiOutlineInboxArrowDown } from 'react-icons/hi2';
import { TbFocusCentered, TbLayoutDashboard, TbStack2 } from 'react-icons/tb';
import { Flex } from 'ui';
import { NavLink } from '../../ui';

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: 'Dashboard',
      icon: <TbLayoutDashboard className='h-5 w-auto dark:text-gray' />,
      href: '/',
    },
    {
      name: 'Analytics',
      icon: <HiChartBar className='h-5 w-auto dark:text-gray' />,
      href: '/analytics',
    },
    {
      name: 'Inbox',
      icon: (
        <HiOutlineInboxArrowDown
          strokeWidth={2.3}
          className='h-5 w-auto dark:text-gray'
        />
      ),
      href: '/inbox',
    },
    {
      name: 'My issues',
      icon: (
        <TbFocusCentered
          strokeWidth={2.3}
          className='h-5 w-auto dark:text-gray'
        />
      ),
      href: '/my-issues',
    },
    {
      name: 'Views',
      icon: <TbStack2 className='h-[1.4rem] w-auto dark:text-gray' />,
      href: '/views',
    },
  ];

  return (
    <Flex direction='column' gap={2}>
      {links.map(({ name, icon, href }) => (
        <NavLink active={pathname === href} href={href} key={name}>
          {icon}
          {name}
        </NavLink>
      ))}
    </Flex>
  );
};
