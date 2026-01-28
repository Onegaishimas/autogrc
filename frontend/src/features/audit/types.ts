// Event types
export type EventType =
  | 'pull'
  | 'push'
  | 'edit'
  | 'conflict_detected'
  | 'conflict_resolved'
  | 'connection_test'
  | 'connection_config'
  | 'system_import'
  | 'system_delete';

// Entity types
export type EntityType = 'system' | 'control' | 'statement' | 'connection';

// Audit event
export interface AuditEvent {
  id: string;
  event_type: EventType;
  entity_type: EntityType;
  entity_id: string;
  action: string;
  status: 'success' | 'failure';
  details?: Record<string, unknown>;
  user_email?: string;
  ip_address?: string;
  created_at: string;
}

// Query filters
export interface AuditFilters {
  event_types?: EventType[];
  entity_types?: EntityType[];
  entity_id?: string;
  status?: string;
  start_date?: string;
  end_date?: string;
  search?: string;
  page?: number;
  page_size?: number;
}

// Query response
export interface AuditQueryResponse {
  events: AuditEvent[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// Stats response
export interface AuditStats {
  total_events: number;
  events_by_type: Record<string, number>;
  events_by_status: Record<string, number>;
  events_today: number;
  events_this_week: number;
  events_this_month: number;
}

// Event type display info
export function getEventTypeInfo(type: EventType): {
  label: string;
  color: string;
  bgColor: string;
} {
  switch (type) {
    case 'pull':
      return { label: 'Pull', color: '#1e40af', bgColor: '#dbeafe' };
    case 'push':
      return { label: 'Push', color: '#065f46', bgColor: '#d1fae5' };
    case 'edit':
      return { label: 'Edit', color: '#92400e', bgColor: '#fef3c7' };
    case 'conflict_detected':
      return { label: 'Conflict', color: '#dc2626', bgColor: '#fef2f2' };
    case 'conflict_resolved':
      return { label: 'Resolved', color: '#059669', bgColor: '#d1fae5' };
    case 'connection_test':
      return { label: 'Test', color: '#6b7280', bgColor: '#f3f4f6' };
    case 'connection_config':
      return { label: 'Config', color: '#7c3aed', bgColor: '#ede9fe' };
    case 'system_import':
      return { label: 'Import', color: '#0891b2', bgColor: '#cffafe' };
    case 'system_delete':
      return { label: 'Delete', color: '#be123c', bgColor: '#fce7f3' };
    default:
      return { label: type, color: '#6b7280', bgColor: '#f3f4f6' };
  }
}

// Entity type display info
export function getEntityTypeInfo(type: EntityType): { label: string } {
  switch (type) {
    case 'system':
      return { label: 'System' };
    case 'control':
      return { label: 'Control' };
    case 'statement':
      return { label: 'Statement' };
    case 'connection':
      return { label: 'Connection' };
    default:
      return { label: type };
  }
}

// Status display info
export function getStatusInfo(status: string): {
  label: string;
  color: string;
  bgColor: string;
} {
  switch (status) {
    case 'success':
      return { label: 'Success', color: '#065f46', bgColor: '#d1fae5' };
    case 'failure':
      return { label: 'Failure', color: '#dc2626', bgColor: '#fef2f2' };
    default:
      return { label: status, color: '#6b7280', bgColor: '#f3f4f6' };
  }
}
