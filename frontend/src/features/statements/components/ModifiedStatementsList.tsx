import { useState } from 'react';
import type { Statement } from '../types';
import { useModifiedStatements, useConflictStatements } from '../hooks/useStatements';
import { StatementCard } from './StatementCard';
import { StatementEditor } from './StatementEditor';
import { ConflictResolver } from './ConflictResolver';

export function ModifiedStatementsList() {
  const [editingStatement, setEditingStatement] = useState<Statement | null>(null);
  const [resolvingStatement, setResolvingStatement] = useState<Statement | null>(null);
  const [activeTab, setActiveTab] = useState<'modified' | 'conflicts'>('modified');

  const {
    data: modifiedData,
    isLoading: modifiedLoading,
    refetch: refetchModified,
  } = useModifiedStatements();

  const {
    data: conflictData,
    isLoading: conflictLoading,
    refetch: refetchConflicts,
  } = useConflictStatements();

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
    refetchModified();
    refetchConflicts();
    setEditingStatement(null);
  };

  const handleResolved = () => {
    refetchModified();
    refetchConflicts();
    setResolvingStatement(null);
  };

  const modifiedStatements = modifiedData?.statements || [];
  const conflictStatements = conflictData?.statements || [];
  const modifiedCount = modifiedData?.count || 0;
  const conflictCount = conflictData?.count || 0;

  return (
    <div className="modified-statements-list">
      {/* Tabs */}
      <div className="modified-statements-tabs">
        <button
          type="button"
          className={`tab-btn ${activeTab === 'modified' ? 'active' : ''}`}
          onClick={() => setActiveTab('modified')}
        >
          Modified ({modifiedCount})
        </button>
        <button
          type="button"
          className={`tab-btn ${activeTab === 'conflicts' ? 'active' : ''} ${conflictCount > 0 ? 'has-conflicts' : ''}`}
          onClick={() => setActiveTab('conflicts')}
        >
          Conflicts ({conflictCount})
        </button>
      </div>

      {/* Content */}
      <div className="modified-statements-content">
        {activeTab === 'modified' && (
          <>
            {modifiedLoading ? (
              <div className="loading">Loading modified statements...</div>
            ) : modifiedStatements.length === 0 ? (
              <div className="empty-state">
                <p>No modified statements.</p>
                <p className="text-muted">
                  Statements you edit will appear here for review before pushing
                  to ServiceNow.
                </p>
              </div>
            ) : (
              <div className="statement-list-grid">
                {modifiedStatements.map((statement) => (
                  <StatementCard
                    key={statement.id}
                    statement={statement}
                    onEdit={handleEdit}
                    onResolve={
                      statement.sync_status === 'conflict'
                        ? handleResolve
                        : undefined
                    }
                  />
                ))}
              </div>
            )}
          </>
        )}

        {activeTab === 'conflicts' && (
          <>
            {conflictLoading ? (
              <div className="loading">Loading conflict statements...</div>
            ) : conflictStatements.length === 0 ? (
              <div className="empty-state">
                <p>No conflicts detected.</p>
                <p className="text-muted">
                  Conflicts occur when a statement is modified both locally and
                  in ServiceNow.
                </p>
              </div>
            ) : (
              <div className="statement-list-grid">
                {conflictStatements.map((statement) => (
                  <StatementCard
                    key={statement.id}
                    statement={statement}
                    onEdit={handleEdit}
                    onResolve={handleResolve}
                  />
                ))}
              </div>
            )}
          </>
        )}
      </div>

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

export default ModifiedStatementsList;
