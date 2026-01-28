import { api } from '../../../lib/api';
import type {
  DiscoverSystemsResponse,
  ListSystemsResponse,
  ListSystemsParams,
  ImportSystemsRequest,
  ImportSystemsResponse,
} from '../types';

const SYSTEMS_BASE_URL = '/api/v1/sync/systems';

/**
 * Discover systems available in ServiceNow.
 * Returns systems with is_imported flag indicating if already in local database.
 */
export async function discoverSystems(): Promise<DiscoverSystemsResponse> {
  const response = await api.get<DiscoverSystemsResponse>(`${SYSTEMS_BASE_URL}/discover`);
  return response.data;
}

/**
 * List locally imported systems with pagination.
 */
export async function listSystems(params?: ListSystemsParams): Promise<ListSystemsResponse> {
  const response = await api.get<ListSystemsResponse>(SYSTEMS_BASE_URL, { params });
  return response.data;
}

/**
 * Import selected systems from ServiceNow into local database.
 */
export async function importSystems(request: ImportSystemsRequest): Promise<ImportSystemsResponse> {
  const response = await api.post<ImportSystemsResponse>(`${SYSTEMS_BASE_URL}/import`, request);
  return response.data;
}

/**
 * Delete a locally imported system and all its associated data.
 */
export async function deleteSystem(id: string): Promise<void> {
  await api.delete(`${SYSTEMS_BASE_URL}/${id}`);
}
