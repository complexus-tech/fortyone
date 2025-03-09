import { useSession } from "next-auth/react";
import { useUserRole } from "./role";

export const useIsOwner = (entityOwnerId?: string) => {
  const { data: session } = useSession();
  if (!entityOwnerId) {
    return {
      isEntityOwner: false,
    };
  }
  return {
    isEntityOwner: session?.user?.id === entityOwnerId,
  };
};

export const useIsAdminOrOwner = (entityOwnerId?: string) => {
  const { data: session } = useSession();
  const { userRole } = useUserRole();
  if (!entityOwnerId) {
    return {
      isAdminOrOwner: userRole === "admin",
    };
  }
  return {
    isAdminOrOwner: userRole === "admin" || session?.user?.id === entityOwnerId,
  };
};
