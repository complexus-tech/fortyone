import { useSession } from "next-auth/react";

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
  if (!entityOwnerId) {
    return {
      isAdminOrOwner: session?.user?.userRole === "admin",
    };
  }
  return {
    isAdminOrOwner:
      session?.user?.userRole === "admin" ||
      session?.user?.id === entityOwnerId,
  };
};
