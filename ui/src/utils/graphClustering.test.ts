import { describe, it, expect } from 'vitest';

// Test the graphClustering utility
describe('graphClustering', () => {
  it('should be importable', async () => {
    const module = await import('./graphClustering');
    expect(module).toBeDefined();
  });
});
