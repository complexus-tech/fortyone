"use client";
import { createContext, ReactNode } from "react";
import { State } from "@/types/states";
import { Objective } from "@/modules/objectives/types";
import { Team } from "@/modules/teams/types";
import { Sprint } from "@/modules/sprints/types";

type Store = {
  teams: Team[];
  states: State[];
  objectives: Objective[];
  sprints: Sprint[];
};

export const StoreContext = createContext<Store | undefined>(undefined);

export const StoreProvider = ({
  children,
  initialState,
}: {
  children: ReactNode;
  initialState: Store;
}) => {
  return (
    <StoreContext.Provider value={initialState}>
      {children}
    </StoreContext.Provider>
  );
};
