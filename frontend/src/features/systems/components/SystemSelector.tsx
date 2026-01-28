import { useState, useCallback, useMemo } from 'react';
import { SystemList } from './SystemList';
import type { DiscoveredSystem } from '../types';

const MAX_SELECTION = 10;

interface SystemSelectorProps {
  systems: DiscoveredSystem[];
  onSelectionChange: (selectedIds: string[]) => void;
  isLoading?: boolean;
  hideImported?: boolean;
}

export function SystemSelector({
  systems,
  onSelectionChange,
  isLoading = false,
  hideImported = false,
}: SystemSelectorProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // Filter systems based on hideImported setting
  const filteredSystems = useMemo(() => {
    if (hideImported) {
      return systems.filter((s) => !s.is_imported);
    }
    return systems;
  }, [systems, hideImported]);

  // Systems that are already imported are disabled for selection
  const disabledIds = useMemo(() => {
    return new Set(systems.filter((s) => s.is_imported).map((s) => s.sn_sys_id));
  }, [systems]);

  // Available systems that can be selected
  const availableCount = useMemo(() => {
    return systems.filter((s) => !s.is_imported).length;
  }, [systems]);

  const handleSelect = useCallback(
    (sysId: string) => {
      if (selectedIds.size >= MAX_SELECTION) {
        return; // Don't allow more than max
      }

      const newSelected = new Set(selectedIds);
      newSelected.add(sysId);
      setSelectedIds(newSelected);
      onSelectionChange(Array.from(newSelected));
    },
    [selectedIds, onSelectionChange]
  );

  const handleDeselect = useCallback(
    (sysId: string) => {
      const newSelected = new Set(selectedIds);
      newSelected.delete(sysId);
      setSelectedIds(newSelected);
      onSelectionChange(Array.from(newSelected));
    },
    [selectedIds, onSelectionChange]
  );

  const handleSelectAll = useCallback(() => {
    const available = systems
      .filter((s) => !s.is_imported)
      .slice(0, MAX_SELECTION)
      .map((s) => s.sn_sys_id);

    const newSelected = new Set(available);
    setSelectedIds(newSelected);
    onSelectionChange(Array.from(newSelected));
  }, [systems, onSelectionChange]);

  const handleDeselectAll = useCallback(() => {
    setSelectedIds(new Set());
    onSelectionChange([]);
  }, [onSelectionChange]);

  const atMaxSelection = selectedIds.size >= MAX_SELECTION;

  return (
    <div className="system-selector">
      <div className="system-selector-header">
        <div className="selection-info">
          <span className="selection-count">
            {selectedIds.size} of {availableCount} selected
          </span>
          {atMaxSelection && (
            <span className="selection-limit">(max {MAX_SELECTION})</span>
          )}
        </div>

        <div className="selection-actions">
          <button
            type="button"
            className="btn-link"
            onClick={handleSelectAll}
            disabled={isLoading || availableCount === 0}
          >
            Select All (up to {MAX_SELECTION})
          </button>
          <button
            type="button"
            className="btn-link"
            onClick={handleDeselectAll}
            disabled={isLoading || selectedIds.size === 0}
          >
            Clear Selection
          </button>
        </div>
      </div>

      <SystemList
        systems={filteredSystems}
        selectedIds={selectedIds}
        onSelect={handleSelect}
        onDeselect={handleDeselect}
        disabledIds={disabledIds}
        isLoading={isLoading}
        emptyMessage={
          hideImported
            ? 'All systems have been imported'
            : 'No systems found in ServiceNow'
        }
      />

      {atMaxSelection && (
        <div className="selection-warning">
          Maximum of {MAX_SELECTION} systems can be selected at once.
        </div>
      )}
    </div>
  );
}

export default SystemSelector;
