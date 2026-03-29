import '@testing-library/jest-dom/vitest';

// localStorage / sessionStorage mock
// zustand persist と WelcomeModal テスト用
// jsdom では defineProperty で上書き不可のため vi.stubGlobal を使用
beforeAll(() => {
  const createStorageMock = () => {
    let store: Record<string, string> = {};
    return {
      getItem: (key: string): string | null => store[key] ?? null,
      setItem: (key: string, value: string): void => { store[key] = String(value); },
      removeItem: (key: string): void => { delete store[key]; },
      clear: (): void => { store = {}; },
      get length(): number { return Object.keys(store).length; },
      key: (index: number): string | null => Object.keys(store)[index] ?? null,
    };
  };
  vi.stubGlobal('localStorage', createStorageMock());
  vi.stubGlobal('sessionStorage', createStorageMock());
});
