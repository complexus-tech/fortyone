import { useUserRole } from "@/hooks";

export const useCanUpdateObjective = () => {
  const { userRole } = useUserRole();

  // Match story editing: guests are read-only; authenticated workspace roles
  // remain interactive while the workspace query resolves.
  return userRole !== "guest";
};
