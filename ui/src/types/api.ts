/**
 * Comprehensive API Response Types
 * Defines all API response structures used throughout the application
 */

// GraphDB Node Structure
export interface Node {
  id: string;
  labels: string[];
  properties: Record<string, unknown>;
}

// GraphDB Neighbors Response
export interface NodeNeighborsResponse {
  data: {
    neighbors: Node[];
  };
}

// GraphDB Dependencies Response
export interface DependenciesResponse {
  data: {
    dependencies: Node[];
  };
}

// GraphDB Dependents Response
export interface DependentsResponse {
  data: {
    dependents: Node[];
  };
}

// GraphDB Node Response
export interface NodeResponse {
  data: {
    node: Node;
  };
}

// GraphDB Impact Radius Response
export interface ImpactRadiusResponse {
  data: {
    nodes: Node[];
  };
}

// Pattern Matching Response
export interface PatternMatchResponse {
  data: {
    matches: Array<Node[]>;
  };
}

// Webhook test response
export interface WebhookTestResponse {
  success: boolean;
  message: string;
}

// Generic API response wrapper
export interface ApiResponse<T> {
  data: T;
  status: number;
  message?: string;
}
