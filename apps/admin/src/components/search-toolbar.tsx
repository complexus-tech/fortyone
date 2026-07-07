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
    <form>
      <Flex align="center" className="flex-wrap gap-2">
        <Box className="w-full min-w-0 md:w-80 md:shrink-0">
          <Input
            className="md:pl-10"
            defaultValue={defaultQuery}
            leftIcon={<SearchIcon className="h-4" />}
            name="q"
            placeholder={placeholder}
          />
        </Box>
        {statusOptions ? (
          <label className="block w-full md:w-42 md:shrink-0">
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
        <Button color="tertiary" type="submit">
          Search
        </Button>
        {defaultQuery || defaultStatus ? (
          <Button
            color="tertiary"
            href={statusOptions ? "/workspaces" : "/users"}
            variant="naked"
          >
            Clear
          </Button>
        ) : null}
      </Flex>
    </form>
  );
};
