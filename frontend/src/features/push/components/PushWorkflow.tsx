import { useState, useCallback } from 'react';
import { useModifiedStatements } from '../../statements/hooks/useStatements';
import { useStartPush, usePushStatus, useCancelPush } from '../hooks/usePush';
import { isPushJobActive } from '../types';
import { PushProgress } from './PushProgress';
import { PushResults } from './PushResults';
import { getSyncStatusInfo } from '../../statements/types';
import type { Statement } from '../../statements/types';

type WorkflowStep = 'select' | 'confirm' | 'progress' | 'results';

interface PushWorkflowProps {
  onComplete: () => void;
}

export function PushWorkflow({ onComplete }: PushWorkflowProps) {
  const [step, setStep] = useState<WorkflowStep>('select');
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [currentJobId, setCurrentJobId] = useState<string | null>(null);

  // Queries and mutations
  const { data: modifiedData, isLoading, refetch } = useModifiedStatements();
  const startPushMutation = useStartPush();
  const cancelPushMutation = useCancelPush();
  const { data: jobStatusData } = usePushStatus(currentJobId);

  const currentJob = jobStatusData?.job;

  // When job completes, move to results
  if (currentJob && !isPushJobActive(currentJob.status) && step === 'progress') {
    setStep('results');
  }

  const modifiedStatements = modifiedData?.statements || [];
  // Filter out statements with conflicts
  const pushableStatements = modifiedStatements.filter(
    (s) => s.sync_status !== 'conflict'
  );
  const conflictStatements = modifiedStatements.filter(
    (s) => s.sync_status === 'conflict'
  );

  const handleToggleSelect = useCallback((id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  }, []);

  const handleSelectAll = useCallback(() => {
    setSelectedIds(pushableStatements.map((s) => s.id));
  }, [pushableStatements]);

  const handleDeselectAll = useCallback(() => {
    setSelectedIds([]);
  }, []);

  const handleProceedToConfirm = useCallback(() => {
    if (selectedIds.length === 0) return;
    setStep('confirm');
  }, [selectedIds]);

  const handleBackToSelect = useCallback(() => {
    setStep('select');
  }, []);

  const handleStartPush = useCallback(async () => {
    try {
      const result = await startPushMutation.mutateAsync({
        statement_ids: selectedIds,
      });
      setCurrentJobId(result.job.id);
      setStep('progress');
    } catch (error) {
      console.error('Failed to start push:', error);
    }
  }, [selectedIds, startPushMutation]);

  const handleCancelPush = useCallback(() => {
    if (currentJobId) {
      cancelPushMutation.mutate(currentJobId);
    }
  }, [currentJobId, cancelPushMutation]);

  const handleDone = useCallback(() => {
    onComplete();
  }, [onComplete]);

  const handleRetry = useCallback(() => {
    setCurrentJobId(null);
    setStep('select');
    refetch();
  }, [refetch]);

  const selectedStatements = pushableStatements.filter((s) =>
    selectedIds.includes(s.id)
  );

  return (
    <div className="push-workflow">
      {/* Step indicator */}
      <div className="wizard-steps">
        <div
          className={`wizard-step ${step === 'select' ? 'active' : ''} ${['confirm', 'progress', 'results'].includes(step) ? 'completed' : ''}`}
        >
          <span className="step-number">1</span>
          <span className="step-label">Select</span>
        </div>
        <div
          className={`wizard-step ${step === 'confirm' ? 'active' : ''} ${['progress', 'results'].includes(step) ? 'completed' : ''}`}
        >
          <span className="step-number">2</span>
          <span className="step-label">Confirm</span>
        </div>
        <div
          className={`wizard-step ${step === 'progress' ? 'active' : ''} ${step === 'results' ? 'completed' : ''}`}
        >
          <span className="step-number">3</span>
          <span className="step-label">Push</span>
        </div>
        <div className={`wizard-step ${step === 'results' ? 'active' : ''}`}>
          <span className="step-number">4</span>
          <span className="step-label">Results</span>
        </div>
      </div>

      {/* Step content */}
      <div className="wizard-content">
        {step === 'select' && (
          <div className="push-step-select">
            <h2>Select Statements to Push</h2>
            <p>
              Choose which modified statements to push back to ServiceNow.
              Statements with conflicts must be resolved first.
            </p>

            {isLoading ? (
              <div className="loading">Loading modified statements...</div>
            ) : pushableStatements.length === 0 && conflictStatements.length === 0 ? (
              <div className="empty-state">
                <p>No modified statements to push.</p>
                <p className="text-muted">
                  Edit statements in the Pending Changes view to create
                  modifications.
                </p>
              </div>
            ) : (
              <>
                {conflictStatements.length > 0 && (
                  <div className="alert alert-warning">
                    {conflictStatements.length} statement(s) have conflicts and
                    cannot be pushed. Resolve conflicts in the Pending Changes
                    view first.
                  </div>
                )}

                <div className="push-select-actions">
                  <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={handleSelectAll}
                    disabled={pushableStatements.length === 0}
                  >
                    Select All
                  </button>
                  <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={handleDeselectAll}
                    disabled={selectedIds.length === 0}
                  >
                    Deselect All
                  </button>
                  <span className="push-select-count">
                    {selectedIds.length} of {pushableStatements.length} selected
                  </span>
                </div>

                <div className="push-statement-list">
                  {pushableStatements.map((stmt) => (
                    <StatementSelectItem
                      key={stmt.id}
                      statement={stmt}
                      selected={selectedIds.includes(stmt.id)}
                      onToggle={() => handleToggleSelect(stmt.id)}
                    />
                  ))}
                </div>
              </>
            )}

            <div className="wizard-actions">
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleProceedToConfirm}
                disabled={selectedIds.length === 0}
              >
                Continue ({selectedIds.length} selected)
              </button>
            </div>
          </div>
        )}

        {step === 'confirm' && (
          <div className="push-step-confirm">
            <h2>Confirm Push</h2>
            <p>
              You are about to push {selectedStatements.length} statement(s) to
              ServiceNow. This will update the remote content with your local
              modifications.
            </p>

            <div className="confirm-list">
              {selectedStatements.slice(0, 10).map((stmt) => (
                <div key={stmt.id} className="confirm-item">
                  <span className="confirm-type">{stmt.statement_type}</span>
                  <span className="confirm-preview">
                    {stmt.effective_content?.slice(0, 100)}...
                  </span>
                </div>
              ))}
              {selectedStatements.length > 10 && (
                <p className="confirm-more">
                  ...and {selectedStatements.length - 10} more
                </p>
              )}
            </div>

            {startPushMutation.isError && (
              <div className="alert alert-error">
                Failed to start push. Please try again.
              </div>
            )}

            <div className="wizard-actions">
              <button
                type="button"
                className="btn btn-secondary"
                onClick={handleBackToSelect}
                disabled={startPushMutation.isPending}
              >
                Back
              </button>
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleStartPush}
                disabled={startPushMutation.isPending}
              >
                {startPushMutation.isPending ? 'Starting...' : 'Push to ServiceNow'}
              </button>
            </div>
          </div>
        )}

        {step === 'progress' && currentJob && (
          <div className="push-step-progress">
            <h2>Pushing to ServiceNow</h2>
            <PushProgress job={currentJob} onCancel={handleCancelPush} />
          </div>
        )}

        {step === 'results' && currentJob && (
          <div className="push-step-results">
            <PushResults
              job={currentJob}
              onDone={handleDone}
              onRetry={handleRetry}
            />
          </div>
        )}
      </div>
    </div>
  );
}

