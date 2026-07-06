import { SearchIcon } from "icons";
import { Box, Button, Flex, Input, Text } from "ui";

type StatusOption = {
  label: string;
  value: string;
};

export const SearchToolbar = ({
  defaultQuery,
  defaultStatus,
  placeholder,
  statusOptions,
}: {
  defaultQuery?: string;
  defaultStatus?: string;
  placeholder: string;
  statusOptions?: StatusOption[];
}) => {
  return (
    <Box className="border-border bg-surface rounded-lg border-[0.5px] p-3">
      <form>
        <Flex align="end" className="gap-2">
          <Box className="min-w-0 flex-1">
            <Input
              defaultValue={defaultQuery}
              leftIcon={<SearchIcon className="h-4" />}
              name="q"
              placeholder={placeholder}
              rounded="lg"
            />
          </Box>
          {statusOptions ? (
            <label className="block w-42 shrink-0">
              <Text className="sr-only">Status</Text>
              <select
                className="border-input bg-surface focus-visible:ring-ring h-[2.8rem] w-full rounded-lg border px-3 outline-none focus-visible:ring-2"
                defaultValue={defaultStatus ?? ""}
                name="status"
              >
                {statusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          ) : null}
          <Button color="tertiary" rounded="lg" type="submit">
            Search
          </Button>
          {defaultQuery || defaultStatus ? (
            <Button
              color="tertiary"
              href={statusOptions ? "/workspaces" : "/users"}
              rounded="lg"
              variant="naked"
            >
              Clear
            </Button>
          ) : null}
        </Flex>
      </form>
    </Box>
  );
};
