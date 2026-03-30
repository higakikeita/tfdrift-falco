import type { Meta, StoryObj } from '@storybook/react';
import PatternSearchPanel from './PatternSearchPanel';

const meta = {
  title: 'Components/PatternSearchPanel',
  component: PatternSearchPanel,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <div style={{
        width: '100%',
        minHeight: '100vh',
        backgroundColor: '#0f172a',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}>
        <Story />
      </div>
    ),
  ],
  parameters: {
    layout: 'fullscreen',
  },
} satisfies Meta<typeof PatternSearchPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    onClose: () => console.log('Panel closed'),
    onNodeSelect: (nodeId) => console.log('Node selected:', nodeId),
  },
};

export const WithPlaceholder: Story = {
  args: {
    onClose: () => console.log('Panel closed'),
  },
};

export const WithNodeSelectCallback: Story = {
  args: {
    onClose: () => console.log('Panel closed'),
    onNodeSelect: (nodeId) => {
      console.log('Selected node:', nodeId);
      alert(`You selected node: ${nodeId}`);
    },
  },
};

export const DefaultSearchState: Story = {
  render: (args) => (
    <div className="w-full max-w-4xl">
      <PatternSearchPanel {...args} />
    </div>
  ),
  args: {
    onClose: () => console.log('Closed'),
    onNodeSelect: (nodeId) => console.log('Selected:', nodeId),
  },
};

export const SearchFormFocused: Story = {
  render: () => (
    <div className="w-full max-w-4xl">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl">
        <div className="px-6 py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-lg">🔍</span>
            <h3 className="font-semibold text-lg">パターンマッチング検索</h3>
          </div>
          <button className="p-1 hover:bg-white/20 rounded transition-colors">
            X
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                開始ノードラベル (カンマ区切り)
              </label>
              <input
                autoFocus
                type="text"
                placeholder="例: EC2, Compute"
                className="w-full px-3 py-2 border-2 border-purple-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                関係タイプ (空欄で全て)
              </label>
              <input
                type="text"
                placeholder="例: DEPENDS_ON, PART_OF"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                終了ノードラベル (カンマ区切り)
              </label>
              <input
                type="text"
                placeholder="例: Subnet, Network"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                終了ノードフィルタ (JSON形式)
              </label>
              <textarea
                placeholder='例: {"id": "subnet-123"}'
                rows={2}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 font-mono text-sm"
              />
            </div>

            <div className="flex gap-2">
              <button className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded font-medium transition-colors flex items-center justify-center gap-2">
                <span>🔍</span>
                検索
              </button>
              <button className="px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 rounded font-medium transition-colors">
                クリア
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  ),
  args: {
    onClose: () => {},
  },
};

export const NoResultsState: Story = {
  render: () => (
    <div className="w-full max-w-4xl">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl">
        <div className="px-6 py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-lg">🔍</span>
            <h3 className="font-semibold text-lg">パターンマッチング検索</h3>
          </div>
          <button className="p-1 hover:bg-white/20 rounded transition-colors">
            X
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                開始ノードラベル
              </label>
              <input
                type="text"
                value="NonExistent"
                readOnly
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>
            <div className="flex gap-2">
              <button className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded font-medium">
                検索
              </button>
            </div>
          </div>

          <div className="text-center py-8 text-gray-500 dark:text-gray-400">
            <p className="text-lg font-medium">マッチする結果が見つかりませんでした</p>
          </div>
        </div>
      </div>
    </div>
  ),
  args: {
    onClose: () => {},
  },
};

export const WithResultsState: Story = {
  render: () => (
    <div className="w-full max-w-4xl">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl">
        <div className="px-6 py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-lg">🔍</span>
            <h3 className="font-semibold text-lg">パターンマッチング検索</h3>
          </div>
          <button className="p-1 hover:bg-white/20 rounded transition-colors">
            X
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6 max-h-96">
          <div className="space-y-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                開始ノードラベル
              </label>
              <input
                type="text"
                value="EC2"
                readOnly
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>
            <div className="flex gap-2">
              <button className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded font-medium">
                検索
              </button>
            </div>
          </div>

          <div className="space-y-3">
            <h4 className="font-semibold text-gray-700 dark:text-gray-300">
              検索結果: 3 件
            </h4>

            {[
              { path: 'ec2-1 → rds-primary → backup-s3' },
              { path: 'ec2-2 → cache-redis → monitoring' },
              { path: 'ec2-3 → lb-alb → cloudfront' },
            ].map((result, idx) => (
              <div
                key={idx}
                className="p-4 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600"
              >
                <div className="flex items-center gap-3">
                  {result.path.split(' → ').map((node, nodeIdx, arr) => (
                    <div key={node} className="flex items-center gap-2">
                      <button className="flex-1 text-left px-3 py-2 bg-white dark:bg-gray-800 rounded border border-gray-300 dark:border-gray-600 hover:border-purple-500 dark:hover:border-purple-400 transition-colors text-sm">
                        <div className="font-medium text-gray-900 dark:text-gray-100">
                          {node}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400 font-mono">
                          resource
                        </div>
                      </button>
                      {nodeIdx < arr.length - 1 && (
                        <div className="text-gray-400 dark:text-gray-500">→</div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  ),
  args: {
    onClose: () => {},
    onNodeSelect: (nodeId) => console.log('Selected:', nodeId),
  },
};

export const WithLoadingState: Story = {
  render: () => (
    <div className="w-full max-w-4xl">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl">
        <div className="px-6 py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-lg">🔍</span>
            <h3 className="font-semibold text-lg">パターンマッチング検索</h3>
          </div>
          <button className="p-1 hover:bg-white/20 rounded transition-colors">
            X
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                開始ノードラベル
              </label>
              <input
                type="text"
                value="EC2"
                readOnly
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>
            <div className="flex gap-2">
              <button disabled className="flex-1 px-4 py-2 bg-purple-400 text-white rounded font-medium transition-colors flex items-center justify-center gap-2">
                <span className="inline-block animate-spin">⟳</span>
                検索中...
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  ),
  args: {
    onClose: () => {},
  },
};
