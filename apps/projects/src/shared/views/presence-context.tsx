"use client";

import type { ReactNode } from "react";
import { createContext, useContext } from "react";

export type ViewsPresence = {
  hasViews: boolean;
  hasTeamViews: (teamId: string) => boolean;
};

const ViewsPresenceContext = createContext<ViewsPresence>({
  hasViews: false,
  hasTeamViews: () => false,
});

export const ViewsPresenceProvider = ({
  children,
  value,
}: {
  children: ReactNode;
  value: ViewsPresence;
}) => (
  <ViewsPresenceContext.Provider value={value}>
    {children}
  </ViewsPresenceContext.Provider>
);

export const useViewsPresence = () => useContext(ViewsPresenceContext);
