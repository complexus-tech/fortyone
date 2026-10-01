"use client";

import { useState } from "react";
import { Box, Button, Flex, Input, Switch, Text } from "ui";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { SecurityPolicy } from "./types";
import { useSaveSecurityPolicy, useSecurityPolicy } from "./hooks";

const PolicyForm = ({ policy }: { policy: SecurityPolicy }) => {
  const [domains, setDomains] = useState(policy.allowedDomains.join(", "));
  const [allowGuests, setAllowGuests] = useState(policy.allowGuests);
  const [age, setAge] = useState(String(policy.maxSessionAgeHours));
  const save = useSaveSecurityPolicy();
  const domainList = Array.from(
    new Set(domains.split(/[,\s]+/).filter(Boolean)),
  );
  const valid =
    /^\d+$/.test(age) && Number(age) <= 720 && domainList.length <= 50;
  const changed =
    domains !== policy.allowedDomains.join(", ") ||
    allowGuests !== policy.allowGuests ||
    Number(age) !== policy.maxSessionAgeHours;
  return (
    <form
      className="space-y-6 px-6 pt-5 pb-6"
      onSubmit={(event) => {
        event.preventDefault();
        if (valid)
          save.mutate({
            allowedDomains: domainList,
            allowGuests,
            maxSessionAgeHours: Number(age),
            expectedVersion: policy.version,
          });
      }}
    >
      <Box className="max-w-2xl">
        <Input
          disabled={save.isPending}
          label="Allowed email domains"
          maxLength={13_000}
          onChange={(event) => {
            setDomains(event.target.value);
          }}
          placeholder="example.com, partner.com"
          value={domains}
        />
        <Text className="mt-2" color="muted">
          Leave blank to allow any existing member. Include your own email
          domain.
        </Text>
      </Box>
      <Flex align="center" className="max-w-2xl gap-5" justify="between">
        <Box>
          <label className="font-medium" htmlFor="security-allow-guests">
            Allow guest access
          </label>
          <Text className="mt-1" color="muted" id="security-guests-description">
            Guests retain their existing team permissions when allowed.
          </Text>
        </Box>
        <Switch
          aria-describedby="security-guests-description"
          aria-label="Allow guest access"
          checked={allowGuests}
          disabled={save.isPending}
          id="security-allow-guests"
          onCheckedChange={setAllowGuests}
        />
      </Flex>
      <Box className="max-w-2xl">
        <Input
          className="max-w-48"
          disabled={save.isPending}
          label="Require a new login after (hours)"
          max={720}
          min={0}
          onChange={(event) => {
            setAge(event.target.value);
          }}
          step={1}
          type="number"
          value={age}
        />
        <Text className="mt-2" color="muted">
          Set 0 for the normal session lifetime. Renewing a session does not
          reset login age.
        </Text>
      </Box>
      {save.isError ? (
        <Text color="danger" role="alert">
          {save.error.message}
        </Text>
      ) : null}
      {!valid ? (
        <Text color="danger" role="alert">
          Enter up to 50 domains and a whole number of hours between 0 and 720.
        </Text>
      ) : null}
      <Button
        disabled={!valid || !changed || save.isPending}
        loading={save.isPending}
        type="submit"
      >
        {save.isPending ? "Saving policy…" : "Save policy"}
      </Button>
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
        <Text aria-live="polite" className="px-6 pt-5 pb-6" color="muted">
          Loading access policies…
        </Text>
      ) : null}
      {query.isError ? (
        <Box className="px-6 pt-5 pb-6">
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
