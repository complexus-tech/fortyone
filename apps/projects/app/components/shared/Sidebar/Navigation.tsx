import { usePathname } from 'next/navigation';
import { HiChartBar } from 'react-icons/hi';
import { HiOutlineInboxArrowDown } from 'react-icons/hi2';
import {
  TbBoxMultiple,
  TbFocusCentered,
  TbLayoutDashboard,
} from 'react-icons/tb';
import { Flex } from 'ui';
import { NavLink } from '../../ui';

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: 'Dashboard',
      icon: <TbLayoutDashboard className='h-5 w-auto' />,
      href: '/',
    },
    {
      name: 'Analytics',
      icon: <HiChartBar className='h-5 w-auto' />,
      href: '/analytics',
    },
    {
      name: 'Inbox',
      icon: (
        <HiOutlineInboxArrowDown strokeWidth={2.3} className='h-5 w-auto' />
      ),
      href: '/inbox',
    },
    {
      name: 'My issues',
      icon: <TbFocusCentered strokeWidth={2.3} className='h-5 w-auto' />,
      href: '/my-issues',
    },
    {
      name: 'Views',
      icon: <TbBoxMultiple className='h-5 w-auto' />,
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
