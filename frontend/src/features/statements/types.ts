// Sync status for statements
export type SyncStatus = 'synced' | 'modified' | 'conflict' | 'new';

// Conflict resolution options
export type ConflictResolution = 'keep_local' | 'keep_remote' | 'merge';

// Statement represents a control implementation statement
export interface Statement {
  id: string;
  control_id: string;
  sn_sys_id: string;
  statement_type: string;

  // Content
  remote_content?: string;
  remote_updated_at?: string;
  local_content?: string;
  is_modified: boolean;
  modified_at?: string;

  // Sync status
  sync_status: SyncStatus;
  conflict_resolved_at?: string;

  // Computed field for display
  effective_content: string;

  // Timestamps
  last_pull_at?: string;
  last_push_at?: string;
  created_at: string;
  updated_at: string;
}

// List statements response
export interface ListStatementsResponse {
  statements: Statement[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// List statements params
export interface ListStatementsParams {
  control_id?: string;
  system_id?: string;  // Filter by system (alternative to control_id)
  page?: number;
  page_size?: number;
  sync_status?: SyncStatus;
  search?: string;
}

// Modified statements response
export interface ModifiedStatementsResponse {
  statements: Statement[];
  count: number;
}

// Conflict statements response
export interface ConflictStatementsResponse {
  statements: Statement[];
  count: number;
}

// Update statement request
export interface UpdateStatementRequest {
  local_content: string;
}

// Resolve conflict request
export interface ResolveConflictRequest {
  resolution: ConflictResolution;
  merged_content?: string;
}

// Helper to get sync status display info
export function getSyncStatusInfo(status: SyncStatus): {
  label: string;
  color: string;
  bgColor: string;
} {
  switch (status) {
    case 'synced':
      return { label: 'Synced', color: '#065f46', bgColor: '#d1fae5' };
    case 'modified':
      return { label: 'Modified', color: '#92400e', bgColor: '#fef3c7' };
    case 'conflict':
      return { label: 'Conflict', color: '#dc2626', bgColor: '#fef2f2' };
    case 'new':
      return { label: 'New', color: '#1e40af', bgColor: '#dbeafe' };
    default:
      return { label: status, color: '#6b7280', bgColor: '#f3f4f6' };
  }
}
