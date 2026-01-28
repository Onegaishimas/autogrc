import { api } from '../../../lib/api';
import type {
  StartPushRequest,
  StartPushResponse,
  PushStatusResponse,
} from '../types';

const PUSH_BASE_URL = '/api/v1/push';

/**
 * Start a new push operation for the specified statements.
 * The push runs as a background job - poll status to track progress.
 */
export async function startPush(
  request: StartPushRequest
): Promise<StartPushResponse> {
  const response = await api.post<StartPushResponse>(PUSH_BASE_URL, request);
  return response.data;
}

/**
 * Get the current status of a push job.
 */
export async function getPushStatus(jobId: string): Promise<PushStatusResponse> {
  const response = await api.get<PushStatusResponse>(`${PUSH_BASE_URL}/${jobId}`);
  return response.data;
}

/**
 * Cancel an active push job.
 */
export async function cancelPush(jobId: string): Promise<void> {
  await api.delete(`${PUSH_BASE_URL}/${jobId}`);
}
