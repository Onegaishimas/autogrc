import { api } from '../../../lib/api';
import type {
  AuditEvent,
  AuditFilters,
  AuditQueryResponse,
  AuditStats,
} from '../types';

const AUDIT_BASE_URL = '/api/v1/audit';

/**
 * Build query string from filters
 */
function buildQueryParams(filters: AuditFilters): URLSearchParams {
  const params = new URLSearchParams();

  if (filters.event_types?.length) {
    params.set('event_types', filters.event_types.join(','));
  }
  if (filters.entity_types?.length) {
    params.set('entity_types', filters.entity_types.join(','));
  }
  if (filters.entity_id) {
    params.set('entity_id', filters.entity_id);
  }
  if (filters.status) {
    params.set('status', filters.status);
  }
  if (filters.start_date) {
    params.set('start_date', filters.start_date);
  }
  if (filters.end_date) {
    params.set('end_date', filters.end_date);
  }
  if (filters.search) {
    params.set('search', filters.search);
  }
  if (filters.page) {
    params.set('page', String(filters.page));
  }
  if (filters.page_size) {
    params.set('page_size', String(filters.page_size));
  }

  return params;
}

/**
 * Query audit events with filters and pagination.
 */
export async function queryAuditEvents(
  filters: AuditFilters
): Promise<AuditQueryResponse> {
  const params = buildQueryParams(filters);
  const response = await api.get<AuditQueryResponse>(
    `${AUDIT_BASE_URL}?${params.toString()}`
  );
  return response.data;
}

/**
 * Get a single audit event by ID.
 */
export async function getAuditEvent(id: string): Promise<AuditEvent> {
  const response = await api.get<AuditEvent>(`${AUDIT_BASE_URL}/${id}`);
  return response.data;
}

/**
 * Get audit statistics.
 */
export async function getAuditStats(): Promise<AuditStats> {
  const response = await api.get<AuditStats>(`${AUDIT_BASE_URL}/stats`);
  return response.data;
}

/**
 * Export audit events as CSV.
 */
export async function exportAuditEvents(filters: AuditFilters): Promise<Blob> {
  const params = buildQueryParams(filters);
  const response = await api.get(`${AUDIT_BASE_URL}/export?${params.toString()}`, {
    responseType: 'blob',
  });
  return response.data;
}
