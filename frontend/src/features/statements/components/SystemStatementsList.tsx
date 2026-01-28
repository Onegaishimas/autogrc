import { useState } from 'react';
import type { Statement, SyncStatus } from '../types';
import { useStatements } from '../hooks/useStatements';
import { StatementCard } from './StatementCard';
import { StatementEditor } from './StatementEditor';
import { ConflictResolver } from './ConflictResolver';

interface SystemStatementsListProps {
  systemId: string;
  systemName: string;
  onBack: () => void;
  pageSize?: number;
}

export function SystemStatementsList({
  systemId,
  systemName,
  onBack,
  pageSize = 20,
}: SystemStatementsListProps) {
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<SyncStatus | ''>('');
  const [search, setSearch] = useState('');
  const [editingStatement, setEditingStatement] = useState<Statement | null>(null);
  const [resolvingStatement, setResolvingStatement] = useState<Statement | null>(null);

  const { data, isLoading, error, refetch } = useStatements({
    system_id: systemId,
    page,
    page_size: pageSize,
    sync_status: statusFilter || undefined,
    search: search || undefined,
  });

  const handleEdit = (statement: Statement) => {
    setEditingStatement(statement);
  };

  const handleResolve = (statement: Statement) => {
    setResolvingStatement(statement);
  };

  const handleEditorClose = () => {
    setEditingStatement(null);
  };

  const handleResolverClose = () => {
    setResolvingStatement(null);
  };

  const handleSaved = () => {
    refetch();
    setEditingStatement(null);
  };

  const handleResolved = () => {
    refetch();
    setResolvingStatement(null);
  };

  const statements = data?.statements || [];
  const totalPages = data?.total_pages || 0;
  const totalCount = data?.total_count || 0;

  return (
    <div className="system-statements-list">
      <div className="system-statements-header">
        <button type="button" className="btn btn-secondary" onClick={onBack}>
          ← Back to Systems
        </button>
        <div className="system-statements-title">
          <h2>{systemName}</h2>
          <p>Manage implementation statements for this system</p>
        </div>
      </div>

      {error ? (
        <div className="statement-list error">
          <div className="alert alert-error">
            Failed to load statements. Please try again.
          </div>
          <button className="btn btn-primary" onClick={() => refetch()}>
            Retry
          </button>
        </div>
      ) : (
        <>
          {/* Filters */}
          <div className="statement-list-filters">
            <div className="statement-list-search">
              <input
                type="text"
                placeholder="Search statements..."
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
              />
            </div>

            <div className="statement-list-status-filter">
              <select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value as SyncStatus | '');
                  setPage(1);
                }}
              >
                <option value="">All statuses</option>
                <option value="synced">Synced</option>
                <option value="modified">Modified</option>
                <option value="conflict">Conflict</option>
                <option value="new">New</option>
              </select>
            </div>

            <div className="statement-list-count">
              {isLoading ? 'Loading...' : `${totalCount} statement${totalCount !== 1 ? 's' : ''}`}
            </div>
          </div>

          {/* Statement list */}
          {isLoading ? (
            <div className="statement-list loading">Loading statements...</div>
          ) : statements.length === 0 ? (
            <div className="statement-list-empty">
              <p>No statements found for this system.</p>
              <p className="text-muted">
                Import controls from ServiceNow to see statements here.
              </p>
            </div>
          ) : (
            <div className="statement-list-grid">
              {statements.map((statement) => (
                <StatementCard
                  key={statement.id}
                  statement={statement}
                  onEdit={handleEdit}
                  onResolve={
                    statement.sync_status === 'conflict' ? handleResolve : undefined
                  }
                />
              ))}
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="statement-list-pagination">
              <button
                className="btn btn-secondary btn-sm"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
              >
                Previous
              </button>
              <span className="statement-list-page-info">
                Page {page} of {totalPages}
              </span>
              <button
                className="btn btn-secondary btn-sm"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}

      {/* Editor modal */}
      {editingStatement && (
        <div className="modal-overlay" onClick={handleEditorClose}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <StatementEditor
              statement={editingStatement}
              onClose={handleEditorClose}
              onSaved={handleSaved}
            />
          </div>
        </div>
      )}

      {/* Conflict resolver modal */}
      {resolvingStatement && (
        <div className="modal-overlay" onClick={handleResolverClose}>
          <div
            className="modal-content modal-content-wide"
            onClick={(e) => e.stopPropagation()}
          >
            <ConflictResolver
              statement={resolvingStatement}
              onClose={handleResolverClose}
              onResolved={handleResolved}
            />
          </div>
        </div>
      )}
    </div>
  );
}

export default SystemStatementsList;
