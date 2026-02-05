"use client";

import { useEffect, useRef } from "react";
import { useSession } from "next-auth/react";
import { exchangeSessionToken } from "@/lib/http/exchange-session";

export const SessionExchange = () => {
  const { data: session } = useSession();
  const lastTokenRef = useRef<string | null>(null);

  useEffect(() => {
    if (!session?.token) return;
    if (lastTokenRef.current === session.token) return;
    lastTokenRef.current = session.token;
    void exchangeSessionToken(session.token);
  }, [session?.token]);

  return null;
};
