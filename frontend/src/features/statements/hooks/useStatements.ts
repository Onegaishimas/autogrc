import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  listStatements,
  getStatement,
  updateStatement,
  listModifiedStatements,
  listConflictStatements,
  resolveConflict,
  revertStatement,
} from '../api/statementsApi';
import type {
  ListStatementsParams,
  UpdateStatementRequest,
  ResolveConflictRequest,
} from '../types';

// Query keys
export const statementKeys = {
  all: ['statements'] as const,
  lists: () => [...statementKeys.all, 'list'] as const,
  list: (params: ListStatementsParams) =>
    [...statementKeys.lists(), params] as const,
  details: () => [...statementKeys.all, 'detail'] as const,
  detail: (id: string) => [...statementKeys.details(), id] as const,
  modified: () => [...statementKeys.all, 'modified'] as const,
  conflicts: () => [...statementKeys.all, 'conflicts'] as const,
};

/**
 * Hook to list statements for a control with pagination.
 */
export function useStatements(params: ListStatementsParams) {
  return useQuery({
    queryKey: statementKeys.list(params),
    queryFn: () => listStatements(params),
    placeholderData: (previousData) => previousData,
    enabled: !!params.control_id,
  });
}

/**
 * Hook to get a single statement.
 */
export function useStatement(id: string | null) {
  return useQuery({
    queryKey: statementKeys.detail(id || ''),
    queryFn: () => getStatement(id!),
    enabled: !!id,
  });
}

/**
 * Hook to list all modified statements.
 */
export function useModifiedStatements() {
  return useQuery({
    queryKey: statementKeys.modified(),
    queryFn: listModifiedStatements,
  });
}

/**
 * Hook to list all statements with conflicts.
 */
export function useConflictStatements() {
  return useQuery({
    queryKey: statementKeys.conflicts(),
    queryFn: listConflictStatements,
  });
}

/**
 * Hook to update a statement's local content.
 */
export function useUpdateStatement() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: string;
      request: UpdateStatementRequest;
    }) => updateStatement(id, request),
    onSuccess: (data) => {
      // Invalidate lists and update detail cache
      queryClient.invalidateQueries({ queryKey: statementKeys.lists() });
      queryClient.invalidateQueries({ queryKey: statementKeys.modified() });
      queryClient.setQueryData(statementKeys.detail(data.id), data);
    },
  });
}

/**
 * Hook to resolve a conflict.
 */
export function useResolveConflict() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: string;
      request: ResolveConflictRequest;
    }) => resolveConflict(id, request),
    onSuccess: (data) => {
      // Invalidate lists and update caches
      queryClient.invalidateQueries({ queryKey: statementKeys.lists() });
      queryClient.invalidateQueries({ queryKey: statementKeys.conflicts() });
      queryClient.invalidateQueries({ queryKey: statementKeys.modified() });
      queryClient.setQueryData(statementKeys.detail(data.id), data);
    },
  });
}

/**
 * Hook to revert a statement to remote content.
 */
export function useRevertStatement() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => revertStatement(id),
    onSuccess: (data) => {
      // Invalidate lists and update caches
      queryClient.invalidateQueries({ queryKey: statementKeys.lists() });
      queryClient.invalidateQueries({ queryKey: statementKeys.modified() });
      queryClient.setQueryData(statementKeys.detail(data.id), data);
    },
  });
}
