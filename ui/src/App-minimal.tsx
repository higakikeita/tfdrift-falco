/* eslint-disable @typescript-eslint/no-unused-vars */
import CytoscapeGraph from './components/CytoscapeGraph';
import { generateSampleCausalChain } from './utils/sampleData';

function AppMinimal() {
  const graphData = generateSampleCausalChain();

  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Simple Header */}
      <div style={{
        padding: '20px',
        background: '#1a202c',
        color: 'white',
        borderBottom: '1px solid #2d3748'
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>TFDrift-Falco Graph</h1>
      </div>

      {/* Graph */}
      <div style={{ flex: 1, background: '#f7fafc' }}>
        <CytoscapeGraph
          elements={graphData}
          layout="dagre"
          onNodeClick={(id, _data) => console.log('Clicked:', id)}
          highlightedPath={[]}
        />
      </div>
    </div>
  );
}

export default AppMinimal;
