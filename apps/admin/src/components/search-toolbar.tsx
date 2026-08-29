"use client";

import Link from "next/link";
import { FilterIcon, SearchIcon } from "icons";
import { Box, Button, Flex, Input, Menu } from "ui";

type FilterOption = {
  label: string;
  value: string;
};

const buildHref = ({
  filterName,
  inputName,
  pathname,
  query,
  value,
}: {
  filterName: string;
  inputName: string;
  pathname: string;
  query?: string;
  value: string;
}) => {
  const params = new URLSearchParams();

  if (query) {
    params.set(inputName, query);
  }
  if (value) {
    params.set(filterName, value);
  }

  const search = params.toString();
  return search ? `${pathname}?${search}` : pathname;
};

export const SearchToolbar = ({
  defaultFilter,
  defaultQuery,
  filterName = "status",
  filterOptions,
  inputName = "q",
  pathname,
  placeholder,
}: {
  defaultFilter?: string;
  defaultQuery?: string;
  filterName?: string;
  filterOptions?: FilterOption[];
  inputName?: string;
  pathname: string;
  placeholder: string;
}) => {
  return (
    <form action={pathname}>
      <Flex align="center" className="flex-wrap gap-2">
        <Box className="w-full min-w-0 md:w-[28rem] md:shrink-0">
          <Input
            className="md:pl-10"
            defaultValue={defaultQuery}
            leftIcon={<SearchIcon className="h-4" />}
            name={inputName}
            placeholder={placeholder}
          />
          {defaultFilter ? (
            <input name={filterName} type="hidden" value={defaultFilter} />
          ) : null}
        </Box>
        {filterOptions ? (
          <Menu>
            <Menu.Button>
              <Button
                active={Boolean(defaultFilter)}
                color="tertiary"
                type="button"
              >
                <FilterIcon className="h-4" />
                Filters
              </Button>
            </Menu.Button>
            <Menu.Items align="start" className="w-52">
              <Menu.Group>
                {filterOptions.map((option) => (
                  <Menu.Item
                    active={(defaultFilter ?? "") === option.value}
                    asChild
                    key={option.value}
                  >
                    <Link
                      href={buildHref({
                        filterName,
                        inputName,
                        pathname,
                        query: defaultQuery,
                        value: option.value,
                      })}
                    >
                      {option.label}
                    </Link>
                  </Menu.Item>
                ))}
              </Menu.Group>
            </Menu.Items>
          </Menu>
        ) : null}
        {defaultQuery || defaultFilter ? (
          <Button color="tertiary" href={pathname} variant="naked">
            Clear
          </Button>
        ) : null}
      </Flex>
    </form>
  );
};
