import { useState } from "react";
import { useSessionMutation } from "@/lib/use-session-mutation";
import { createDeleteAccountOperation } from "../actions/delete-account";

export const useDeleteAccountMutation = () => {
  const [deleteAccount] = useState(createDeleteAccountOperation);
  return useSessionMutation({ mutationFn: deleteAccount });
};
