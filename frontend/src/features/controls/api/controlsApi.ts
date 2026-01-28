import { api } from '../../../lib/api';
import type { PolicyStatementsResponse, PolicyStatementsParams, PolicyStatement } from '../types';

const CONTROLS_BASE = '/v1/controls';

export async function getPolicyStatements(
  params?: PolicyStatementsParams
): Promise<PolicyStatementsResponse> {
  const response = await api.get<PolicyStatementsResponse>(
    `${CONTROLS_BASE}/policy-statements`,
    { params }
  );
  return response.data;
}

export async function getPolicyStatement(id: string): Promise<PolicyStatement> {
  const response = await api.get<PolicyStatement>(
    `${CONTROLS_BASE}/policy-statements/${id}`
  );
  return response.data;
}
