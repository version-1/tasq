import { QueryClient } from "@tanstack/react-query";

export function createAppQueryClient(): QueryClient {
  return new QueryClient();
}

export const appQueryClient = createAppQueryClient();
