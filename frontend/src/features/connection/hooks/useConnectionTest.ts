import { useMutation, useQueryClient } from '@tanstack/react-query';
import { testConnection } from '../api/connectionApi';
import { connectionKeys } from './useConnection';
import type { TestResponse } from '../types';

/**
 * Hook to test ServiceNow connection
 * Invalidates status cache on success to update last test info
 */
export function useTestConnection() {
  const queryClient = useQueryClient();

  return useMutation<TestResponse, Error>({
    mutationFn: testConnection,
    onSuccess: () => {
      // Invalidate status to get updated last_test_at and last_test_status
      queryClient.invalidateQueries({ queryKey: connectionKeys.status() });
    },
  });
}
