import { api } from '../../../lib/api';
import type {
  Statement,
  ListStatementsResponse,
  ListStatementsParams,
  ModifiedStatementsResponse,
  ConflictStatementsResponse,
  UpdateStatementRequest,
  ResolveConflictRequest,
} from '../types';

const STATEMENTS_BASE_URL = '/api/v1/statements';

/**
 * List statements by control with pagination.
 */
export async function listStatements(
  params: ListStatementsParams
): Promise<ListStatementsResponse> {
  const response = await api.get<ListStatementsResponse>(STATEMENTS_BASE_URL, {
    params,
  });
  return response.data;
}

/**
 * Get a single statement by ID.
 */
export async function getStatement(id: string): Promise<Statement> {
  const response = await api.get<Statement>(`${STATEMENTS_BASE_URL}/${id}`);
  return response.data;
}

/**
 * Update local content for a statement.
 */
export async function updateStatement(
  id: string,
  request: UpdateStatementRequest
): Promise<Statement> {
  const response = await api.put<Statement>(
    `${STATEMENTS_BASE_URL}/${id}`,
    request
  );
  return response.data;
}

/**
 * List all modified statements.
 */
export async function listModifiedStatements(): Promise<ModifiedStatementsResponse> {
  const response = await api.get<ModifiedStatementsResponse>(
    `${STATEMENTS_BASE_URL}/modified`
  );
  return response.data;
}

/**
 * List all statements with conflicts.
 */
export async function listConflictStatements(): Promise<ConflictStatementsResponse> {
  const response = await api.get<ConflictStatementsResponse>(
    `${STATEMENTS_BASE_URL}/conflicts`
  );
  return response.data;
}

/**
 * Resolve a sync conflict for a statement.
 */
export async function resolveConflict(
  id: string,
  request: ResolveConflictRequest
): Promise<Statement> {
  const response = await api.post<Statement>(
    `${STATEMENTS_BASE_URL}/${id}/resolve`,
    request
  );
  return response.data;
}

/**
 * Revert a statement to its remote content.
 */
export async function revertStatement(id: string): Promise<Statement> {
  const response = await api.post<Statement>(
    `${STATEMENTS_BASE_URL}/${id}/revert`
  );
  return response.data;
}
