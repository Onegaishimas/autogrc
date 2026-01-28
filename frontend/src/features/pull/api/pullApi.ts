import { api } from '../../../lib/api';
import type {
  StartPullRequest,
  StartPullResponse,
  PullStatusResponse,
} from '../types';

const PULL_BASE_URL = '/api/v1/sync/pull';

/**
 * Start a new pull operation for the specified systems.
 * The pull runs as a background job - poll status to track progress.
 */
export async function startPull(request: StartPullRequest): Promise<StartPullResponse> {
  const response = await api.post<StartPullResponse>(PULL_BASE_URL, request);
  return response.data;
}

/**
 * Get the current status of a pull job.
 */
export async function getPullStatus(jobId: string): Promise<PullStatusResponse> {
  const response = await api.get<PullStatusResponse>(`${PULL_BASE_URL}/${jobId}`);
  return response.data;
}

/**
 * Cancel an active pull job.
 */
export async function cancelPull(jobId: string): Promise<void> {
  await api.delete(`${PULL_BASE_URL}/${jobId}`);
}
