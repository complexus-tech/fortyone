"use server";

import { headers } from "next/headers";

export const getHeaders = async () => {
  const headersList = await headers();
  return headersList;
};
