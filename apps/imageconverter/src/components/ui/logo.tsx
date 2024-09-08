import Link from "next/link";
import { cn } from "lib";
import { Flex, Text } from "ui";
import logo from "../../../public/logo.png";
import Image from "next/image";

export const Logo = ({ className }: { className?: string }) => {
  return (
    <Link href="/">
      <Flex align="center" className="gap-1.5">
        <Image alt="Logo" className="h-[1.8rem] w-auto" src={logo} />
        <Text className={cn("font-satoshi font-bold", className)} fontSize="lg">
          imageconveta
        </Text>
      </Flex>
    </Link>
  );
};