// Helper component for statement selection
interface StatementSelectItemProps {
  statement: Statement;
  selected: boolean;
  onToggle: () => void;
}

function StatementSelectItem({
  statement,
  selected,
  onToggle,
}: StatementSelectItemProps) {
  const statusInfo = getSyncStatusInfo(statement.sync_status);

  return (
    <div
      className={`push-statement-item ${selected ? 'selected' : ''}`}
      onClick={onToggle}
    >
      <input
        type="checkbox"
        checked={selected}
        onChange={onToggle}
        onClick={(e) => e.stopPropagation()}
      />
      <div className="push-statement-info">
        <div className="push-statement-header">
          <span className="push-statement-type">{statement.statement_type}</span>
          <span
            className="push-statement-badge"
            style={{
              color: statusInfo.color,
              backgroundColor: statusInfo.bgColor,
            }}
          >
            {statusInfo.label}
          </span>
        </div>
        <p className="push-statement-preview">
          {statement.effective_content?.slice(0, 150) || 'No content'}
          {(statement.effective_content?.length || 0) > 150 ? '...' : ''}
        </p>
        {statement.modified_at && (
          <span className="push-statement-meta">
            Modified: {new Date(statement.modified_at).toLocaleDateString()}
          </span>
        )}
      </div>
    </div>
  );
}

export default PushWorkflow;
