import type { Config } from "@react-router/dev/config";

export default {
  ssr: false,
  basename: process.env.NODE_ENV === 'production' ? '/ui' : '/',
} satisfies Config;
