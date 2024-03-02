import MillionCompiler from "@million/lint";

// module.exports = {
//   reactStrictMode: true,
//   transpilePackages: ["ui", "icons"],
// };

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["ui", "icons"],
};

export default MillionCompiler.next()(nextConfig);
