import type { NextConfig } from "next";
import nextra from "nextra";

const withNextra = nextra({
  search: {
    codeblocks: false,
  },
});

const nextConfig: NextConfig = {
  /* config options here */
};

export default withNextra(nextConfig);
