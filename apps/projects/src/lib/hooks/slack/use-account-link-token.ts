import { useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useLinkSlackAccount } from "./use-link-account";

export type SlackAccountLinkStatus =
  | "idle"
  | "linking"
  | "success"
  | "already_connected"
  | "error";

const slackLinkTokenParameter = "slack_link_token";

export const useSlackAccountLinkToken = () => {
  const searchParams = useSearchParams();
  const linkSlackAccount = useLinkSlackAccount();
  const attemptedTokenRef = useRef<string | null>(null);
  const [status, setStatus] = useState<SlackAccountLinkStatus>("idle");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const token = searchParams.get(slackLinkTokenParameter);

  const clearTokenFromURL = useCallback((linkedToken: string) => {
    const url = new URL(window.location.href);
    if (url.searchParams.get(slackLinkTokenParameter) !== linkedToken) {
      return;
    }
    url.searchParams.delete(slackLinkTokenParameter);
    window.history.replaceState({}, "", url.toString());
  }, []);

  const connect = useCallback(
    (linkToken: string) => {
      attemptedTokenRef.current = linkToken;
      setStatus("linking");
      setErrorMessage(null);
      linkSlackAccount.mutate(linkToken, {
        onSuccess: (result) => {
          if (result.error) {
            setStatus("error");
            setErrorMessage(result.error.message);
            return;
          }
          setStatus(
            result.data?.status === "already_connected"
              ? "already_connected"
              : "success",
          );
          clearTokenFromURL(linkToken);
        },
        onError: (error) => {
          setStatus("error");
          setErrorMessage(
            error instanceof Error
              ? error.message
              : "FortyOne could not connect this account.",
          );
        },
      });
    },
    [clearTokenFromURL, linkSlackAccount],
  );

  useEffect(() => {
    if (!token || attemptedTokenRef.current === token) {
      return;
    }
    connect(token);
  }, [connect, token]);

  return {
    errorMessage,
    hasToken: Boolean(token),
    retry: token
      ? () => {
          connect(token);
        }
      : undefined,
    status,
  };
};
