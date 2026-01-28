// Types for pull operations

export type PullJobStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface PullProgress {
  total_systems: number;
  completed_systems: number;
  total_controls: number;
  completed_controls: number;
  total_statements: number;
  completed_statements: number;
  current_system?: string;
  errors?: string[];
}

export interface PullJob {
  id: string;
  system_ids: string[];
  status: PullJobStatus;
  progress: PullProgress;
  started_at?: string;
  completed_at?: string;
  error?: string;
  created_at: string;
}

export interface StartPullRequest {
  system_ids: string[];
}

export interface StartPullResponse {
  job: PullJob;
}

export interface PullStatusResponse {
  job: PullJob;
}

// Helper to check if a pull job is still active
export function isPullJobActive(status: PullJobStatus): boolean {
  return status === 'pending' || status === 'running';
}

// Helper to calculate overall progress percentage
export function calculatePullProgress(progress: PullProgress): number {
  const total = progress.total_systems + progress.total_controls + progress.total_statements;
  const completed = progress.completed_systems + progress.completed_controls + progress.completed_statements;

  if (total === 0) return 0;
  return Math.round((completed / total) * 100);
}
