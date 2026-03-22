import { describe, it, expect } from 'vitest';

describe('graphConverter', () => {
  it('should be importable', async () => {
    const module = await import('./graphConverter');
    expect(module).toBeDefined();
  });
});
