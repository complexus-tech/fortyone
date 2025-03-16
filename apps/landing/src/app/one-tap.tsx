/* eslint-disable no-console -- Google One Tap is not typed */
/* eslint-disable @typescript-eslint/no-explicit-any -- Google One Tap is not typed */
"use client";

import { useEffect, useCallback, useState } from "react";
import { useSession } from "next-auth/react";
import Script from "next/script";
import { useSearchParams } from "next/navigation";
import { signInWithGoogleOneTap } from "@/lib/actions/sign-in";
import { getRedirectUrl } from "@/utils";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { getSession } from "./verify/[email]/[token]/actions";

const AUTH_GOOGLE_ID = process.env.NEXT_PUBLIC_AUTH_GOOGLE_ID;
declare global {
  interface Window {
    google: {
      accounts: {
        id: {
          initialize: (config: any) => void;
          prompt: (callback: (notification: any) => void) => void;
          cancel: () => void;
          revoke: (hint: string, callback: () => void) => void;
          disableAutoSelect: () => void;
        };
      };
    };
  }
}

export default function GoogleOneTap() {
  const { data: session } = useSession();
  const searchParams = useSearchParams();

  const [isGoogleScriptLoaded, setIsGoogleScriptLoaded] = useState(false);

  const handleCredentialResponse = useCallback(async (response: any) => {
    try {
      await signInWithGoogleOneTap(response?.credential as string);
      const newSession = await getSession();
      if (newSession) {
        const invitations = await getMyInvitations();
        window.location.href = getRedirectUrl(
          newSession,
          invitations.data || [],
        );
      }
    } catch (error) {
      console.error("Error signing in:", error);
    }
  }, []);

  const initializeGoogleOneTap = useCallback(() => {
    if (window?.google && !session) {
      try {
        window.google.accounts.id.initialize({
          client_id: AUTH_GOOGLE_ID,
          callback: handleCredentialResponse,
          context: "signin",
          ux_mode: "popup",
          auto_select: true,
          use_fedcm_for_prompt: true,
        });

        // disable auto select if signedOut is true
        if (searchParams.get("signedOut")) {
          window.google.accounts.id.disableAutoSelect();
        }

        window.google.accounts.id.prompt((notification: any) => {
          if (notification?.isNotDisplayed()) {
            console.log(
              "One Tap was not displayed:",
              notification?.getNotDisplayedReason(),
            );
          } else if (notification?.isSkippedMoment()) {
            console.log(
              "One Tap was skipped:",
              notification?.getSkippedReason(),
            );
          } else if (notification?.isDismissedMoment()) {
            console.log(
              "One Tap was dismissed:",
              notification?.getDismissedReason(),
            );
          }
        });
      } catch (error) {
        if (
          error instanceof Error &&
          error.message.includes(
            "Only one navigator.credentials.get request may be outstanding at one time",
          )
        ) {
          console.log(
            "FedCM request already in progress. Waiting before retrying...",
          );
          setTimeout(initializeGoogleOneTap, 1000);
        } else {
          console.error("Error initializing Google One Tap:", error);
        }
      }
    }
  }, [session, handleCredentialResponse]);

  useEffect(() => {
    if (isGoogleScriptLoaded) {
      initializeGoogleOneTap();
    }
  }, [isGoogleScriptLoaded, initializeGoogleOneTap]);

  useEffect(() => {
    if (session) {
      // If user is signed in, cancel any ongoing One Tap prompts
      window.google?.accounts.id.cancel();
    }
  }, [session]);

  return (
    <Script
      async
      defer
      onLoad={() => {
        setIsGoogleScriptLoaded(true);
      }}
      src="https://accounts.google.com/gsi/client"
      strategy="afterInteractive"
    />
  );
}
