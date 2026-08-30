"use client";

import { ApiKeyIcon, ServiceAccountIcon, WebhookIcon } from "icons";
import { Box, Tabs, Text } from "ui";
import { useUserRole } from "@/hooks/role";
import { PersonalAccessTokens } from "./components/personal-access-tokens";
import { ServiceAccounts } from "./components/service-accounts";
import { WebhookEndpoints } from "./components/webhook-endpoints";

export const ApiSettings = () => {
  const { userRole } = useUserRole();
  const isAdmin = userRole === "admin";

  return (
    <Box>
      <Text as="h1" className="mb-2 text-2xl font-medium">
        API
      </Text>
      <Text className="max-w-3xl" color="muted">
        Connect scripts and external applications to FortyOne with scoped,
        expiring credentials and signed event delivery. Personal tokens belong
        to you and are limited to this workspace.
      </Text>

      <Tabs className="mt-6" defaultValue="tokens">
        <Tabs.List className="mx-0 mb-5 flex-nowrap overflow-x-auto md:mx-0">
          <Tabs.Tab leftIcon={<ApiKeyIcon />} value="tokens">
            Access tokens
          </Tabs.Tab>
          {isAdmin ? (
            <>
              <Tabs.Tab
                leftIcon={<ServiceAccountIcon />}
                value="service-accounts"
              >
                Service accounts
              </Tabs.Tab>
              <Tabs.Tab leftIcon={<WebhookIcon />} value="webhooks">
                Webhooks
              </Tabs.Tab>
            </>
          ) : null}
        </Tabs.List>
        <Tabs.Panel value="tokens">
          <PersonalAccessTokens />
        </Tabs.Panel>
        {isAdmin ? (
          <>
            <Tabs.Panel value="service-accounts">
              <ServiceAccounts />
            </Tabs.Panel>
            <Tabs.Panel value="webhooks">
              <WebhookEndpoints />
            </Tabs.Panel>
          </>
        ) : null}
      </Tabs>

      <Box className="border-border bg-surface-elevated mt-5 rounded-xl border p-4">
        <Text className="font-medium">Security defaults</Text>
        <Text className="mt-1 max-w-3xl" color="muted">
          Secrets are shown once, stored only as protected digests or encrypted
          envelopes, and can be revoked without changing your FortyOne password.
          Use the narrowest scopes and shortest practical lifetime.
        </Text>
      </Box>
    </Box>
  );
};
