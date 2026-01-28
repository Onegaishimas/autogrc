import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { getPullStatus } from '../api/pullApi';
import { isPullJobActive } from '../types';
import { pullKeys } from './usePull';
import { systemsKeys } from '../../systems/hooks/useSystems';

const POLL_INTERVAL = 2000; // 2 seconds

/**
 * Hook to poll pull job status while the job is active.
 * Automatically stops polling when job completes or fails.
 * Invalidates systems list when job completes successfully.
 */
export function usePullStatus(jobId: string | null) {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: pullKeys.job(jobId ?? ''),
    queryFn: () => getPullStatus(jobId!),
    enabled: !!jobId,
    refetchInterval: (query) => {
      // Only poll while job is active
      const data = query.state.data;
      if (data && isPullJobActive(data.job.status)) {
        return POLL_INTERVAL;
      }
      return false;
    },
  });

  // Invalidate systems list when job completes
  useEffect(() => {
    if (query.data?.job.status === 'completed') {
      queryClient.invalidateQueries({ queryKey: systemsKeys.lists() });
    }
  }, [query.data?.job.status, queryClient]);

  return query;
}
