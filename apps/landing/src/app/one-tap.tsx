/* eslint-disable turbo/no-undeclared-env-vars -- Google One Tap is not typed */
/* eslint-disable no-console -- Google One Tap is not typed */
/* eslint-disable @typescript-eslint/no-explicit-any -- Google One Tap is not typed */
"use client";

import { useEffect, useCallback, useState } from "react";
import { signIn, useSession } from "next-auth/react";
import Script from "next/script";

const AUTH_GOOGLE_ID = process.env.AUTH_GOOGLE_ID;

declare global {
  interface Window {
    google: {
      accounts: {
        id: {
          initialize: (config: any) => void;
          prompt: (callback: (notification: any) => void) => void;
          cancel: () => void;
          revoke: (hint: string, callback: () => void) => void;
        };
      };
    };
  }
}

export default function GoogleOneTap() {
  const { data: session } = useSession();
  const [isGoogleScriptLoaded, setIsGoogleScriptLoaded] = useState(false);

  const handleCredentialResponse = useCallback((response: any) => {
    console.log("handleCredentialResponse", response);
    console.log("credential", response.credential);
    signIn("one-tap", {
      credential: { idToken: response.credential },
      redirectTo: "/auth-callback",
    }).catch((error) => {
      console.error("Error signing in:", error);
    });
  }, []);

  const initializeGoogleOneTap = useCallback(() => {
    if (window.google && !session) {
      try {
        window.google.accounts.id.initialize({
          client_id: AUTH_GOOGLE_ID,
          callback: handleCredentialResponse,
          context: "signin",
          ux_mode: "popup",
          auto_select: true,
          use_fedcm_for_prompt: true,
        });

        window.google.accounts.id.prompt((notification: any) => {
          if (notification.isNotDisplayed()) {
            console.log(
              "One Tap was not displayed:",
              notification.getNotDisplayedReason(),
            );
          } else if (notification.isSkippedMoment()) {
            console.log(
              "One Tap was skipped:",
              notification.getSkippedReason(),
            );
          } else if (notification.isDismissedMoment()) {
            console.log(
              "One Tap was dismissed:",
              notification.getDismissedReason(),
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
