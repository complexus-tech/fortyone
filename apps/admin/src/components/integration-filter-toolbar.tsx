"use client";

import Link from "next/link";
import { FilterIcon, SearchIcon } from "icons";
import { Box, Button, Flex, Input, Menu } from "ui";

const providerOptions = [
  { label: "All providers", value: "" },
  { label: "Slack", value: "slack" },
  { label: "GitHub", value: "github" },
  { label: "Calendar", value: "calendar" },
  { label: "Figma", value: "figma" },
];

const statusOptions = [
  { label: "Any connection state", value: "" },
  { label: "Connected", value: "connected" },
  { label: "Needs attention", value: "attention" },
  { label: "Not connected", value: "not_connected" },
];

const optionLabel = (
  options: { label: string; value: string }[],
  value?: string,
) =>
  options.find((option) => option.value === value)?.label ?? options[0].label;

const filterHref = ({
  provider,
  query,
  status,
}: {
  provider?: string;
  query?: string;
  status?: string;
}) => {
  const params = new URLSearchParams();
  if (query) params.set("q", query);
  if (provider) params.set("provider", provider);
  if (status) params.set("status", status);
  const search = params.toString();
  return search ? `/integrations?${search}` : "/integrations";
};

export const IntegrationFilterToolbar = ({
  defaultProvider,
  defaultQuery,
  defaultStatus,
}: {
  defaultProvider?: string;
  defaultQuery?: string;
  defaultStatus?: string;
}) => (
  <form action="/integrations">
    <Flex align="center" className="flex-wrap gap-2">
      <Box className="w-full min-w-0 md:w-[28rem] md:shrink-0">
        <Input
          className="md:pl-10"
          defaultValue={defaultQuery}
          leftIcon={<SearchIcon className="h-4" />}
          name="q"
          placeholder="Search by workspace or slug"
        />
        {defaultProvider ? (
          <input name="provider" type="hidden" value={defaultProvider} />
        ) : null}
        {defaultStatus ? (
          <input name="status" type="hidden" value={defaultStatus} />
        ) : null}
      </Box>

      <Menu>
        <Menu.Button>
          <Button
            active={Boolean(defaultProvider)}
            color="tertiary"
            type="button"
          >
            <FilterIcon className="h-4" />
            {optionLabel(providerOptions, defaultProvider)}
          </Button>
        </Menu.Button>
        <Menu.Items align="start" className="w-52">
          <Menu.Group>
            {providerOptions.map((option) => (
              <Menu.Item
                active={(defaultProvider ?? "") === option.value}
                asChild
                key={option.value}
              >
                <Link
                  href={filterHref({
                    provider: option.value,
                    query: defaultQuery,
                    status: defaultStatus,
                  })}
                >
                  {option.label}
                </Link>
              </Menu.Item>
            ))}
          </Menu.Group>
        </Menu.Items>
      </Menu>

      <Menu>
        <Menu.Button>
          <Button
            active={Boolean(defaultStatus)}
            color="tertiary"
            type="button"
          >
            <FilterIcon className="h-4" />
            {optionLabel(statusOptions, defaultStatus)}
          </Button>
        </Menu.Button>
        <Menu.Items align="start" className="w-56">
          <Menu.Group>
            {statusOptions.map((option) => (
              <Menu.Item
                active={(defaultStatus ?? "") === option.value}
                asChild
                key={option.value}
              >
                <Link
                  href={filterHref({
                    provider: defaultProvider,
                    query: defaultQuery,
                    status: option.value,
                  })}
                >
                  {option.label}
                </Link>
              </Menu.Item>
            ))}
          </Menu.Group>
        </Menu.Items>
      </Menu>

      {defaultQuery || defaultProvider || defaultStatus ? (
        <Button color="tertiary" href="/integrations" variant="naked">
          Clear
        </Button>
      ) : null}
    </Flex>
  </form>
);
