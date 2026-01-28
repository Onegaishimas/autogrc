import { useState, useCallback } from 'react';
import { SystemCard } from './SystemCard';
import { useListSystems, useDeleteSystem } from '../hooks/useSystems';
import type { LocalSystem } from '../types';

interface SystemsViewProps {
  onSelectSystem?: (system: { id: string; name: string }) => void;
}

export function SystemsView({ onSelectSystem }: SystemsViewProps) {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const { data, isLoading, error } = useListSystems();
  const deleteMutation = useDeleteSystem();

  const handleDelete = useCallback(async (system: LocalSystem) => {
    if (confirm(`Delete system "${system.name}"? This will remove all associated controls and statements.`)) {
      try {
        await deleteMutation.mutateAsync(system.id);
        setSelectedId(null);
      } catch (err) {
        console.error('Failed to delete system:', err);
      }
    }
  }, [deleteMutation]);

  if (isLoading) {
    return (
      <div className="systems-view">
        <div className="systems-header">
          <h1>Systems</h1>
          <p>Manage your imported systems and their controls.</p>
        </div>
        <div className="system-list loading">
          Loading systems...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="systems-view">
        <div className="systems-header">
          <h1>Systems</h1>
        </div>
        <div className="alert alert-error">
          Failed to load systems. Please check your connection settings.
        </div>
      </div>
    );
  }

  const systems = data?.systems || [];

  return (
    <div className="systems-view">
      <div className="systems-header">
        <h1>Systems</h1>
        <p>
          {systems.length === 0
            ? 'No systems imported yet. Use the Import tab to pull systems from ServiceNow.'
            : `${systems.length} system${systems.length !== 1 ? 's' : ''} imported.`}
        </p>
      </div>

      {systems.length === 0 ? (
        <div className="system-list empty">
          <p>No systems have been imported yet.</p>
          <p>Click the "Import" tab to discover and import systems from ServiceNow.</p>
        </div>
      ) : (
        <div className="system-list-grid">
          {systems.map((system) => (
            <div key={system.id} className="system-card-wrapper">
              <SystemCard
                system={system}
                isSelected={selectedId === system.id}
                onSelect={() => setSelectedId(system.id)}
                onDeselect={() => setSelectedId(null)}
                showStats={true}
              />
              <div className="system-card-actions">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={() => onSelectSystem?.({ id: system.id, name: system.name })}
                >
                  Edit Statements
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => handleDelete(system)}
                  disabled={deleteMutation.isPending}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default SystemsView;
