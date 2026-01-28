import { useMutation, useQueryClient } from '@tanstack/react-query';
import { startPull, cancelPull } from '../api/pullApi';
import type { StartPullRequest } from '../types';
import { systemsKeys } from '../../systems/hooks/useSystems';

// Query keys for pull operations
export const pullKeys = {
  all: ['pull'] as const,
  jobs: () => [...pullKeys.all, 'jobs'] as const,
  job: (id: string) => [...pullKeys.all, 'job', id] as const,
};

/**
 * Hook to start a new pull operation.
 * Invalidates systems list on success (since controls/statements counts change).
 */
export function useStartPull() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: StartPullRequest) => startPull(request),
    onSuccess: () => {
      // Pull completion will update control/statement counts
      queryClient.invalidateQueries({ queryKey: systemsKeys.lists() });
    },
  });
}

/**
 * Hook to cancel an active pull job.
 */
export function useCancelPull() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (jobId: string) => cancelPull(jobId),
    onSuccess: (_, jobId) => {
      queryClient.invalidateQueries({ queryKey: pullKeys.job(jobId) });
    },
  });
}
