import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { getPolicyStatements, getPolicyStatement } from '../api/controlsApi';
import type { PolicyStatementsParams } from '../types';

// Query keys for cache management
export const controlsKeys = {
  all: ['controls'] as const,
  policyStatements: () => [...controlsKeys.all, 'policy-statements'] as const,
  policyStatementsList: (params?: PolicyStatementsParams) =>
    [...controlsKeys.policyStatements(), 'list', params] as const,
  policyStatement: (id: string) =>
    [...controlsKeys.policyStatements(), 'detail', id] as const,
};

export function usePolicyStatements(params?: PolicyStatementsParams) {
  return useQuery({
    queryKey: controlsKeys.policyStatementsList(params),
    queryFn: () => getPolicyStatements(params),
    placeholderData: keepPreviousData,
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function usePolicyStatement(id: string) {
  return useQuery({
    queryKey: controlsKeys.policyStatement(id),
    queryFn: () => getPolicyStatement(id),
    enabled: !!id,
    staleTime: 60 * 1000, // 1 minute
  });
}
