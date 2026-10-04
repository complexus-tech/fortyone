"use client";

import { useState } from "react";
import { Box, Button, Checkbox, Switch, Text } from "ui";
import { toast } from "sonner";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { SecurityPolicy } from "./types";
import { useSaveSecurityPolicy, useSecurityPolicy } from "./hooks";
import { SecuritySettingRow } from "./setting-row";
import { SecurityInput } from "./security-input";

const PolicyForm = ({ policy }: { policy: SecurityPolicy }) => {
  const [domains, setDomains] = useState(policy.allowedDomains.join(", "));
  const [allowGuests, setAllowGuests] = useState(policy.allowGuests);
  const [limitSessionAge, setLimitSessionAge] = useState(
    policy.maxSessionAgeHours > 0,
  );
  const [age, setAge] = useState(String(policy.maxSessionAgeHours || 24));
  const save = useSaveSecurityPolicy();
  const domainList = Array.from(
    new Set(domains.split(/[,\s]+/).filter(Boolean)),
  );
  const valid =
    domainList.length <= 50 &&
    (!limitSessionAge ||
      (/^\d+$/.test(age) && Number(age) >= 1 && Number(age) <= 720));
  const maxSessionAgeHours = limitSessionAge ? Number(age) : 0;
  const changed =
    domains !== policy.allowedDomains.join(", ") ||
    allowGuests !== policy.allowGuests ||
    maxSessionAgeHours !== policy.maxSessionAgeHours;
  return (
    <form
      className="divide-border divide-y-[0.5px]"
      onSubmit={(event) => {
        event.preventDefault();
        if (valid)
          save.mutate(
            {
              allowedDomains: domainList,
              allowGuests,
              maxSessionAgeHours,
              expectedVersion: policy.version,
            },
            {
              onSuccess: () => {
                toast.success("Access policies updated");
              },
            },
          );
      }}
    >
      <SecuritySettingRow
        description="Leave blank to allow any existing member. Include your own email domain."
        htmlFor="security-allowed-domains"
        title="Allowed email domains"
      >
        <Box className="w-full md:w-64 md:shrink-0">
          <SecurityInput
            aria-describedby="security-allowed-domains-description"
            disabled={save.isPending}
            id="security-allowed-domains"
            maxLength={13_000}
            onChange={(event) => {
              setDomains(event.target.value);
            }}
            placeholder="example.com, partner.com"
            value={domains}
          />
        </Box>
      </SecuritySettingRow>
      <SecuritySettingRow
        description="Guests retain their existing team permissions when allowed."
        htmlFor="security-allow-guests"
        layout="inline"
        title="Allow guest access"
      >
        <Switch
          aria-describedby="security-allow-guests-description"
          aria-label="Allow guest access"
          checked={allowGuests}
          disabled={save.isPending}
          id="security-allow-guests"
          onCheckedChange={setAllowGuests}
        />
      </SecuritySettingRow>
      <SecuritySettingRow
        description="When off, use the normal session lifetime. Renewing a session does not reset login age."
        htmlFor="security-session-age-enabled"
        layout="inline"
        title="Require a new login after a set time"
      >
        <Checkbox
          aria-describedby="security-session-age-enabled-description"
          checked={limitSessionAge}
          disabled={save.isPending}
          id="security-session-age-enabled"
          onCheckedChange={(checked) => {
            setLimitSessionAge(checked === true);
          }}
        />
      </SecuritySettingRow>
      {limitSessionAge ? (
        <SecuritySettingRow
          description="Choose a whole number from 1 to 720."
          htmlFor="security-session-age"
          title="Hours"
        >
          <Box className="w-full md:w-32 md:shrink-0">
            <SecurityInput
              aria-describedby="security-session-age-description"
              disabled={save.isPending}
              id="security-session-age"
              max={720}
              min={1}
              onChange={(event) => {
                setAge(event.target.value);
              }}
              step={1}
              type="number"
              value={age}
            />
          </Box>
        </SecuritySettingRow>
      ) : null}
      <Box className="space-y-3 px-6 py-4">
        {save.isError ? (
          <Text color="danger" role="alert">
            {save.error.message}
          </Text>
        ) : null}
        {!valid ? (
          <Text color="danger" role="alert">
            Enter up to 50 domains and, when enabled, a whole number of hours
            between 1 and 720.
          </Text>
        ) : null}
        <Button
          disabled={!valid || !changed || save.isPending}
          loading={save.isPending}
          loadingText="Saving policy…"
          type="submit"
        >
          Save policy
        </Button>
      </Box>
    </form>
  );
};

export const SecurityPolicies = () => {
  const query = useSecurityPolicy();
  return (
    <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
      <SectionHeader
        description="Control member access to this workspace."
        title="Access policies"
      />
      {query.isPending ? (
        <Text aria-live="polite" className="px-6 py-4" color="muted">
          Loading access policies…
        </Text>
      ) : null}
      {query.isError ? (
        <Box className="px-6 py-4">
          <Text color="danger" role="alert">
            Access policies could not be loaded.
          </Text>
          <Button
            className="mt-3"
            color="tertiary"
            onClick={() => void query.refetch()}
            variant="outline"
          >
            Try again
          </Button>
        </Box>
      ) : null}
      {query.data ? (
        <PolicyForm key={query.data.version} policy={query.data} />
      ) : null}
    </Box>
  );
};
