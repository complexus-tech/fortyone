"use client";

import { Box, Tabs, Text } from "ui";
import { useUserRole } from "@/hooks/role";
import { SecurityPolicies } from "./policy-panel";
import { SecuritySessions } from "./sessions-panel";
import { SecurityAudit } from "./audit-panel";
import { SecuritySingleSignOn } from "./sso-panel";
import { SecurityProvisioning } from "./scim-panel";

export const WorkspaceSecuritySettings = ({
  initialTab = "policies",
}: {
  initialTab?: string;
}) => {
  const { userRole } = useUserRole();
  if (userRole !== "admin")
    return <Text color="muted">Workspace admins manage security.</Text>;
  return (
    <Box>
      <Text as="h1" className="mb-2 text-2xl font-medium">
        Workspace security
      </Text>
      <Text className="max-w-3xl" color="muted">
        Manage access, sessions, and administrative activity.
      </Text>
      <Tabs className="mt-6" defaultValue={initialTab}>
        <Tabs.List className="mx-0 mb-5 flex-nowrap overflow-x-auto md:mx-0">
          <Tabs.Tab value="policies">Access policies</Tabs.Tab>
          <Tabs.Tab value="sso">Single sign-on</Tabs.Tab>
          <Tabs.Tab value="provisioning">Provisioning</Tabs.Tab>
          <Tabs.Tab value="sessions">Member sessions</Tabs.Tab>
          <Tabs.Tab value="audit">Audit log</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="policies">
          <SecurityPolicies />
        </Tabs.Panel>
        <Tabs.Panel value="sso">
          <SecuritySingleSignOn />
        </Tabs.Panel>
        <Tabs.Panel value="provisioning">
          <SecurityProvisioning />
        </Tabs.Panel>
        <Tabs.Panel value="sessions">
          <SecuritySessions />
        </Tabs.Panel>
        <Tabs.Panel value="audit">
          <SecurityAudit />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
