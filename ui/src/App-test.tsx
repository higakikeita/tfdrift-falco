/**
 * Test Component - Step 1: Drift Table Only
 */

import { useState } from 'react';
import DriftHistoryTable from './components/DriftHistoryTable';
import DriftDetailPanel from './components/DriftDetailPanel';
import { generateSampleDrifts } from './utils/sampleDrifts';
import type { DriftEvent } from './types/drift';

export default function AppTest() {
  const [selectedDrift, setSelectedDrift] = useState<DriftEvent | null>(null);
  const [drifts] = useState<DriftEvent[]>(() => generateSampleDrifts(100));

  return (
    <div className="flex h-screen bg-gray-100">
      <div className="flex-1 p-4">
        <DriftHistoryTable
          drifts={drifts}
          onSelectDrift={setSelectedDrift}
        />
      </div>
      {selectedDrift && (
        <div className="w-96">
          <DriftDetailPanel
            drift={selectedDrift}
            onClose={() => setSelectedDrift(null)}
          />
        </div>
      )}
    </div>
  );
}
