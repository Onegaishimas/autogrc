import { useQuery, useMutation } from '@tanstack/react-query';
import {
  queryAuditEvents,
  getAuditEvent,
  getAuditStats,
  exportAuditEvents,
} from '../api/auditApi';
import type { AuditFilters } from '../types';

// Query keys
export const auditKeys = {
  all: ['audit'] as const,
  lists: () => [...auditKeys.all, 'list'] as const,
  list: (filters: AuditFilters) => [...auditKeys.lists(), filters] as const,
  details: () => [...auditKeys.all, 'detail'] as const,
  detail: (id: string) => [...auditKeys.details(), id] as const,
  stats: () => [...auditKeys.all, 'stats'] as const,
};

/**
 * Hook to query audit events with filters and pagination.
 */
export function useAuditEvents(filters: AuditFilters) {
  return useQuery({
    queryKey: auditKeys.list(filters),
    queryFn: () => queryAuditEvents(filters),
    placeholderData: (previousData) => previousData,
  });
}

/**
 * Hook to get a single audit event.
 */
export function useAuditEvent(id: string | null) {
  return useQuery({
    queryKey: auditKeys.detail(id || ''),
    queryFn: () => getAuditEvent(id!),
    enabled: !!id,
  });
}

/**
 * Hook to get audit statistics.
 */
export function useAuditStats() {
  return useQuery({
    queryKey: auditKeys.stats(),
    queryFn: getAuditStats,
  });
}

/**
 * Hook to export audit events.
 */
export function useExportAudit() {
  return useMutation({
    mutationFn: async (filters: AuditFilters) => {
      const blob = await exportAuditEvents(filters);
      // Create download link
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `audit_export_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    },
  });
}
