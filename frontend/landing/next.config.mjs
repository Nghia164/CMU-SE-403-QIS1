/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Cho phép biên dịch component của MASM Design System nằm ngoài thư mục app (../design-system)
  experimental: { externalDir: true },
};

export default nextConfig;
