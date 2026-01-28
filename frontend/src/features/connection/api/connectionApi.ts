import { api } from '../../../lib/api';
import type {
  StatusResponse,
  ConfigRequest,
  ConfigResponse,
  TestResponse,
} from '../types';

const CONNECTION_BASE = '/v1/connection';

/**
 * Get current connection status
 */
export async function getStatus(): Promise<StatusResponse> {
  const response = await api.get<StatusResponse>(`${CONNECTION_BASE}/status`);
  return response.data;
}

/**
 * Save connection configuration
 */
export async function saveConfig(config: ConfigRequest): Promise<ConfigResponse> {
  const response = await api.post<ConfigResponse>(`${CONNECTION_BASE}/config`, config);
  return response.data;
}

/**
 * Test current connection
 */
export async function testConnection(): Promise<TestResponse> {
  const response = await api.post<TestResponse>(`${CONNECTION_BASE}/test`);
  return response.data;
}

/**
 * Delete current connection
 */
export async function deleteConnection(): Promise<{ message: string }> {
  const response = await api.delete<{ message: string }>(CONNECTION_BASE);
  return response.data;
}
