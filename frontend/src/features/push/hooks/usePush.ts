import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { startPush, getPushStatus, cancelPush } from '../api/pushApi';
import { isPushJobActive } from '../types';
import type { StartPushRequest } from '../types';
import { statementKeys } from '../../statements/hooks/useStatements';

// Query keys
export const pushKeys = {
  all: ['push'] as const,
  jobs: () => [...pushKeys.all, 'job'] as const,
  job: (id: string) => [...pushKeys.jobs(), id] as const,
};

/**
 * Hook to start a push job.
 */
export function useStartPush() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: StartPushRequest) => startPush(request),
    onSuccess: () => {
      // Invalidate statement queries since push marks statements as synced
      queryClient.invalidateQueries({ queryKey: statementKeys.all });
    },
  });
}

/**
 * Hook to poll push job status.
 */
export function usePushStatus(jobId: string | null) {
  return useQuery({
    queryKey: pushKeys.job(jobId || ''),
    queryFn: () => getPushStatus(jobId!),
    enabled: !!jobId,
    refetchInterval: (query) => {
      // Poll every 1 second while job is active
      const data = query.state.data;
      if (data && isPushJobActive(data.job.status)) {
        return 1000;
      }
      return false;
    },
  });
}

/**
 * Hook to cancel a push job.
 */
export function useCancelPush() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (jobId: string) => cancelPush(jobId),
    onSuccess: (_, jobId) => {
      queryClient.invalidateQueries({ queryKey: pushKeys.job(jobId) });
    },
  });
}
