"use client";

import type { ReactNode } from "react";
import { createContext, useContext } from "react";
import type { Session } from "@/auth";

type SessionStatus = "authenticated" | "unauthenticated";

type SessionContextValue = {
  data: Session | null;
  status: SessionStatus;
};

const SessionContext = createContext<SessionContextValue>({
  data: null,
  status: "unauthenticated",
});

export const SessionProvider = ({
  children,
  session,
}: {
  children: ReactNode;
  session: Session | null;
}) => {
  return (
    <SessionContext.Provider
      value={{
        data: session,
        status: session ? "authenticated" : "unauthenticated",
      }}
    >
      {children}
    </SessionContext.Provider>
  );
};

export const useSession = () => useContext(SessionContext);
