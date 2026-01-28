import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  discoverSystems,
  listSystems,
  importSystems,
  deleteSystem,
} from '../api/systemsApi';
import type { ListSystemsParams, ImportSystemsRequest } from '../types';

// Query keys for cache management
export const systemsKeys = {
  all: ['systems'] as const,
  discover: () => [...systemsKeys.all, 'discover'] as const,
  lists: () => [...systemsKeys.all, 'list'] as const,
  list: (params?: ListSystemsParams) => [...systemsKeys.lists(), params] as const,
  detail: (id: string) => [...systemsKeys.all, 'detail', id] as const,
};

/**
 * Hook to discover systems available in ServiceNow.
 * Includes is_imported flag for each system.
 */
export function useDiscoverSystems() {
  return useQuery({
    queryKey: systemsKeys.discover(),
    queryFn: discoverSystems,
    staleTime: 30 * 1000, // 30 seconds - allow re-fetching after import
  });
}

/**
 * Hook to list locally imported systems with pagination.
 */
export function useListSystems(params?: ListSystemsParams) {
  return useQuery({
    queryKey: systemsKeys.list(params),
    queryFn: () => listSystems(params),
    staleTime: 60 * 1000, // 1 minute
    placeholderData: (previousData) => previousData, // Keep previous data while fetching
  });
}

/**
 * Hook to import systems from ServiceNow.
 * Invalidates discover and list caches on success.
 */
export function useImportSystems() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: ImportSystemsRequest) => importSystems(request),
    onSuccess: () => {
      // Invalidate both discover (to update is_imported flags) and list
      queryClient.invalidateQueries({ queryKey: systemsKeys.discover() });
      queryClient.invalidateQueries({ queryKey: systemsKeys.lists() });
    },
  });
}

/**
 * Hook to delete a locally imported system.
 * Invalidates discover and list caches on success.
 */
export function useDeleteSystem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteSystem(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: systemsKeys.discover() });
      queryClient.invalidateQueries({ queryKey: systemsKeys.lists() });
    },
  });
}
