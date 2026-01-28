import { useState, useCallback } from 'react';
import { SystemSelector } from '../../systems/components/SystemSelector';
import { PullProgress } from './PullProgress';
import { PullResults } from './PullResults';
import { useDiscoverSystems, useImportSystems } from '../../systems/hooks/useSystems';
import { useStartPull, useCancelPull } from '../hooks/usePull';
import { usePullStatus } from '../hooks/usePullStatus';
import { isPullJobActive } from '../types';

type WizardStep = 'select' | 'confirm' | 'progress' | 'results';

interface PullWizardProps {
  onComplete: () => void;
}

export function PullWizard({ onComplete }: PullWizardProps) {
  const [step, setStep] = useState<WizardStep>('select');
  const [selectedSystemIds, setSelectedSystemIds] = useState<string[]>([]);
  const [currentJobId, setCurrentJobId] = useState<string | null>(null);

  // Queries and mutations
  const { data: discoverData, isLoading: isDiscovering, refetch: refetchSystems } = useDiscoverSystems();
  const importMutation = useImportSystems();
  const startPullMutation = useStartPull();
  const cancelPullMutation = useCancelPull();
  const { data: jobStatusData } = usePullStatus(currentJobId);

  const currentJob = jobStatusData?.job;

  // When job completes or fails, move to results
  if (currentJob && !isPullJobActive(currentJob.status) && step === 'progress') {
    setStep('results');
  }

  const handleSelectionChange = useCallback((ids: string[]) => {
    setSelectedSystemIds(ids);
  }, []);

  const handleProceedToConfirm = useCallback(() => {
    if (selectedSystemIds.length === 0) return;
    setStep('confirm');
  }, [selectedSystemIds]);

  const handleBackToSelect = useCallback(() => {
    setStep('select');
  }, []);

  const handleStartPull = useCallback(async () => {
    try {
      // First, import the selected systems
      const importResult = await importMutation.mutateAsync({
        sn_sys_ids: selectedSystemIds,
      });

      // Then start the pull job for the imported systems
      const systemIds = importResult.imported.map((s) => s.id);
      const pullResult = await startPullMutation.mutateAsync({
        system_ids: systemIds,
      });

      setCurrentJobId(pullResult.job.id);
      setStep('progress');
    } catch (error) {
      console.error('Failed to start pull:', error);
      // Stay on confirm step to show error
    }
  }, [selectedSystemIds, importMutation, startPullMutation]);

  const handleCancelPull = useCallback(() => {
    if (currentJobId) {
      cancelPullMutation.mutate(currentJobId);
    }
  }, [currentJobId, cancelPullMutation]);

  const handleDone = useCallback(() => {
    onComplete();
  }, [onComplete]);

  const handleRetry = useCallback(() => {
    setCurrentJobId(null);
    setStep('select');
    refetchSystems();
  }, [refetchSystems]);

  // Get selected system names for confirmation
  const selectedSystems = discoverData?.systems.filter((s) =>
    selectedSystemIds.includes(s.sn_sys_id)
  ) || [];

  return (
    <div className="pull-wizard">
      {/* Step indicator */}
      <div className="wizard-steps">
        <div className={`wizard-step ${step === 'select' ? 'active' : ''} ${['confirm', 'progress', 'results'].includes(step) ? 'completed' : ''}`}>
          <span className="step-number">1</span>
          <span className="step-label">Select Systems</span>
        </div>
        <div className={`wizard-step ${step === 'confirm' ? 'active' : ''} ${['progress', 'results'].includes(step) ? 'completed' : ''}`}>
          <span className="step-number">2</span>
          <span className="step-label">Confirm</span>
        </div>
        <div className={`wizard-step ${step === 'progress' ? 'active' : ''} ${step === 'results' ? 'completed' : ''}`}>
          <span className="step-number">3</span>
          <span className="step-label">Pull</span>
        </div>
        <div className={`wizard-step ${step === 'results' ? 'active' : ''}`}>
          <span className="step-number">4</span>
          <span className="step-label">Results</span>
        </div>
      </div>

      {/* Step content */}
      <div className="wizard-content">
        {step === 'select' && (
          <div className="wizard-step-select">
            <h2>Select Systems to Import</h2>
            <p>
              Choose the systems you want to import from ServiceNow. Controls and
              implementation statements will be pulled for each selected system.
            </p>

            <SystemSelector
              systems={discoverData?.systems || []}
              onSelectionChange={handleSelectionChange}
              isLoading={isDiscovering}
              hideImported={false}
            />

            <div className="wizard-actions">
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleProceedToConfirm}
                disabled={selectedSystemIds.length === 0}
              >
                Continue ({selectedSystemIds.length} selected)
              </button>
            </div>
          </div>
        )}

        {step === 'confirm' && (
          <div className="wizard-step-confirm">
            <h2>Confirm Import</h2>
            <p>
              You are about to import the following systems and pull their
              controls and implementation statements from ServiceNow:
            </p>

            <ul className="confirm-list">
              {selectedSystems.map((system) => (
                <li key={system.sn_sys_id}>
                  <strong>{system.name}</strong>
                  {system.description && <span> - {system.description}</span>}
                </li>
              ))}
            </ul>

            {importMutation.isError && (
              <div className="alert alert-error">
                Failed to import systems. Please try again.
              </div>
            )}

            {startPullMutation.isError && (
              <div className="alert alert-error">
                Failed to start pull operation. Please try again.
              </div>
            )}

            <div className="wizard-actions">
              <button
                type="button"
                className="btn btn-secondary"
                onClick={handleBackToSelect}
                disabled={importMutation.isPending || startPullMutation.isPending}
              >
                Back
              </button>
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleStartPull}
                disabled={importMutation.isPending || startPullMutation.isPending}
              >
                {importMutation.isPending || startPullMutation.isPending
                  ? 'Starting...'
                  : 'Start Import'}
              </button>
            </div>
          </div>
        )}

        {step === 'progress' && currentJob && (
          <div className="wizard-step-progress">
            <h2>Importing Data</h2>
            <PullProgress job={currentJob} onCancel={handleCancelPull} />
          </div>
        )}

        {step === 'results' && currentJob && (
          <div className="wizard-step-results">
            <PullResults
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

export default PullWizard;
