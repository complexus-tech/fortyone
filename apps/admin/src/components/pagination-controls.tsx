import { Button, Flex, Text } from "ui";
import type { Pagination } from "@/lib/types";

const EMPTY_PARAMS: Record<string, string | number | undefined> = {};

const buildHref = (
  pathname: string,
  params: Record<string, string | number | undefined>,
  page: number,
) => {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === "") {
      return;
    }
    searchParams.set(key, String(value));
  });
  searchParams.set("page", String(page));

  return `${pathname}?${searchParams.toString()}`;
};

export const PaginationControls = ({
  pagination,
  pathname,
  params = EMPTY_PARAMS,
}: {
  pagination: Pagination;
  pathname: string;
  params?: Record<string, string | number | undefined>;
}) => {
  const totalPages = Math.max(
    1,
    Math.ceil(pagination.total / pagination.limit),
  );
  const previousPage = Math.max(1, pagination.page - 1);
  const nextPage = Math.min(totalPages, pagination.page + 1);

  return (
    <Flex
      align="center"
      className="border-border border-t-[0.5px] px-4 py-3"
      justify="between"
    >
      <Text className="text-[0.95rem]" color="muted">
        Showing {pagination.total === 0 ? 0 : pagination.offset + 1}-
        {Math.min(pagination.offset + pagination.limit, pagination.total)} of{" "}
        {pagination.total}
      </Text>
      <Flex align="center" className="gap-2">
        <Button
          color="tertiary"
          disabled={pagination.page <= 1}
          href={
            pagination.page <= 1
              ? undefined
              : buildHref(pathname, params, previousPage)
          }
          size="sm"
          variant="naked"
        >
          Previous
        </Button>
        <Text className="text-[0.95rem]" color="muted">
          Page {pagination.page} of {totalPages}
        </Text>
        <Button
          color="tertiary"
          disabled={pagination.page >= totalPages}
          href={
            pagination.page >= totalPages
              ? undefined
              : buildHref(pathname, params, nextPage)
          }
          size="sm"
          variant="naked"
        >
          Next
        </Button>
      </Flex>
    </Flex>
  );
};
