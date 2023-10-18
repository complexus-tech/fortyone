import { usePathname } from "next/navigation";
import { HiChartBar, HiViewGrid } from "react-icons/hi";
import { HiOutlineInboxArrowDown } from "react-icons/hi2";
import { TbFocusCentered, TbLayoutDashboard } from "react-icons/tb";
import { Flex } from "ui";
import { NavLink } from "@/components/ui";

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: "Dashboard",
      icon: <TbLayoutDashboard className="h-5 w-auto dark:text-gray" />,
      href: "/",
    },
    {
      name: "Analytics",
      icon: <HiChartBar className="h-5 w-auto dark:text-gray" />,
      href: "/analytics",
    },
    {
      name: "Inbox",
      icon: (
        <HiOutlineInboxArrowDown
          className="h-5 w-auto dark:text-gray"
          strokeWidth={2.3}
        />
      ),
      href: "/inbox",
    },
    {
      name: "My issues",
      icon: (
        <TbFocusCentered
          className="h-5 w-auto dark:text-gray"
          strokeWidth={2.3}
        />
      ),
      href: "/my-issues",
    },
    {
      name: "Projects",
      icon: <HiViewGrid className="h-[1.4rem] w-auto dark:text-gray" />,
      href: "/projects",
    },
  ];

  return (
    <Flex direction="column" gap={2}>
      {links.map(({ name, icon, href }) => (
        <NavLink active={pathname === href} href={href} key={name}>
          {icon}
          {name}
        </NavLink>
      ))}
    </Flex>
  );
};
