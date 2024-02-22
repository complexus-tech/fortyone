import type { ReactNode } from "react";

export const KanbanList = ({ children }: { children: ReactNode }) => {
  return (
    <div className="flex h-[calc(100vh-7.5rem)] flex-col gap-3 overflow-y-auto pb-6">
      {children}
    </div>
  );
};
