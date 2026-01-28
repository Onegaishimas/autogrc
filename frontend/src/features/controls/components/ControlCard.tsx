import type { PolicyStatement } from '../types';

interface ControlCardProps {
  control: PolicyStatement;
  onClick?: () => void;
}

// =============================================================================
// DEMO MODE: State Mappings
// =============================================================================
// Currently using incident table which has numeric states (1,2,3,6,7).
// IRM policy statements use string states like "draft", "active", "retired".
//
// TO SWITCH TO IRM:
// 1. Update getStateClass to only check string states (remove numeric checks)
// 2. Update getStateLabel to use IRM state strings instead of incident numbers
// 3. Verify actual IRM state values in your ServiceNow instance
//
// See: 0xcc/docs/INCIDENT_TO_IRM_MIGRATION.md for complete migration guide
// =============================================================================

export function ControlCard({ control, onClick }: ControlCardProps) {
  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    try {
      return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return dateString;
    }
  };

  // Maps state values to CSS classes for styling
  // DEMO: Includes numeric incident states (1,2,3,7)
  // IRM: Would only need string states (active, draft, retired, archived)
  const getStateClass = (state: string) => {
    const stateLower = state?.toLowerCase() || '';
    // IRM states (string-based)
    if (stateLower.includes('active')) return 'state-active';
    if (stateLower.includes('draft') || stateLower.includes('review')) return 'state-draft';
    if (stateLower.includes('inactive') || stateLower.includes('retired') || stateLower.includes('archived')) return 'state-inactive';
    // DEMO: Incident states (numeric) - Remove these for IRM
    if (stateLower === '1' || stateLower === 'new') return 'state-active';
    if (stateLower === '2') return 'state-draft';        // In Progress
    if (stateLower === '3') return 'state-draft';        // On Hold
    if (stateLower === '7' || stateLower.includes('closed')) return 'state-inactive';
    return 'state-default';
  };

  // Maps state values to human-readable labels
  // DEMO: Incident uses numeric states (1=New, 2=In Progress, etc.)
  // IRM: Uses string states (draft, active, retired, etc.)
  const getStateLabel = (state: string) => {
    // DEMO: Incident numeric state mapping - Remove this block for IRM
    const stateNum = parseInt(state, 10);
    if (!isNaN(stateNum)) {
      const incidentStateMap: Record<number, string> = {
        1: 'New',
        2: 'In Progress',
        3: 'On Hold',
        6: 'Resolved',
        7: 'Closed',
      };
      return incidentStateMap[stateNum] || `State ${state}`;
    }
    // IRM: String state - capitalize first letter
    // For IRM, this is all you need (remove incident mapping above)
    if (state) {
      return state.charAt(0).toUpperCase() + state.slice(1).toLowerCase();
    }
    return 'Unknown';
  };

  return (
    <div className="control-card" onClick={onClick} {...(onClick && { role: 'button' })}>
      <div className="control-card-header">
        <span className="control-number">{control.number}</span>
        <span className={`control-state ${getStateClass(control.state)}`}>
          {getStateLabel(control.state)}
        </span>
      </div>
      <h3 className="control-name">{control.name}</h3>
      {control.short_description && (
        <p className="control-description">{control.short_description}</p>
      )}
      <div className="control-meta">
        {control.control_family && (
          <span className="control-family">{control.control_family}</span>
        )}
        <span className="control-date">Updated: {formatDate(control.updated_at)}</span>
      </div>
    </div>
  );
}
