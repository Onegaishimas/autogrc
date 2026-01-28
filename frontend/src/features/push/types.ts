// Push job status
export type PushJobStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled';

// Statement result from push operation
export interface StatementResult {
  statement_id: string;
  success: boolean;
  error?: string;
  pushed_at?: string;
}

// Push job
export interface PushJob {
  id: string;
  status: PushJobStatus;
  total_count: number;
  completed: number;
  succeeded: number;
  failed: number;
  results: StatementResult[];
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

// Start push request
export interface StartPushRequest {
  statement_ids: string[];
}

// Start push response
export interface StartPushResponse {
  job: PushJob;
}

// Push status response
export interface PushStatusResponse {
  job: PushJob;
}

// Helper to check if job is active
export function isPushJobActive(status: PushJobStatus): boolean {
  return status === 'pending' || status === 'running';
}

// Get status display info
export function getPushStatusInfo(status: PushJobStatus): {
  label: string;
  color: string;
  bgColor: string;
} {
  switch (status) {
    case 'pending':
      return { label: 'Pending', color: '#6b7280', bgColor: '#f3f4f6' };
    case 'running':
      return { label: 'Running', color: '#1e40af', bgColor: '#dbeafe' };
    case 'completed':
      return { label: 'Completed', color: '#065f46', bgColor: '#d1fae5' };
    case 'failed':
      return { label: 'Failed', color: '#dc2626', bgColor: '#fef2f2' };
    case 'cancelled':
      return { label: 'Cancelled', color: '#92400e', bgColor: '#fef3c7' };
    default:
      return { label: status, color: '#6b7280', bgColor: '#f3f4f6' };
  }
}
