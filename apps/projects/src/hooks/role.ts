import { useSession } from "next-auth/react";

export const useUserRole = () => {
  const { data: session } = useSession();
  return {
    userRole: session?.user?.userRole,
  };
};
