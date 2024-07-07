"use client";

import { createContext, ReactNode, useContext } from "react";
import { State } from "@/types/states";

type Store = {
  states: State[];
};

const StoreContext = createContext<Store | undefined>(undefined);

export const StoreProvider = ({
  children,
  states,
}: {
  children: ReactNode;
  states: State[];
}) => {
  const initialState: Store = {
    states,
  };
  return (
    <StoreContext.Provider value={initialState}>
      {children}
    </StoreContext.Provider>
  );
};

export const useStore = () => {
  const context = useContext(StoreContext);
  if (!context) {
    throw new Error("useStore must be used within a StoreProvider");
  }
  return context;
};
