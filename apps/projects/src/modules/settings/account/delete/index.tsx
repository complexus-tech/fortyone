"use client";

import { Box, Text, Button } from "ui";
import { useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { ConfirmDialog } from "@/components/ui";
import { clearAllStorage } from "@/components/shared/sidebar/utils";
import { useProfile } from "@/lib/hooks/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { MOBILE_ACCOUNT_DELETION_PATH } from "@/lib/mobile-auth";
import { withCallbackUrl } from "@/utils/callback-url";
import { finalizeDeletionCleanup } from "@/lib/finalize-deletion-cleanup";
import { SectionHeader } from "../../components";
import { deleteAccount } from "./action";

export const DeleteAccountSettings = ({
  account,
  mobileAuth = false,
}: {
  account?: { id: string; email: string };
  mobileAuth?: boolean;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string>();
  const [confirmedUserId, setConfirmedUserId] = useState<string>();
  const deleting = useRef(false);
  const queryClient = useQueryClient();
  const { data: profile } = useProfile();
  const currentAccount = account ?? profile;
  const router = useRouter();
  const workspaces = useQuery({
    queryKey: ["account-deletion", "workspaces", currentAccount?.id],
    queryFn: () => getWorkspaces(),
    enabled: Boolean(currentAccount?.id),
    staleTime: 0,
    refetchOnMount: "always",
    refetchOnWindowFocus: "always",
  });
  const refreshWorkspaces = workspaces.refetch;
  useEffect(() => {
    const refresh = () => {
      if (document.visibilityState !== "visible" || deleting.current) return;
      void refreshWorkspaces();
      if (mobileAuth) router.refresh();
    };
    window.addEventListener("pageshow", refresh);
    document.addEventListener("visibilitychange", refresh);
    return () => {
      window.removeEventListener("pageshow", refresh);
      document.removeEventListener("visibilitychange", refresh);
    };
  }, [mobileAuth, refreshWorkspaces, router]);
  const adminWorkspaces = (workspaces.data ?? []).filter(
    (workspace) => workspace.userRole === "admin" && !workspace.deletedAt,
  );

  const handleDelete = async () => {
    if (deleting.current || !confirmedUserId) return;
    deleting.current = true;
    setIsDeleting(true);
    setError(undefined);
    let result;
    try {
      result = await deleteAccount(confirmedUserId);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Could not delete your account. Please try again.",
      );
      deleting.current = false;
      setIsDeleting(false);
      return;
    }
    // A storage cleanup failure cannot turn a confirmed server deletion into
    // a retryable deletion. Leave the authenticated UI in every success case.
    await finalizeDeletionCleanup(
      async () => {
        await Promise.allSettled([queryClient.cancelQueries()]);
        queryClient.clear();
        clearAllStorage();
      },
      () => {
        window.location.replace(
          (result === "cleanup_pending"
            ? "/?accountDeleted=true&cleanupPending=true"
            : "/?accountDeleted=true") + (mobileAuth ? "&mobileApp=true" : ""),
        );
      },
    );
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Delete Account
      </Text>
      {currentAccount?.email ? (
        <Text className="mb-6" color="muted">
          Signed in as <strong>{currentAccount.email}</strong>. This is the
          account you will delete.
        </Text>
      ) : null}

      <Box className="border-border bg-surface mb-6 rounded-2xl border p-6">
        <Text as="h2" className="mb-2 text-lg" fontWeight="medium">
          Workspace ownership
        </Text>
        <Text className="mb-4" color="muted">
          If you are a workspace’s only administrator, give another member
          administrator access or delete the workspace first. You can delete a
          workspace you use alone without inviting anyone. After making changes,
          return here to finish deleting your account.
        </Text>
        {workspaces.isPending ? (
          <Text color="muted">Checking workspaces…</Text>
        ) : null}
        {workspaces.isError ? (
          <Text color="danger" role="alert">
            Could not load your workspaces. Refresh to try again.
          </Text>
        ) : null}
        {!workspaces.isPending &&
        !workspaces.isError &&
        adminWorkspaces.length === 0 ? (
          <Text color="muted">
            You have no active workspaces to administer. You can continue with
            account deletion below.
          </Text>
        ) : null}
        {adminWorkspaces.length > 0 ? (
          <ul className="divide-border divide-y">
            {adminWorkspaces.map((workspace) => (
              <li className="py-4" key={workspace.id}>
                <Text className="mb-2" fontWeight="medium">
                  {workspace.name}
                </Text>
                <div className="flex flex-wrap gap-3">
                  <Button
                    color="tertiary"
                    href={withCallbackUrl(
                      `/${encodeURIComponent(workspace.slug)}/settings/workspace/members`,
                      MOBILE_ACCOUNT_DELETION_PATH,
                    )}
                    size="sm"
                    variant="outline"
                  >
                    Manage administrators
                  </Button>
                  <Button
                    color="danger"
                    href={`${withCallbackUrl(`/${encodeURIComponent(workspace.slug)}/settings`, MOBILE_ACCOUNT_DELETION_PATH)}#delete-workspace`}
                    size="sm"
                    variant="outline"
                  >
                    Delete workspace
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        ) : null}
        <Button
          className="mt-4"
          color="tertiary"
          loading={workspaces.isFetching}
          onClick={() => {
            void refreshWorkspaces();
            if (mobileAuth) router.refresh();
          }}
          size="sm"
          variant="outline"
        >
          Refresh workspaces
        </Button>
      </Box>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          description="Once you delete your account, there is no going back. Please be certain."
          title="Delete your account"
        />
        <Box className="p-6">
          <Text className="mb-4" color="muted">
            Delete your profile, private documents, and connected credentials,
            and leave your teams and workspaces. Shared workspace tasks,
            comments, documents, feedback, and history remain for your
            teammates, attributed to Former user. Personal information you typed
            into shared content may remain. Limited security records are
            retained.
          </Text>
          <Button
            color="danger"
            disabled={!currentAccount?.id}
            onClick={() => {
              setConfirmedUserId(currentAccount?.id);
              setError(undefined);
              setIsOpen(true);
            }}
          >
            Delete Account
          </Button>
        </Box>
      </Box>

      <ConfirmDialog
        confirmPhrase="DELETE"
        confirmText="Delete Account"
        description="This cannot be undone. Your account, profile, private documents, and credentials will be deleted. Shared workspace tasks, comments, documents, feedback, and history remain attributed to Former user. Personal information within retained content may remain. Applications and integrations owned by your account will disconnect for workspaces using them. Connected-service cleanup may finish in the background. If you are a workspace’s only administrator, transfer that role or delete the workspace first."
        errorMessage={error}
        isLoading={isDeleting}
        isOpen={isOpen}
        loadingText="Deleting account…"
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleDelete}
        title="Delete my account"
      />
    </Box>
  );
};
