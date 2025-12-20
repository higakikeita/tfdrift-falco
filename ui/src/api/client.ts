// API Client using native fetch
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: number;
    message: string;
  };
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

class APIClient {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const apiResponse: APIResponse<T> = await response.json();

      if (!apiResponse.success) {
        throw new Error(
          apiResponse.error?.message || 'API request failed'
        );
      }

      return apiResponse.data as T;
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Graph API
  async getGraph() {
    return this.request('/graph');
  }

  async getNodes(params?: { page?: number; limit?: number }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.limit) searchParams.append('limit', params.limit.toString());

    const query = searchParams.toString();
    return this.request(`/graph/nodes${query ? `?${query}` : ''}`);
  }

  async getEdges(params?: { page?: number; limit?: number }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.limit) searchParams.append('limit', params.limit.toString());

    const query = searchParams.toString();
    return this.request(`/graph/edges${query ? `?${query}` : ''}`);
  }

  // State API
  async getState() {
    return this.request('/state');
  }

  async getStateResources(params?: { page?: number; limit?: number }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.limit) searchParams.append('limit', params.limit.toString());

    const query = searchParams.toString();
    return this.request(`/state/resources${query ? `?${query}` : ''}`);
  }

  // Events API
  async getEvents(params?: {
    page?: number;
    limit?: number;
    severity?: string;
    provider?: string;
  }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    if (params?.severity) searchParams.append('severity', params.severity);
    if (params?.provider) searchParams.append('provider', params.provider);

    const query = searchParams.toString();
    return this.request(`/events${query ? `?${query}` : ''}`);
  }

  async getEvent(id: string) {
    return this.request(`/events/${id}`);
  }

  // Drifts API
  async getDrifts(params?: {
    page?: number;
    limit?: number;
    severity?: string;
    resource_type?: string;
  }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.limit) searchParams.append('limit', params.limit.toString());
    if (params?.severity) searchParams.append('severity', params.severity);
    if (params?.resource_type)
      searchParams.append('resource_type', params.resource_type);

    const query = searchParams.toString();
    return this.request(`/drifts${query ? `?${query}` : ''}`);
  }

  async getDrift(id: string) {
    return this.request(`/drifts/${id}`);
  }

  // Stats API
  async getStats() {
    return this.request('/stats');
  }
}

// Export singleton instance
export const apiClient = new APIClient(API_BASE_URL);
