import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  // Default route redirects to logs
  index("routes/home.tsx"),

  // Main sections
  route("logs", "routes/logs.tsx"),
  route("mocks", "routes/mocks.tsx"),
  route("hosts", "routes/hosts.tsx"),
] satisfies RouteConfig;
