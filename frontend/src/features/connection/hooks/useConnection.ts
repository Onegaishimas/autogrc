import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getStatus, saveConfig, deleteConnection } from '../api/connectionApi';
import type { ConfigRequest, StatusResponse, ConfigResponse } from '../types';

// Query keys for cache management
export const connectionKeys = {
  all: ['connection'] as const,
  status: () => [...connectionKeys.all, 'status'] as const,
};

/**
 * Hook to fetch connection status
 * Refreshes every 60 seconds when window is focused
 */
export function useConnectionStatus() {
  return useQuery<StatusResponse>({
    queryKey: connectionKeys.status(),
    queryFn: getStatus,
    staleTime: 60 * 1000, // 1 minute
    refetchOnWindowFocus: true,
  });
}

/**
 * Hook to save connection configuration
 * Invalidates status cache on success
 */
export function useSaveConnection() {
  const queryClient = useQueryClient();

  return useMutation<ConfigResponse, Error, ConfigRequest>({
    mutationFn: saveConfig,
    onSuccess: () => {
      // Invalidate status query to refetch latest data
      queryClient.invalidateQueries({ queryKey: connectionKeys.status() });
    },
  });
}

/**
 * Hook to delete connection
 * Invalidates status cache on success
 */
export function useDeleteConnection() {
  const queryClient = useQueryClient();

  return useMutation<{ message: string }, Error>({
    mutationFn: deleteConnection,
    onSuccess: () => {
      // Invalidate status query to refetch latest data
      queryClient.invalidateQueries({ queryKey: connectionKeys.status() });
    },
  });
}
