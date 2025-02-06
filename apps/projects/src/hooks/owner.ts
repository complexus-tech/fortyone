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
