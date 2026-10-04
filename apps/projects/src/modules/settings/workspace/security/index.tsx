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
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Security
      </Text>
      <Tabs defaultValue={initialTab}>
        <Tabs.List className="mx-0 mb-4 max-w-full flex-nowrap overflow-x-auto md:mx-0 [&>button]:shrink-0">
          <Tabs.Tab value="policies">Access policies</Tabs.Tab>
          <Tabs.Tab value="sso">Single sign-on</Tabs.Tab>
          <Tabs.Tab value="provisioning">Provisioning</Tabs.Tab>
          <Tabs.Tab value="sessions">Member sessions</Tabs.Tab>
          <Tabs.Tab value="audit">Audit log</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel className="max-w-[48rem]" value="policies">
          <SecurityPolicies />
        </Tabs.Panel>
        <Tabs.Panel className="max-w-[48rem]" value="sso">
          <SecuritySingleSignOn />
        </Tabs.Panel>
        <Tabs.Panel className="max-w-[48rem]" value="provisioning">
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
